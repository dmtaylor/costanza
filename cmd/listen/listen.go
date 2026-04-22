// Package listen Command functions for the `listen` command
package listen

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dmtaylor/costanza/config"
	"github.com/dmtaylor/costanza/internal/util"
)

// Cmd listenCmd represents the listen command
var Cmd = &cobra.Command{
	Use:     "listen",
	Short:   "Start bot listening on server",
	Long:    `Start Bot & begin processing incoming events`,
	RunE:    runListen,
	Example: "costanza listen -i \"1234,2345\" -r \"9876,8765\"",
}

type Server struct {
	app config.App
	m   metrics
}

func init() {
	Cmd.PersistentFlags().StringSliceP(
		"insomniacIds",
		"i",
		nil,
		"Overwrite insomniac ids for bedtime reminders",
	)
	viper.BindPFlag("discord.insomniac_ids", Cmd.PersistentFlags().Lookup("insomniacIds"))
	Cmd.PersistentFlags().StringSliceP(
		"insomniacRoles",
		"r",
		nil,
		"Overwrite insomniac roles for bedtime reminders",
	)
	viper.BindPFlag("discord.insomniac_roles", Cmd.PersistentFlags().Lookup("insomniacRoles"))

	Cmd.PersistentFlags().Bool(
		"healthcheck",
		false,
		"enable healthcheck endpoint",
	)
	viper.BindPFlag("metrics.healthcheck_enabled", Cmd.PersistentFlags().Lookup("healthcheck"))
	Cmd.PersistentFlags().Bool(
		"metrics",
		false,
		"enable prometheus metrics",
	)
	viper.BindPFlag("metrics.metrics_enabled", Cmd.PersistentFlags().Lookup("metrics"))

	Cmd.PersistentFlags().String(
		"appname",
		"costanza-local",
		"appname for use in logging",
	)
	viper.BindPFlag("metrics.appname", Cmd.PersistentFlags().Lookup("appname"))
	viper.BindEnv("metrics.appname", "COSTANZA_METRICS_APPNAME")

	Cmd.PersistentFlags().UintP(
		"metricsPort",
		"p",
		8585,
		"port used for serving healthcheck & metrics endpoints",
	)
	viper.BindPFlag("metrics.port", Cmd.PersistentFlags().Lookup("metricsPort"))
	viper.BindEnv("metrics.port", "COSTANZA_METRICS_PORT")

}

func newServer() (*Server, error) {
	app, err := config.LoadApp()
	if err != nil {
		return nil, fmt.Errorf("failed to load server conf: %w", err)
	}
	return &Server{
		app: *app,
	}, nil
}

func runListen(_ *cobra.Command, _ []string) error {
	server, err := newServer()
	if err != nil {
		slog.Error("failed to build state: " + err.Error())
		return err
	}
	defer server.app.ConnPool.Close()

	dg, err := discordgo.New("Bot " + config.GlobalConfig.Discord.Token)
	if err != nil {
		slog.Error("failed to config bot: " + err.Error())
		return err
	}

	// Set shard info in discord ws connection
	if config.GlobalConfig.Discord.ShardCount > 1 {
		dg.ShardID = int(config.GlobalConfig.Discord.ShardId)
		dg.ShardCount = int(config.GlobalConfig.Discord.ShardCount)
	}
	// Reconnect if the websocket disconnects, this should be default behavior
	dg.ShouldReconnectOnError = true

	dg.AddHandlerOnce(func(sess *discordgo.Session, ready *discordgo.Ready) {
		listen := false
		if config.GlobalConfig.Metrics.HealthcheckEnabled {
			http.HandleFunc("/api/v1/healthcheck", util.Healthcheck)
			listen = true
		}
		if config.GlobalConfig.Metrics.MetricsEnabled {
			http.Handle("/metrics", server.setupMetrics())
			listen = true
		}
		if listen { // only http listen if health checks or metrics are configured
			go func() {
				err := http.ListenAndServe(":"+strconv.FormatUint(config.GlobalConfig.Metrics.MetricsPort, 10), nil)
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					slog.Error("healthcheck listen error: " + err.Error())
					panic(err)
				}
			}()
		}
		slog.Info("Bot started, CTL-C to quit")
	})
	dg.AddHandler(server.slashCommandFanout)
	dg.AddHandler(server.messageCreateFanout)
	dg.AddHandler(server.guildMemberAddMetricsMiddleware(server.welcomeMessage))
	dg.AddHandler(server.messageReactionAddMetricsMiddleware(server.logReactionActivity))
	// dg.AddHandler(server.interactionCreateMetricsMiddleware(server.quoteTestCommand)) // Uncomment this to add test quote command handler
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages | discordgo.IntentsGuildMembers | discordgo.IntentsGuildMessageReactions

	err = dg.Open()
	if err != nil {
		slog.Error("failed to open bot connection: " + err.Error())
		return err
	}
	var closeErr error = nil
	defer func() {
		closeErr = dg.Close()
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	return closeErr
}

func (s *Server) slashCommandFanout(sess *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.User != nil && i.User.Bot {
		return
	}
	if i.Member != nil && i.Member.User.Bot {
		return
	}
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	var handlerFn func(context.Context, *discordgo.Session, *discordgo.InteractionCreate)
	var err error
	ctx, cancel := util.ContextFromDiscordInteractionCreate(context.Background(), i, interactionTimeout)
	defer cancel()
	switch i.ApplicationCommandData().Name {
	case helpCommandName:
		handlerFn = s.help
	case licenseCommandName:
		handlerFn = s.license
	case rollCommandName:
		fallthrough
	case shadowrunCommandName:
		fallthrough
	case worldOfDarknessCommandName:
		fallthrough
	case darkHeresyTestCommandName:
		handlerFn = s.dispatchRollCommands
	case weatherCommandName:
		handlerFn = s.weatherCommand
	case leaderboardCommandName:
		handlerFn = s.getLeaderboardStats
	case cursedAdminCommandName:
		handlerFn = s.processCursedAdminCommand

	default:
		err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Unrecognized command: " + i.ApplicationCommandData().Name,
			}})
		if err != nil {
			slog.ErrorContext(ctx, "failed to interact with command: "+err.Error())
		}
	}
	if s.m.enabled {
		s.m.eventReceives.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent}).Inc()
		defer s.m.eventsHandled.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent}).Inc()
	}
	handlerFn(ctx, sess, i)
}

func (s *Server) messageCreateFanout(sess *discordgo.Session, m *discordgo.MessageCreate) {
	if util.MessageExcluded(sess, m) {
		return
	}

	ctx := util.ContextFromDiscordMessageCreate(context.Background(), m)
	if s.m.enabled {
		s.m.eventReceives.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent}).Inc()
		defer s.m.eventsHandled.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent}).Inc()
	}
	var wg sync.WaitGroup

	wg.Go(func() {
		s.echoQuote(ctx, sess, m)
	})
	wg.Go(func() {
		s.echoInsomniac(ctx, sess, m)
	})
	wg.Go(func() {
		s.dailyGameHandler(ctx, sess, m)
	})
	wg.Go(func() {
		s.logMessageActivity(ctx, m)
	})
	wg.Go(func() {
		s.logCursedChannelStat(ctx, sess, m)
	})
	wg.Go(func() {
		s.logCursedPostStat(ctx, sess, m)
	})
	wg.Wait()
}

package cron

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/go-co-op/gocron"
	"github.com/hashicorp/go-multierror"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dmtaylor/costanza/config"
	"github.com/dmtaylor/costanza/internal/util"
)

// Cmd represents command for running cron process for scheduled tasks
var Cmd = &cobra.Command{
	Use:     "cron",
	Short:   "Run scheduled task runner",
	RunE:    runCron,
	Example: "costanza cron -p 8585",
}

// Use UTC for scheduled times. I hope I don't regret this
var tz = time.UTC

type cronConfig struct {
	app  *config.App
	sess *discordgo.Session
	m    metrics
}

func init() {
	Cmd.PersistentFlags().UintP(
		"metricsPort",
		"p",
		8585,
		"port used for serving healthcheck & metrics endpoints",
	)
	viper.BindPFlag("metrics.port", Cmd.PersistentFlags().Lookup("metricsPort"))
	viper.BindEnv("metrics.port", "COSTANZA_METRICS_PORT")
}

func runCron(_ *cobra.Command, _ []string) error {
	app, err := config.LoadApp()
	if err != nil {
		return fmt.Errorf("failed to load app state: %w", err)
	}
	sess, err := discordgo.New("Bot " + config.GlobalConfig.Discord.Token)
	if err != nil {
		return fmt.Errorf("failed to create discord session: %w", err)
	}
	err = sess.Open()
	if err != nil {
		return fmt.Errorf("failed to open session: %w", err)
	}
	defer sess.Close()
	c := &cronConfig{
		app:  app,
		sess: sess,
	}
	var metricsServerStarted sync.WaitGroup
	metricsServerStarted.Add(1)
	go func() {
		if config.GlobalConfig.Metrics.MetricsEnabled {
			http.HandleFunc("/api/v1/healthcheck", util.Healthcheck)
			http.Handle("/metrics", c.setupMetrics())
			metricsServerStarted.Done()
			err := http.ListenAndServe(":"+strconv.FormatUint(config.GlobalConfig.Metrics.MetricsPort, 10), nil)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				slog.Error("healthcheck listen error: " + err.Error())
				panic(err)
			}
		} else {
			metricsServerStarted.Done() // unblock if metrics aren't enabled you idiot
		}
	}()
	s := gocron.NewScheduler(tz)
	for _, lconfig := range config.GlobalConfig.Discord.ListenConfigs {
		runtime, err := time.Parse("15:04", lconfig.StartTime)
		if err != nil {
			return fmt.Errorf("failed to parse config time %s: %w", lconfig.StartTime, err)
		}
		_, err = s.Every(1).Month(1).At(runtime).Do(func(lconfig config.ListenConfig) {
			promLabels := prometheus.Labels{listenGuildIdLabel: lconfig.GuildId}
			start := time.Now()
			if c.m.enabled {
				defer func() {
					c.m.reportDurationSeconds.With(promLabels).Observe(time.Since(start).Seconds())
				}()
			}
			ctx := util.ContextFromListenConfig(context.Background(), lconfig.GuildId, lconfig.ReportChannelId)
			month := util.GetLastMonth(time.Now().UTC())
			var err *multierror.Error
			err = multierror.Append(err, c.reportMessageStats(ctx, lconfig, month))
			err = multierror.Append(err, c.reportDailyGameWins(ctx, lconfig, month))
			if err.Len() > 0 {
				if c.m.enabled {
					c.m.failedReports.With(promLabels).Inc()
				}
				slog.ErrorContext(ctx, "report(s) failed: "+err.Error())
			} else {
				if c.m.enabled {
					c.m.successfulReports.With(promLabels).Inc()
				}
			}
		}, lconfig)
		if err != nil {
			return fmt.Errorf("failed to schedule job: %w", err)
		}
	}

	_, err = s.Every(1).Month(2).At(time.Date(0, 0, 0, 16, 0, 0, 0, tz)).Do(func() {
		month := util.GetLastMonth(time.Now().UTC())
		err := c.removeStats(month)
		if err != nil {
			slog.Error("report log cleanup failed: " + err.Error())
		} else {
			slog.Info("cleaned up report log for " + month)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to schedule cleanup job: %w", err)
	}

	metricsServerStarted.Wait()
	s.StartAsync()
	slog.Info("cron service started, interrupt to shutdown")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	s.Stop()

	return nil
}

func (c *cronConfig) reportMessageStats(ctx context.Context, listenConfig config.ListenConfig, month string) error {
	guildId, err := strconv.ParseUint(listenConfig.GuildId, 10, 64)
	if err != nil {
		return fmt.Errorf("unable to parse guild id %s: %w", listenConfig.GuildId, err)
	}

	topUsers, err := c.app.Stats.GetLeaders(ctx, guildId, month)
	if err != nil {
		return fmt.Errorf("failed to get leaders: %w", err)
	}
	if len(topUsers) < 1 {
		return nil
	}
	builder := strings.Builder{}

	_, err = builder.WriteString("Top posters for the month are:\n")
	if err != nil {
		return fmt.Errorf("failed to build string: %w", err)
	}
	for i, userStat := range topUsers {
		user, err := c.sess.User(strconv.FormatUint(userStat.UserId, 10))
		if err != nil {
			return fmt.Errorf("unable to get user: %w", err)
		}
		line := fmt.Sprintf("#%d: %s with %d messages\n", i+1, user.Mention(), userStat.MessageCount)
		_, err = builder.WriteString(line)
		if err != nil {
			return fmt.Errorf("failed to write line: %w", err)
		}
	}
	_, err = c.sess.ChannelMessageSend(listenConfig.ReportChannelId, builder.String())
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

func (c *cronConfig) reportDailyGameWins(ctx context.Context, listenConfig config.ListenConfig, month string) error {
	guildId, err := strconv.ParseUint(listenConfig.GuildId, 10, 64)
	if err != nil {
		return fmt.Errorf("unable to parse guild id %s, %w", listenConfig.GuildId, err)
	}
	topWinners, err := c.app.Stats.GetDailyGameLeaders(ctx, guildId, month)
	if err != nil {
		return fmt.Errorf("failed to get winners: %w", err)
	}
	if len(topWinners) < 1 {
		return nil
	}
	builder := strings.Builder{}

	_, err = builder.WriteString("Top posters for the month are:\n")
	if err != nil {
		return fmt.Errorf("failed to build string: %w", err)
	}
	for i, dailyGameWins := range topWinners {
		user, err := c.sess.User(strconv.FormatUint(dailyGameWins.UserId, 10))
		if err != nil {
			return fmt.Errorf("unable to get user: %w", err)
		}
		line := fmt.Sprintf("#%d: %s with %s\n", i+1, user.Mention(), dailyGameWins.FormatWins())
		builder.WriteString(line)
	}

	_, err = c.sess.ChannelMessageSend(listenConfig.ReportChannelId, builder.String())
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

func (c *cronConfig) removeStats(month string) error {
	var err *multierror.Error
	err = multierror.Append(err, c.app.Stats.RemoveMonthActivity(context.Background(), month))
	err = multierror.Append(err, c.app.Stats.RemoveDailyGameLeadersForMonth(context.Background(), month))
	return err.ErrorOrNil()
}

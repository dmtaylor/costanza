// Package listen Command functions for the `listen` command
package listen

import (
	"log"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dmtaylor/costanza/config"
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
	app              config.App
	dailyWinPatterns []*regexp.Regexp
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
}

func newServer() (*Server, error) {
	app, err := config.LoadApp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to load server conf")
	}
	framedPattern, err := regexp.Compile(`Framed\s+#\d+\s+🎥 🟩 ⬛ ⬛ ⬛ ⬛ ⬛`)
	if err != nil {
		return nil, errors.Wrap(err, "failed to compile framed pattern")
	}

	tradlePattern, err := regexp.Compile(`#Tradle\s.*#\d+\s1/6`)
	if err != nil {
		return nil, errors.Wrap(err, "failed to compile tradle pattern")
	}

	wordlePattern, err := regexp.Compile(`#Wordle\s#?\d+\s+1/6`)
	if err != nil {
		return nil, errors.Wrap(err, "failed to compile wordle pattern")
	}

	heardlePattern, err := regexp.Compile(`#Heardle\s#\d+\s+🔊🟩⬜⬜⬜⬜⬜\n`)
	if err != nil {
		return nil, errors.Wrap(err, "failed to compile heardle pattern")
	}

	return &Server{
		app:              *app,
		dailyWinPatterns: []*regexp.Regexp{framedPattern, tradlePattern, wordlePattern, heardlePattern},
	}, nil
}

func runListen(cmd *cobra.Command, args []string) error {
	server, err := newServer()
	if err != nil {
		log.Printf("failed to build state")
		return err
	}
	defer server.app.ConnPool.Close()

	dg, err := discordgo.New("Bot " + config.GlobalConfig.Discord.Token)
	if err != nil {
		log.Printf("failed to start bot: %s\n", err)
		return err
	}
	dg.AddHandler(server.Help)
	dg.AddHandler(server.EchoQuote)
	dg.AddHandler(server.EchoInsomniac)
	dg.AddHandler(server.DispatchRollCommands)
	dg.AddHandler(server.DailyWinReact)
	dg.AddHandler(server.LogMessageActivity)
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages

	err = dg.Open()
	if err != nil {
		log.Printf("failed to open bot connection: %s\n", err)
		return err
	}
	var closeErr error = nil
	defer func() {
		closeErr = dg.Close()
	}()
	log.Printf("Bot started, CTL-C to quit")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	return closeErr
}

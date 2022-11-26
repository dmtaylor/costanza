// Package listen Command functions for the `listen` command
package listen

import (
	"context"
	"log"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/dmtaylor/costanza/config"
	"github.com/dmtaylor/costanza/internal/parser"
	"github.com/dmtaylor/costanza/internal/quotes"
	"github.com/dmtaylor/costanza/internal/roller"
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
	config           *config.Config
	quotes           *quotes.QuoteEngine
	dNotationParser  *parser.DNotationParser
	thresholdRoller  *roller.ThresholdRoller
	connPool         *pgxpool.Pool
	dailyWinPatterns []*regexp.Regexp
}

func init() {
	Cmd.PersistentFlags().StringSliceVarP(
		&config.OverwriteInsomniacIds,
		"insomniacIds",
		"i",
		nil,
		"Overwrite insomniac ids for bedtime reminders",
	)
	Cmd.PersistentFlags().StringSliceVarP(
		&config.OverwriteInsomniacRoles,
		"insomniacRoles",
		"r",
		nil,
		"Overwrite insomniac roles for bedtime reminders",
	)
}

func newServer() (*Server, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, errors.Wrap(err, "failed to load cfgs while building server")
	}
	pool, err := pgxpool.Connect(context.Background(), cfg.DbConnectionStr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build connection pool")
	}
	qEngine, err := quotes.NewQuoteEngine(pool)
	if err != nil {
		return nil, errors.Wrap(err, "server failed to build quote engine")
	}
	dNotationParser, err := parser.NewDNotationParser()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build basic parser")
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
		config:           cfg,
		quotes:           qEngine,
		dNotationParser:  dNotationParser,
		thresholdRoller:  roller.NewThresholdRoller(),
		connPool:         pool,
		dailyWinPatterns: []*regexp.Regexp{framedPattern, tradlePattern, wordlePattern, heardlePattern},
	}, nil
}

func runListen(cmd *cobra.Command, args []string) error {
	server, err := newServer()
	if err != nil {
		log.Printf("failed to build state")
		return err
	}
	defer server.connPool.Close()

	dg, err := discordgo.New("Bot " + server.config.DiscordToken)
	if err != nil {
		log.Printf("failed to start bot: %s\n", err)
		return err
	}
	dg.AddHandler(server.Help)
	dg.AddHandler(server.EchoQuote)
	dg.AddHandler(server.EchoInsomniac)
	dg.AddHandler(server.DispatchRollCommands)
	dg.AddHandler(server.DailyWinReact)
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

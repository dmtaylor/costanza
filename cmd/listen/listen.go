/*
Copyright © 2021 David Taylor <dmtaylor2011@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package listen

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/dmtaylor/costanza/config"
	"github.com/dmtaylor/costanza/internal/parser"
	"github.com/dmtaylor/costanza/internal/quotes"
	"github.com/dmtaylor/costanza/internal/roller"
	"github.com/dmtaylor/costanza/internal/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// listenCmd represents the listen command
var Cmd = &cobra.Command{
	Use:   "listen",
	Short: "Start bot listening on server",
	Long:  `Start Bot & begin processing incoming events`,
	RunE:  runListen,
}

type Server struct {
	config          *config.Config
	quotes          *quotes.QuoteEngine
	dNotationParser *parser.DNotationParser
	thresholdRoller *roller.ThresholdRoller
	connPool        *util.SqliteConnectionPool
	//roller Roller TODO
}

func init() {
	Cmd.PersistentFlags().StringSliceVarP(
		&config.OverwriteInsomniacIds,
		"insomniacIds",
		"i",
		nil,
		"Overwrite insomniac ids for bedtime reminders",
	)
}

func newServer() (*Server, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, errors.Wrap(err, "failed to load cfgs while building server")
	}
	pool, err := util.NewSqliteConnectionPool(cfg.DbConnectionStr, 5)
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
	return &Server{
		config:          cfg,
		quotes:          qEngine,
		dNotationParser: dNotationParser,
		thresholdRoller: roller.NewThresholdRoller(),
		connPool:        pool,
	}, nil
}

func runListen(cmd *cobra.Command, args []string) error {
	server, err := newServer()
	if err != nil {
		log.Printf("failed to build state")
		return err
	}

	dg, err := discordgo.New("Bot " + server.config.DiscordToken)
	if err != nil {
		log.Printf("failed to start bot: %s\n", err)
		return err
	}
	dg.AddHandler(server.EchoQuote)
	dg.AddHandler(server.EchoInsomniac)
	dg.AddHandler(server.DispatchRollCommands)
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	dg.Open()
	defer dg.Close()
	log.Printf("Bot started, CTL-C to quit")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	return nil
}

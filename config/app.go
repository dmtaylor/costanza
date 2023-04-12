package config

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"

	"github.com/dmtaylor/costanza/internal/parser"
	"github.com/dmtaylor/costanza/internal/quotes"
	"github.com/dmtaylor/costanza/internal/roller"
	"github.com/dmtaylor/costanza/internal/stats"
)

type App struct {
	Quotes          *quotes.QuoteEngine
	DNotationParser *parser.DNotationParser
	ThresholdRoller *roller.ThresholdRoller
	ConnPool        *pgxpool.Pool
	Stats           *stats.Stats
}

var loader sync.Once
var app App

func LoadApp() (*App, error) {
	var err error
	loader.Do(func() {
		err = LoadConfig()
		if err != nil {
			err = errors.Wrap(err, "failed to load cfgs while loading app")
			return
		}
		pool, err := pgxpool.New(context.Background(), GlobalConfig.Db.Connection)
		if err != nil {
			err = errors.Wrap(err, "failed to build connection pool")
			return
		}
		qEngine, err := quotes.NewQuoteEngine(pool)
		if err != nil {
			err = errors.Wrap(err, "server failed to build quote engine")
			return
		}
		statsSvc := stats.New(pool)
		dNotationParser, err := parser.NewDNotationParser()
		if err != nil {
			err = errors.Wrap(err, "failed to build parser")
			return
		}
		app = App{
			Quotes:          qEngine,
			DNotationParser: dNotationParser,
			ThresholdRoller: roller.NewThresholdRoller(),
			ConnPool:        pool,
			Stats:           &statsSvc,
		}
	})
	if err != nil {
		return nil, err
	}
	return &app, nil
}

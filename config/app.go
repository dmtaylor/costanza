package config

import (
	"context"
	"regexp"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"

	"github.com/dmtaylor/costanza/internal/parser"
	"github.com/dmtaylor/costanza/internal/quotes"
	"github.com/dmtaylor/costanza/internal/roller"
)

type App struct {
	Quotes           *quotes.QuoteEngine
	DNotationParser  *parser.DNotationParser
	ThresholdRoller  *roller.ThresholdRoller
	ConnPool         *pgxpool.Pool
	DailyWinPatterns []*regexp.Regexp
}

func LoadApp() (*App, error) {
	err := LoadConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to load cfgs while loading app")
	}

	pool, err := pgxpool.New(context.Background(), GlobalConfig.DbConnectionStr)
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

	return &App{
		Quotes:          qEngine,
		DNotationParser: dNotationParser,
		ThresholdRoller: roller.NewThresholdRoller(),
		ConnPool:        pool,
	}, nil
}

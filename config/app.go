package config

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dmtaylor/costanza/internal/model"
	"github.com/dmtaylor/costanza/internal/parser"
	"github.com/dmtaylor/costanza/internal/quotes"
	"github.com/dmtaylor/costanza/internal/roller"
	"github.com/dmtaylor/costanza/internal/stats"
)

// VersionString current tagged version for application
const VersionString = "v1.5.0"

// App represents the current app components & state
type App struct {
	Quotes          quotes.QuoteEngine
	DNotationParser *parser.DNotationParser
	ThresholdRoller *roller.ThresholdRoller
	ConnPool        model.DbPool
	Stats           *stats.Stats
}

var loader sync.Once
var app App // use singleton App

// LoadApp loads default app state
func LoadApp() (*App, error) {
	var err error
	loader.Do(func() {
		err = LoadConfig()
		if err != nil {
			err = fmt.Errorf("failed to load cfgs while loading app: %w", err)
			return
		}
		pool, err := pgxpool.New(context.Background(), GlobalConfig.Db.Connection)
		if err != nil {
			err = fmt.Errorf("failed to build connection pool: %w", err)
			return
		}
		// test db connection before starting app
		err = pool.Ping(context.Background())
		if err != nil {
			err = fmt.Errorf("failed to get db connection: %w", err)
			return
		}
		qEngine, err := quotes.NewQuoteEngine(pool, uint64(time.Now().UnixNano()))
		if err != nil {
			err = fmt.Errorf("server failed to build quote engine: %w", err)
			return
		}
		statsSvc := stats.New(pool)
		dNotationParser, err := parser.NewDNotationParser()
		if err != nil {
			err = fmt.Errorf("failed to build parser: %w", err)
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

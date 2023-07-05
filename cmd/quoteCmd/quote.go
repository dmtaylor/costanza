package quoteCmd

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"

	"github.com/dmtaylor/costanza/config"
	"github.com/dmtaylor/costanza/internal/quotes"
)

// Cmd represents the quote command
var Cmd = &cobra.Command{
	Use:   "quote",
	Short: "Test getting a quote",
	Long:  `Utility test command for pulling quotes from the quote source`,
	RunE:  runQuote,
}

var n uint

func init() {
	Cmd.PersistentFlags().UintVarP(
		&n,
		"times",
		"n",
		1,
		"Number of quotes to get",
	)
}

func runQuote(_ *cobra.Command, _ []string) error {
	err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	pool, err := pgxpool.New(context.Background(), config.GlobalConfig.Db.Connection)
	if err != nil {
		return fmt.Errorf("failed to build conn pool: %w", err)
	}
	engine, err := quotes.NewQuoteEngine(pool)
	if err != nil {
		return fmt.Errorf("failed to build engine: %w", err)
	}
	for i := uint(0); i < n; i++ {
		quote, err := engine.GetQuoteSql(context.Background())
		if err != nil {
			return fmt.Errorf("failed to get quote: %w", err)
		}
		fmt.Printf("%d: %+v\n", i+1, quote)
	}
	return nil
}

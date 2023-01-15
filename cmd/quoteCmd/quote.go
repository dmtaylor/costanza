package quoteCmd

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/dmtaylor/costanza/config"
	"github.com/dmtaylor/costanza/internal/quotes"
)

// quoteCmd represents the quote command
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

func runQuote(cmd *cobra.Command, args []string) error {
	err := config.LoadConfig()
	if err != nil {
		return errors.Wrap(err, "failed to load config")
	}
	pool, err := pgxpool.New(context.Background(), config.GlobalConfig.DbConnectionStr)
	if err != nil {
		return errors.Wrap(err, "failed to build conn pool")
	}
	engine, err := quotes.NewQuoteEngine(pool)
	if err != nil {
		return errors.Wrap(err, "failed to build engine")
	}
	for i := uint(0); i < n; i++ {
		quote, err := engine.GetQuoteSql(context.Background())
		if err != nil {
			return errors.Wrap(err, "failed to get quote")
		}
		fmt.Printf("%d: %s\n", i+1, quote)
	}
	return nil
}

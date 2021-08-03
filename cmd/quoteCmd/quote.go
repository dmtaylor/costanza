/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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
package quoteCmd

import (
	"fmt"

	"github.com/dmtaylor/costanza/config"
	"github.com/dmtaylor/costanza/internal/quotes"
	"github.com/dmtaylor/costanza/internal/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
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
	cfg, err := config.Load()
	if err != nil {
		return errors.Wrap(err, "failed to load config")
	}
	pool, err := util.NewSqliteConnectionPool(cfg.DbConnectionStr, 5)
	if err != nil {
		return errors.Wrap(err, "failed to build conn pool")
	}
	engine, err := quotes.NewQuoteEngine(pool)
	if err != nil {
		return errors.Wrap(err, "failed to build engine")
	}
	for i := uint(0); i < n; i++ {
		quote, err := engine.GetQuoteSql()
		if err != nil {
			return errors.Wrap(err, "failed to get quote")
		}
		fmt.Printf("%d: %s\n", i+1, quote)
	}
	return nil
}

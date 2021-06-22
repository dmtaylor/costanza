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
package roll

import (
	"fmt"

	"github.com/alecthomas/participle/v2"
	"github.com/dmtaylor/costanza/internal/parser"
	"github.com/spf13/cobra"
)

var printEBNF bool

var basicParser = participle.MustBuild(&parser.Expression{})

// rollCmd represents the roll command
var Cmd = &cobra.Command{
	Use:   "roll",
	Short: "Parse & do roll",
	Long: `Testing command to perform basic roll via the cli.
	
	This should approximate the 'roll' slash command`,
	RunE: runRoll,
}

func init() {
	Cmd.PersistentFlags().BoolVarP(
		&printEBNF,
		"print-ebnf",
		"p",
		false,
		"Print EBNF for basic roll rather than parse expr",
	)

}

func runRoll(cmd *cobra.Command, args []string) error {
	// TODO implement roller
	parser := parser.NewBasicParser()
	if printEBNF {
		fmt.Printf("EBNF:\n")
		fmt.Printf("%s\n", parser.GetEBNF())
	} else {
		_ = 1 // NOP to silence warning
		// NOT Implemented
	}
	return nil
}

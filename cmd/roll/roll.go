package roll

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dmtaylor/costanza/internal/parser"
)

var printEBNF bool

// Cmd rollCmd represents the roll command
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

func runRoll(_ *cobra.Command, args []string) error {
	input := strings.Join(args, " ")
	rollParser, err := parser.NewDNotationParser()
	if err != nil {
		return fmt.Errorf("failed to create rollParser: %w", err)
	}
	if printEBNF {
		fmt.Printf("EBNF:\n")
		fmt.Printf("%s\n", rollParser.GetEBNF())
	} else {
		results, err := rollParser.DoParse(input)
		if err != nil {
			return fmt.Errorf("failed to do parse: %w", err)
		}
		fmt.Printf("%s = %s\n", results.StrValue, strconv.Itoa(results.Value))
	}
	return nil
}

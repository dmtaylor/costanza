package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dmtaylor/costanza/config"
)

var dumpVars bool

// cfgCmd represents the cfg command
var cfgCmd = &cobra.Command{
	Use:   "cfg",
	Short: "Load & validate config",
	Long: `Loads & validates the config object for testing

    This loads & echos the loaded configuration to debug config loading.
    This will potentially echo secrets to stdout, do not call in sensitive envs.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Got error when loading: %v\n", err)
		}
		if dumpVars {
			fmt.Printf("Config: %+v\n", cfg)
		} else {
			fmt.Printf("Config Loaded\n")
		}
	},
}

func init() {
	cfgCmd.Flags().BoolVarP(&dumpVars, "dumpvars", "d", false, "Echo loaded config")
}

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/dmtaylor/costanza/cmd/listen"
	"github.com/dmtaylor/costanza/cmd/quoteCmd"
	"github.com/dmtaylor/costanza/cmd/roll"
	"github.com/dmtaylor/costanza/config"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "costanza",
	Short: "A discord bot for architects",
	Long: `A discord bot which does a few things.
	
	It responds to mentions with George Costanza quotes a la Gandalf. It also
	implements several slash commands for evaluating & performing dice rolls based
	on "d notation".`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&config.OverwriteDiscordToken, "token", "t", "", "Overwrite bot token")
	rootCmd.PersistentFlags().StringVarP(
		&config.OverwriteDbConnectionStr,
		"connectionStr",
		"c",
		"",
		"Overwrite postgres connection string from env",
	)
	rootCmd.AddCommand(listen.Cmd, roll.Cmd, quoteCmd.Cmd, cfgCmd)
}

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dmtaylor/costanza/cmd/cron"
	"github.com/dmtaylor/costanza/cmd/listen"
	"github.com/dmtaylor/costanza/cmd/quoteCmd"
	"github.com/dmtaylor/costanza/cmd/register"
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
	cobra.OnInitialize(config.SetConfigDefaults)
	rootCmd.PersistentFlags().StringP("token", "t", "", "Overwrite bot token")
	viper.BindPFlag(config.TokenPath, rootCmd.PersistentFlags().Lookup("token"))
	rootCmd.PersistentFlags().StringP(
		"connectionStr",
		"c",
		"",
		"Overwrite postgres connection string from env",
	)
	viper.BindPFlag("db.connection", rootCmd.PersistentFlags().Lookup("connectionStr"))
	viper.BindEnv("db.connection", "COSTANZA_DB_URL")
	rootCmd.AddCommand(listen.Cmd, roll.Cmd, quoteCmd.Cmd, cfgCmd, register.Cmd, cron.Cmd)
}

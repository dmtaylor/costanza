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
package cmd

import (
	"log"

	"github.com/dmtaylor/costanza/internal/config"
	"github.com/spf13/cobra"
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
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&config.Values.DiscordToken, "token", "i", "", "Overwrite bot token")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configs: %s", err)
	}

}

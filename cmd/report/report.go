// Package report command to output stats
package report

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"

	"github.com/dmtaylor/costanza/config"
)

type statsHandle struct {
	app  *config.App
	sess *discordgo.Session
}

// Cmd represents the report command
var Cmd = &cobra.Command{
	Use:   "report",
	Short: "Send usage stats to configured channels",
	Long:  `Send usage stats to the channels configured in the report configs`,
	RunE:  runStats,
}

func init() {
	Cmd.AddCommand(removeCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// reportCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// reportCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	Cmd.PersistentFlags().StringP("month", "m", time.Now().Format("2006-01"), "Month for querying stats")
}

func runStats(cmd *cobra.Command, args []string) error {
	app, err := config.LoadApp()
	if err != nil {
		return errors.Wrap(err, "failed to load app state")
	}
	month, err := cmd.PersistentFlags().GetString("month")
	if err != nil {
		return errors.Wrap(err, "error getting month")
	}
	log.Printf("starting getting stats for %s", month)
	sess, err := discordgo.New("Bot " + config.GlobalConfig.Discord.Token)
	handle := statsHandle{app: app, sess: sess}
	var wg sync.WaitGroup
	var subErrors error = nil
	for _, listenConfig := range config.GlobalConfig.Discord.ListenConfigs {
		wg.Add(1)
		go func(lconfig config.ListenConfig) {
			defer wg.Done()
			ctx := context.Background()
			err := handle.reportMessageStats(ctx, lconfig, month)
			if err != nil {
				subErrors = multierr.Append(subErrors, err)
			}
		}(listenConfig)
	}
	wg.Wait()
	log.Printf("finished getting stats for %s", month)
	return subErrors
}

func (s statsHandle) reportMessageStats(ctx context.Context, listenConfig config.ListenConfig, month string) error {
	guildId, err := strconv.ParseUint(listenConfig.GuildId, 10, 64)
	if err != nil {
		return errors.Wrap(err, "unable to parse guild id "+listenConfig.GuildId)
	}

	topUsers, err := s.app.Stats.GetLeaders(ctx, guildId, month)
	builder := strings.Builder{}
	_, err = builder.WriteString("Top posters for the month are:\n")
	if err != nil {
		return errors.Wrap(err, "failed to build string")
	}
	if len(topUsers) == 0 {
		// return early if there are no top posters for guild
		return nil
	}
	for i, userStat := range topUsers {
		user, err := s.sess.User(strconv.FormatUint(userStat.UserId, 10))
		if err != nil {
			return errors.Wrap(err, "unable to get user")
		}
		line := fmt.Sprintf("#%d: %s with %d messages\n", i+1, user.Mention(), userStat.MessageCount)
		_, err = builder.WriteString(line)
	}
	_, err = s.sess.ChannelMessageSend(listenConfig.ReportChannelId, builder.String())
	if err != nil {
		return errors.Wrap(err, "failed to send message")
	}

	return nil
}

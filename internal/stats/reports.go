package stats

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/dmtaylor/costanza/internal/model"
)

// BuildMessageReport creates the message for message stat report
func BuildMessageReport(stats []*model.DiscordUsageStat) string {
	builder := strings.Builder{}
	builder.WriteString("Top posters for the month are:\n")
	for i, userStat := range stats {
		user := discordgo.User{ID: strconv.FormatUint(userStat.UserId, 10)}
		line := fmt.Sprintf("#%d: %s with %d messages\n", i+1, user.Mention(), userStat.MessageCount)
		builder.WriteString(line)
	}

	return builder.String()
}

// BuildGameWinReport creates the message for daily game winner reports
func BuildGameWinReport(topWinners []*model.DailyGameWinStat) string {
	builder := strings.Builder{}
	builder.WriteString("Top game winners for the month are:\n")
	for i, dailyGameStat := range topWinners {
		user := discordgo.User{ID: strconv.FormatUint(dailyGameStat.UserId, 10)}
		line := fmt.Sprintf("#%d: %s with %s\n", i+1, user.Mention(), dailyGameStat.FormatWins())
		builder.WriteString(line)
	}
	return builder.String()
}

// BuildReactionScoreReport creates the message for reaction scores
func BuildReactionScoreReport(topReactionScores []*model.DiscordReactionScore) string {
	builder := strings.Builder{}
	builder.WriteString("Top reaction scores are:\n")
	for i, reactionScore := range topReactionScores {
		user := discordgo.User{ID: strconv.FormatUint(reactionScore.UserId, 10)}
		line := fmt.Sprintf("#%d: %s\n", i+1, reactionScore.FormatResult(user.Mention()))
		builder.WriteString(line)
	}

	return builder.String()
}

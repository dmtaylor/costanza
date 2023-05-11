package listen

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/exp/slog"
)

// dailyWinReact performs reaction if it detects a win pattern in the message
func (s *Server) dailyWinReact(sess *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == sess.State.User.ID {
		return
	}
	ctx := context.WithValue(context.Background(), "messageId", m.ID)
	ctx = context.WithValue(ctx, "guildId", m.GuildID)

	for _, pattern := range s.dailyWinPatterns {
		if pattern.MatchString(m.Message.Content) {
			err := sess.MessageReactionAdd(m.ChannelID, m.Message.ID, "💯")
			if err != nil {
				slog.ErrorCtx(ctx, fmt.Sprintf("error adding reaction: %s", err))
			}
			return
		}
	}
}

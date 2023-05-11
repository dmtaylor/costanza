package listen

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/exp/slog"
)

// echoQuote handler function for sending George Costanza quotes
func (s *Server) echoQuote(sess *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == sess.State.User.ID {
		return
	}
	ctx := context.WithValue(context.Background(), "messageId", m.ID)
	ctx = context.WithValue(ctx, "guildId", m.GuildID)

	for _, mentionedUser := range m.Mentions {
		if mentionedUser.ID == sess.State.User.ID {
			s.sendQuote(ctx, sess, m)
			return
		}
	}
}

func (s *Server) sendQuote(ctx context.Context, sess *discordgo.Session, m *discordgo.MessageCreate) {
	quote, err := s.app.Quotes.GetQuoteSql(ctx)
	if err != nil {
		slog.ErrorCtx(ctx, "failed to get quote: "+err.Error())
		_, err := sess.ChannelMessageSendReply(
			m.ChannelID,
			"I was unable to get a quote. Why must there always be a problem?",
			m.Reference(),
		)
		if err != nil {
			slog.ErrorCtx(ctx, "error sending message: ", err.Error())
		}
		return
	}
	_, err = sess.ChannelMessageSendReply(m.ChannelID, quote, m.Reference())
	if err != nil {
		slog.ErrorCtx(ctx, "error sending message: "+err.Error())
	}
}

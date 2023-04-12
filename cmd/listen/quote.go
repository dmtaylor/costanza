package listen

import (
	"context"
	"log"

	"github.com/bwmarrin/discordgo"
)

// echoQuote handler function for sending George Costanza quotes
func (s *Server) echoQuote(sess *discordgo.Session, m *discordgo.MessageCreate) {
	ctx := context.Background()
	if m.Author.ID == sess.State.User.ID {
		return
	}

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
		log.Printf("failed to get quote: %s\n", err)
		_, err := sess.ChannelMessageSendReply(
			m.ChannelID,
			"I was unable to get a quote. Why must there always be a problem?",
			m.Reference(),
		)
		if err != nil {
			log.Printf("error sending message: %s\n", err)
		}
		return
	}
	_, err = sess.ChannelMessageSendReply(m.ChannelID, quote, m.Reference())
	if err != nil {
		log.Printf("error sending message: %s\n", err)
	}
}

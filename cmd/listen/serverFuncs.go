package listen

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func (s *Server) EchoQuote(sess *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == sess.State.User.ID {
		return
	}

	for _, mentionedUser := range m.Mentions {
		if mentionedUser.ID == sess.State.User.ID {
			log.Printf("got message\n")
			_, err := sess.ChannelMessageSend(m.ChannelID, s.quotes.GetQuote())
			if err != nil {
				log.Printf("error sending message: %s\n", err)
			}
			return
		}
	}
}

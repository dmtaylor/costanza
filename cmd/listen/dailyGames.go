package listen

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

// dailyWinReact performs reaction if it detects a win pattern in the message
func (s *Server) dailyWinReact(sess *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == sess.State.User.ID {
		return
	}

	for _, pattern := range s.dailyWinPatterns {
		if pattern.MatchString(m.Message.Content) {
			err := sess.MessageReactionAdd(m.ChannelID, m.Message.ID, "💯")
			if err != nil {
				log.Printf("error adding reaction: %s\n", err)
			}
			return
		}
	}
}

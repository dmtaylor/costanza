package listen

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const helpMessage string = `
Costanza commands:
` +
	"```" + `
!chelp:   this message.
!roll:    parse text as d-notation and evaluate expression.
!srroll:  parse text as d-notation, evaluate, and use result for Shadowrun roll.
!wodroll: parse text as d-notation, evaluate, and use result for World of Darkness roll.
          Can be modified with '8again', '9again' and 'chance'. Rolls of < 1 dice are done as chance rolls.
!dhtest:  parse text as d-notation, evaluate, and use result for FF Warhammer 40k RPG roll (over-under on 1d100).
` +
	"```"

// help handler function for help messages
func (s *Server) help(sess *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == sess.State.User.ID {
		return
	}

	words := strings.Fields(m.Message.Content)
	if len(words) < 1 {
		return
	}
	if words[0] == "!chelp" {
		_, err := sess.ChannelMessageSend(
			m.ChannelID,
			helpMessage,
		)
		if err != nil {
			log.Printf("error sending message: %s\n", err)
		}
	}
}

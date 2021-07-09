package listen

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var startLateHours, endLateHours time.Time

func (s *Server) EchoQuote(sess *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == sess.State.User.ID {
		return
	}

	for _, mentionedUser := range m.Mentions {
		if mentionedUser.ID == sess.State.User.ID {
			_, err := sess.ChannelMessageSendReply(m.ChannelID, s.quotes.GetQuote(), m.Reference())
			if err != nil {
				log.Printf("error sending message: %s\n", err)
			}
			return
		}
	}
}

func (s *Server) EchoInsomniac(sess *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == sess.State.User.ID {
		return
	}

	for _, insomniac := range s.config.InsomniacIds {
		if m.Author.ID == insomniac && isAfterHours() {
			_, err := sess.ChannelMessageSendReply(
				m.ChannelID,
				fmt.Sprintf("%s All right. That's enough for today. You're tired. Get some sleep. I'll see you first thing in the morning.",
					m.Author.Mention()),
				m.Reference(),
			)
			if err != nil {
				log.Printf("error sending message: %s\n", err)
			}
			return
		}
	}

}

func (s *Server) DispatchRollCommands(sess *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == sess.State.User.ID {
		return
	}
	command := strings.Fields(m.Message.Content)

	switch command[0] {
	case "!roll":
		s.doDNotationRoll(sess, m, strings.Join(command[1:], " "))
	}
}

func (s *Server) doDNotationRoll(sess *discordgo.Session, m *discordgo.MessageCreate, rollStr string) {
	result, err := s.dNotationParser.DoParse(rollStr)
	if err != nil {
		log.Printf("error parsing string: %s\n", err)
		_, err := sess.ChannelMessageSendReply(
			m.ChannelID,
			fmt.Sprintf("I was unable to understand your roll \"%s\". Why must there always be a problem?", rollStr),
			m.Reference(),
		)
		if err != nil {
			log.Printf("error sending message: %s\n", err)
		}
		return
	}
	response := fmt.Sprintf("%s = %d", result.StrValue, result.Value)
	_, err = sess.ChannelMessageSendReply(
		m.ChannelID,
		response,
		m.Reference(),
	)
	if err != nil {
		log.Printf("error sending message: %s\n", err)
	}
}

func isAfterHours() bool {
	var err error
	if startLateHours.IsZero() {
		startLateHours, err = time.Parse(time.Kitchen, "12:30AM")
		if err != nil {
			log.Panicf("Error parsing start date format: %s\n", err)
		}
	}
	if endLateHours.IsZero() {
		endLateHours, err = time.Parse(time.Kitchen, "06:00AM")
		if err != nil {
			log.Panicf("Error parsing end date format: %s\n", err)
		}
	}

	currentTime, err := time.Parse(time.Kitchen, time.Now().Format(time.Kitchen))
	if err != nil {
		log.Printf("failed to parse current time: %s. Failing closed\n", err)
		return false
	}
	return startLateHours.Before(currentTime) && endLateHours.After(currentTime)
}

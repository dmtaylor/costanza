package listen

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/dmtaylor/costanza/config"
)

var startLateHours, endLateHours time.Time
var timeLoader sync.Once

func (s *Server) echoInsomniac(sess *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == sess.State.User.ID {
		return
	}

	if isAfterHours() && s.isInsomniacUser(m.Author, m.Member) {
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

func (s *Server) isInsomniacUser(user *discordgo.User, member *discordgo.Member) bool {
	if user == nil || member == nil {
		return false
	}

	for _, uid := range config.GlobalConfig.Discord.InsomniacIds {
		if user.ID == uid {
			return true
		}
	}

	for _, role := range config.GlobalConfig.Discord.InsomniacRoles {
		for _, userRole := range member.Roles {
			if role == userRole {
				return true
			}
		}
	}
	return false

}

func isAfterHours() bool {
	var err error
	timeLoader.Do(func() {
		startLateHours, err = time.Parse(time.Kitchen, "12:30AM")
		if err != nil {
			log.Panicf("error parsing start date format: %s\n", err)
		}
		endLateHours, err = time.Parse(time.Kitchen, "06:00AM")
		if err != nil {
			log.Panicf("error parsing end date format: %s\n", err)
		}
	})
	currentTime, err := time.Parse(time.Kitchen, time.Now().Format(time.Kitchen))
	if err != nil {
		log.Printf("failed to parse current time: %s. Failing closed\n", err)
		return false
	}
	return startLateHours.Before(currentTime) && endLateHours.After(currentTime)
}

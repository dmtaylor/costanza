package listen

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/exp/slog"

	"github.com/dmtaylor/costanza/config"
)

var startLateHours, endLateHours time.Time
var timeLoader sync.Once

func (s *Server) echoInsomniac(sess *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == sess.State.User.ID {
		return
	}
	ctx := context.WithValue(context.Background(), "messageId", m.ID)

	if isAfterHours(ctx) && s.isInsomniacUser(m.Author, m.Member) {
		_, err := sess.ChannelMessageSendReply(
			m.ChannelID,
			fmt.Sprintf("%s All right. That's enough for today. You're tired. Get some sleep. I'll see you first thing in the morning.",
				m.Author.Mention()),
			m.Reference(),
		)
		if err != nil {
			slog.ErrorCtx(ctx, fmt.Sprintf("error sending message: %s", err))
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

func isAfterHours(ctx context.Context) bool {
	var err error
	timeLoader.Do(func() {
		startLateHours, err = time.Parse(time.Kitchen, "12:30AM")
		if err != nil {
			slog.ErrorCtx(ctx, fmt.Sprintf("error parsing start date format: %s", err))
			panic(err)
		}
		endLateHours, err = time.Parse(time.Kitchen, "06:00AM")
		if err != nil {
			slog.ErrorCtx(ctx, fmt.Sprintf("error parsing end date format: %s", err))
			panic(err)
		}
	})
	currentTime, err := time.Parse(time.Kitchen, time.Now().Format(time.Kitchen))
	if err != nil {
		slog.ErrorCtx(ctx, fmt.Sprintf("failed to parse current time: %s", err))
		return false
	}
	return startLateHours.Before(currentTime) && endLateHours.After(currentTime)
}

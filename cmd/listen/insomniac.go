package listen

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/exp/slog"

	"github.com/dmtaylor/costanza/config"
	"github.com/dmtaylor/costanza/internal/util"
)

const insomniacEventName = "insomniac"

var startLateHours, endLateHours time.Time
var timeLoader sync.Once

func (s *Server) echoInsomniac(sess *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == sess.State.User.ID {
		return
	}

	if s.m.enabled {
		start := time.Now()
		defer func() {
			s.m.eventDuration.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent, eventNameLabel: insomniacEventName}).Observe(time.Since(start).Seconds())
		}()
	}
	ctx := util.ContextFromDiscordMessageCreate(context.Background(), m)

	if isAfterHours(ctx) && s.isInsomniacUser(ctx, m.Author, m.Member) {
		callStart := time.Now()
		_, err := sess.ChannelMessageSendReply(
			m.ChannelID,
			fmt.Sprintf("%s All right. That's enough for today. You're tired. Get some sleep. I'll see you first thing in the morning.",
				m.Author.Mention()),
			m.Reference(),
		)
		if s.m.enabled {
			s.m.externalApiDuration.With(prometheus.Labels{eventNameLabel: insomniacEventName, externalApiLabel: externalDiscordCallName}).Observe(time.Since(callStart).Seconds())
		}
		if err != nil {
			if s.m.enabled {
				s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent, eventNameLabel: insomniacEventName, isTimeoutLabel: "false"}).Inc()
			}
			slog.ErrorCtx(ctx, "error sending message: "+err.Error())
		} else {
			if s.m.enabled {
				s.m.eventSuccess.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent, eventNameLabel: insomniacEventName}).Inc()
			}
		}
		return
	}
	if s.m.enabled {
		s.m.eventSuccess.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent, eventNameLabel: insomniacEventName}).Inc()
	}
}

func (s *Server) isInsomniacUser(ctx context.Context, user *discordgo.User, member *discordgo.Member) bool {
	if user == nil || member == nil {
		slog.DebugCtx(ctx, "user or member is nil, skipping")
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
			slog.ErrorCtx(ctx, "error parsing start date format: "+err.Error())
			panic(err)
		}
		endLateHours, err = time.Parse(time.Kitchen, "06:00AM")
		if err != nil {
			slog.ErrorCtx(ctx, "error parsing end date format: "+err.Error())
			panic(err)
		}
	})
	currentTime, err := time.Parse(time.Kitchen, time.Now().Format(time.Kitchen))
	if err != nil {
		slog.WarnCtx(ctx, "failed to parse current time: "+err.Error())
		return false
	}
	return startLateHours.Before(currentTime) && endLateHours.After(currentTime)
}

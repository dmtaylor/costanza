package listen

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/exp/slog"
)

const welcomeMessageFmt = `Welcome to the party %s!`

func (s *Server) welcomeMessage(sess *discordgo.Session, j *discordgo.GuildMemberAdd) {
	if j.User.ID == sess.State.User.ID { // Don't welcome yourself
		return
	}
	if j.User.Bot { // Don't welcome robots
		return
	}
	if s.m.enabled {
		start := time.Now()
		defer func() {
			s.m.eventDuration.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent, eventNameLabel: "welcome"}).Observe(time.Since(start).Seconds())
		}()
	}
	ctx := context.WithValue(context.Background(), "memberId", j.User.ID)
	ctx = context.WithValue(ctx, "guildId", j.GuildID)

	channels, err := sess.GuildChannels(j.GuildID)
	if err != nil {
		if s.m.enabled {
			s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: guildMemberAddGatewayEvent, eventNameLabel: "welcome", isTimeoutLabel: "false"}).Inc()
		}
		slog.ErrorCtx(ctx, "error getting channel list: "+err.Error())
		return
	}
	if len(channels) < 1 {
		slog.WarnCtx(ctx, "no guild channels pulled, ignoring")
		return
	}
	for _, channel := range channels {
		if channel.Type == discordgo.ChannelTypeGuildText && channel.Position == 0 {
			_, err = sess.ChannelMessageSend(channel.ID, fmt.Sprintf(welcomeMessageFmt, j.User.Mention()))
			if err != nil {
				if s.m.enabled {
					s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: guildMemberAddGatewayEvent, eventNameLabel: "welcome", isTimeoutLabel: "false"}).Inc()
				}
				slog.ErrorCtx(ctx, "failed to send message: "+err.Error(), "channel", channel.ID)
			}
			break
		}
	}
	if s.m.enabled {
		s.m.eventSuccess.With(prometheus.Labels{gatewayEventTypeLabel: guildMemberAddGatewayEvent, eventNameLabel: "welcome"}).Inc()
	}
}

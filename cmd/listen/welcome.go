package listen

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"
)

const welcomeMessageFmt = `Welcome to the party %s!`
const welcomeEventName = "welcome"

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

	callStart := time.Now()
	channels, err := sess.GuildChannels(j.GuildID)
	if s.m.enabled {
		s.m.externalApiDuration.With(prometheus.Labels{eventNameLabel: welcomeEventName, externalApiLabel: externalDiscordCallName}).Observe(time.Since(callStart).Seconds())
	}
	if err != nil {
		if s.m.enabled {
			s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: guildMemberAddGatewayEvent, eventNameLabel: welcomeEventName, isTimeoutLabel: "false"}).Inc()
		}
		slog.ErrorContext(ctx, "error getting channel list: "+err.Error())
		return
	}
	if len(channels) < 1 {
		slog.WarnContext(ctx, "no guild channels pulled, ignoring")
		return
	}
	for _, channel := range channels {
		if channel.Type == discordgo.ChannelTypeGuildText && channel.Position == 0 {
			callStart = time.Now()
			_, err = sess.ChannelMessageSend(channel.ID, fmt.Sprintf(welcomeMessageFmt, j.User.Mention()))
			if s.m.enabled {
				s.m.externalApiDuration.With(prometheus.Labels{eventNameLabel: welcomeEventName, externalApiLabel: externalDiscordCallName}).Observe(time.Since(callStart).Seconds())
			}
			if err != nil {
				if s.m.enabled {
					s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: guildMemberAddGatewayEvent, eventNameLabel: "welcome", isTimeoutLabel: "false"}).Inc()
				}
				slog.ErrorContext(ctx, "failed to send message: "+err.Error(), "channel", channel.ID)
			}
			break
		}
	}
	if s.m.enabled {
		s.m.eventSuccess.With(prometheus.Labels{gatewayEventTypeLabel: guildMemberAddGatewayEvent, eventNameLabel: "welcome"}).Inc()
	}
}

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
	var err error
	if s.m.enabled {
		start := time.Now()
		defer func() {
			s.m.eventDuration.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent, eventNameLabel: "welcome"}).Observe(time.Since(start).Seconds())
			if err != nil {
				s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: guildMemberAddGatewayEvent, eventNameLabel: welcomeEventName, isTimeoutLabel: "false"}).Inc()
			} else {
				s.m.eventSuccess.With(prometheus.Labels{gatewayEventTypeLabel: guildMemberAddGatewayEvent, eventNameLabel: welcomeEventName}).Inc()
			}
		}()
	}
	ctx := context.WithValue(context.Background(), "memberId", j.User.ID)
	ctx = context.WithValue(ctx, "guildId", j.GuildID)
	ctx = context.WithValue(ctx, "type", "welcome")

	guild, err := sess.Guild(j.GuildID)
	if err != nil {
		slog.ErrorContext(ctx, "failed getting guild data: "+err.Error())
		return
	}
	if guild.SystemChannelFlags&discordgo.SystemChannelFlagsSuppressJoinNotifications != 0 {
		return
	}
	_, err = sess.ChannelMessageSend(guild.SystemChannelID, fmt.Sprintf(welcomeMessageFmt, j.User.Mention()))
	if err != nil {
		slog.ErrorContext(ctx, "failed to send welcome message: "+err.Error())
		return
	}
}

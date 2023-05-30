package listen

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/exp/slog"

	"github.com/dmtaylor/costanza/internal/util"
)

const dailyWinReactMetricName = "dailyWinReact"

// dailyWinReact performs reaction if it detects a win pattern in the message
func (s *Server) dailyWinReact(sess *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == sess.State.User.ID {
		return
	}

	if s.m.enabled {
		start := time.Now()
		defer func() {
			s.m.eventDuration.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent, eventNameLabel: dailyWinReactMetricName}).Observe(time.Since(start).Seconds())
		}()
	}
	ctx := util.ContextFromDiscordMessageCreate(context.Background(), m)

	for _, pattern := range s.dailyWinPatterns {
		if pattern.MatchString(m.Message.Content) {
			callStart := time.Now()
			err := sess.MessageReactionAdd(m.ChannelID, m.Message.ID, "💯")
			if err != nil {
				if s.m.enabled {
					s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent, eventNameLabel: dailyWinReactMetricName, isTimeoutLabel: "false"}).Inc()
					s.m.externalApiDuration.With(prometheus.Labels{eventNameLabel: dailyWinReactMetricName, externalApiLabel: externalDiscordCallName}).Observe(time.Since(callStart).Seconds())
				}
				slog.ErrorCtx(ctx, fmt.Sprintf("error adding reaction: %s", err))
				return
			}
			if s.m.enabled {
				s.m.eventSuccess.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent, eventNameLabel: dailyWinReactMetricName}).Inc()
				s.m.externalApiDuration.With(prometheus.Labels{eventNameLabel: dailyWinReactMetricName, externalApiLabel: externalDiscordCallName}).Observe(time.Since(callStart).Seconds())
			}
			return
		}
	}
	if s.m.enabled {
		s.m.eventSuccess.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent, eventNameLabel: dailyWinReactMetricName}).Inc()
	}
}

package listen

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/dmtaylor/costanza/config"
	"github.com/dmtaylor/costanza/internal/util"
)

const logActivityMetricEventName = "logActivity"
const logReactionMetricEventName = "logReaction"

func (s *Server) logMessageActivity(sess *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == sess.State.User.ID {
		return
	}

	if m.Author.Bot {
		return
	}
	// Only log stats if channel included in configs
	if _, found := config.GlobalConfig.Discord.ListenChannelSet[m.GuildID]; !found {
		return
	}

	if s.m.enabled {
		start := time.Now()
		defer func() {
			s.m.eventDuration.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent, eventNameLabel: logActivityMetricEventName}).Observe(time.Since(start).Seconds())
		}()
	}
	ctx := util.ContextFromDiscordMessageCreate(context.Background(), m)

	if m.Type == discordgo.MessageTypeDefault || m.Type == discordgo.MessageTypeReply {
		var err error
		defer func() {
			if err != nil && s.m.enabled {
				s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent, eventNameLabel: logActivityMetricEventName, isTimeoutLabel: "false"}).Inc()
			}
		}()
		guildId, err := strconv.ParseUint(m.GuildID, 10, 64)
		if err != nil {
			slog.ErrorContext(ctx, "error logging activity: "+err.Error())
			return
		}
		userId, err := strconv.ParseUint(m.Author.ID, 10, 64)
		if err != nil {
			slog.ErrorContext(ctx, "error logging activity: "+err.Error())
			return
		}
		err = s.app.Stats.LogActivity(ctx, guildId, userId, m.Timestamp.Format("2006-01"))
		if err != nil {
			slog.ErrorContext(ctx, "error creating activity log: "+err.Error())
			return
		}
	}
	if s.m.enabled {
		s.m.eventSuccess.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent, eventNameLabel: logActivityMetricEventName}).Inc()
	}
}

func (s *Server) logReactionActivity(sess *discordgo.Session, r *discordgo.MessageReactionAdd) {
	// Don't log bot reactions
	if r.UserID == sess.State.User.ID {
		return
	}
	if r.Member != nil && r.Member.User != nil && r.Member.User.Bot {
		return
	}
	// Only log stats if channel included in configs
	if _, found := config.GlobalConfig.Discord.ListenChannelSet[r.GuildID]; !found {
		return
	}
	if s.m.enabled {
		start := time.Now()
		defer func() {
			s.m.eventDuration.With(prometheus.Labels{gatewayEventTypeLabel: messageReactionAddGatewayEvent, eventNameLabel: logReactionMetricEventName}).Observe(time.Since(start).Seconds())
		}()
	}
	var err error
	defer func() {
		if err != nil && s.m.enabled {
			s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: messageReactionAddGatewayEvent, eventNameLabel: logReactionMetricEventName, isTimeoutLabel: "false"}).Inc()
		}
	}()
	ctx := util.ContextFromDiscordReactionAdd(context.Background(), r)
	guildId, err := strconv.ParseUint(r.GuildID, 10, 64)
	if err != nil {
		slog.ErrorContext(ctx, "error logging activity: "+err.Error())
		return
	}
	userId, err := strconv.ParseUint(r.UserID, 10, 64)
	if err != nil {
		slog.ErrorContext(ctx, "error logging activity: "+err.Error())
		return
	}
	err = s.app.Stats.LogReaction(ctx, guildId, userId, time.Now().Format("2006-01"))
	if err != nil {
		slog.ErrorContext(ctx, "error creating activity log: "+err.Error())
	} else {
		if s.m.enabled {
			s.m.eventSuccess.With(prometheus.Labels{gatewayEventTypeLabel: messageReactionAddGatewayEvent, eventNameLabel: logReactionMetricEventName}).Inc()
		}
	}
}

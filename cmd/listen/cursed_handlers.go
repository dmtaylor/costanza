package listen

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/exp/slices"

	"github.com/dmtaylor/costanza/config"
	"github.com/dmtaylor/costanza/internal/util"
)

const cursedChannelLogEventName = "cursed_channel"

var cursedChannelBaseLabels = prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent, eventNameLabel: cursedChannelLogEventName}

func (s *Server) logCursedChannelStat(sess *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot || m.Author.ID == sess.State.User.ID {
		return
	}
	if _, found := config.GlobalConfig.Discord.ListenChannelSet[m.GuildID]; !found {
		return
	}
	ctx := util.ContextFromDiscordMessageCreate(context.Background(), m)
	var err error
	if s.m.enabled {
		start := time.Now()
		defer func() {
			s.m.eventDuration.With(cursedChannelBaseLabels).Observe(time.Since(start).Seconds())
			if err != nil {
				var timeout = "false"
				if errors.Is(err, context.DeadlineExceeded) {
					timeout = "true"
				}
				s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent, eventNameLabel: cursedChannelLogEventName, isTimeoutLabel: timeout}).Inc()
			} else {
				s.m.eventSuccess.With(cursedChannelBaseLabels).Inc()
			}
		}()
	}
	guildId, err := strconv.ParseUint(m.GuildID, 10, 64)
	if err != nil {
		slog.ErrorContext(ctx, "error logging cursed channel: "+err.Error())
		return
	}
	userId, err := strconv.ParseUint(m.Author.ID, 10, 64)
	if err != nil {
		slog.ErrorContext(ctx, "error logging activity: "+err.Error())
		return
	}
	channelId, err := strconv.ParseUint(m.ChannelID, 10, 64)
	if err != nil {
		slog.ErrorContext(ctx, "error logging activity: "+err.Error())
		return
	}
	cursedChannels, err := s.app.CursedChannelCache.Get(ctx, guildId)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get cursed channel list: "+err.Error())
		return
	}
	if slices.Index(cursedChannels, channelId) != -1 {
		err = s.app.Stats.LogCursedChannelPost(ctx, guildId, userId, time.Now().Format("2006-01"))
		if err != nil {
			slog.ErrorContext(ctx, "failed to update cursed channel log: "+err.Error())
			return
		}
	}
}

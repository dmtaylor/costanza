package listen

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/go-multierror"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/dmtaylor/costanza/config"
	"github.com/dmtaylor/costanza/internal/stats"
	"github.com/dmtaylor/costanza/internal/util"
)

const logActivityMetricEventName = "logActivity"
const logReactionMetricEventName = "logReaction"

const leaderboardCommandName = "leaderboard"

const reportCount = 3 // update this count for number of reports pulled

var leaderboardSlashCommand = &discordgo.ApplicationCommand{
	Name:        leaderboardCommandName,
	Type:        discordgo.ChatApplicationCommand,
	Description: "Get the current guild leaderboard standings",
}

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

func (s *Server) getLeaderboardStats(sess *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand || i.ApplicationCommandData().Name != leaderboardCommandName {
		return
	}

	var err error
	if s.m.enabled {
		start := time.Now()
		defer func() {
			s.m.eventDuration.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent, eventNameLabel: leaderboardCommandName}).Observe(time.Since(start).Seconds())
			if err != nil {
				isTimeout := strconv.FormatBool(errors.Is(err, context.DeadlineExceeded))
				s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent, eventNameLabel: leaderboardCommandName, isTimeoutLabel: isTimeout}).Inc()
			} else {
				s.m.eventSuccess.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent, eventNameLabel: leaderboardCommandName}).Inc()
			}
		}()
	}
	ctx, cancel := util.ContextFromDiscordInteractionCreate(context.Background(), i, interactionTimeout)
	defer cancel()

	// if guild isn't configured to listen, send message saying so
	if _, ok := config.GlobalConfig.Discord.ListenChannelSet[i.GuildID]; !ok {
		err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Leaderboards aren't enabled on this guild. Please reach out to admin to enable",
			},
		})
		if err != nil {
			slog.ErrorContext(ctx, "failed to send empty response: "+err.Error())
		}
		return
	}
	err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{},
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to create deferred response: "+err.Error())
		return
	}

	guildId, err := strconv.ParseUint(i.GuildID, 10, 64)
	if err != nil {
		err = fmt.Errorf("failed to format guild id: %w", err)
		slog.ErrorContext(ctx, "bad guild id: "+err.Error())
		return
	}
	var wg sync.WaitGroup
	errs := make(chan error, reportCount)
	var merr *multierror.Error
	wg.Add(reportCount) // TODO update this value for number of stats to be pulled
	go func() {
		wg.Wait()
		close(errs)
	}()
	go func() { // Messages stats
		defer wg.Done()
		messageStats, ierr := s.app.Stats.GetLeaders(ctx, guildId, time.Now().Format("2006-01"))
		if ierr != nil {
			errs <- ierr
			return
		}
		msg := stats.BuildMessageReport(messageStats)
		_, ierr = sess.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
			Content: msg,
		})
		if ierr != nil {
			errs <- ierr
		}
	}()
	go func() { // Daily game stats
		defer wg.Done()
		gameStats, ierr := s.app.Stats.GetDailyGameLeaders(ctx, guildId, time.Now().Format("2006-01"))
		if ierr != nil {
			errs <- ierr
			return
		}
		msg := stats.BuildGameWinReport(gameStats)
		_, ierr = sess.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
			Content: msg,
		})
		if ierr != nil {
			errs <- ierr
		}
	}()
	go func() { // Reaction score report
		defer wg.Done()
		scores, ierr := s.app.Stats.GetReactionLeadersForMonth(ctx, guildId, time.Now().Format("2006-01"))
		if ierr != nil {
			errs <- ierr
			return
		}
		msg := stats.BuildReactionScoreReport(scores)
		_, ierr = sess.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
			Content: msg,
		})
		if ierr != nil {
			errs <- ierr
		}
	}()

	for e := range errs {
		slog.ErrorContext(ctx, "failed to pull stat: "+e.Error())
		merr = multierror.Append(merr, e)
	}
	err = merr.ErrorOrNil()

}

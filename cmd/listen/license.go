package listen

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/exp/slog"

	"github.com/dmtaylor/costanza/config"
	"github.com/dmtaylor/costanza/internal/util"
)

const licenseCommandName = "license"

var licenseSlashCommand = &discordgo.ApplicationCommand{
	Name:        licenseCommandName,
	Type:        discordgo.ChatApplicationCommand,
	Description: "Gets app info for costanza",
}

func (s *Server) license(sess *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	if i.User != nil && i.User.Bot {
		return
	}
	if i.Member != nil && i.Member.User.Bot {
		return
	}
	if i.ApplicationCommandData().Name != licenseCommandName {
		return
	}
	if s.m.enabled {
		start := time.Now()
		defer func() {
			s.m.eventDuration.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent, eventNameLabel: licenseCommandName}).Observe(time.Since(start).Seconds())
		}()
	}
	ctx, cancel := util.ContextFromDiscordInteractionCreate(context.Background(), i, interactionTimeout)
	defer cancel()
	callStart := time.Now()
	err := sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "costanza " + config.VersionString + "\nLicensed with Apache 2.0\nFor more information & source code: https://github.com/dmtaylor/costanza",
		},
	})
	if s.m.enabled {
		s.m.externalApiDuration.With(prometheus.Labels{eventNameLabel: licenseCommandName, externalApiLabel: externalDiscordCallName}).Observe(time.Since(callStart).Seconds())
		if err != nil {
			s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent, eventNameLabel: licenseCommandName, isTimeoutLabel: "false"}).Inc()
		} else {
			s.m.eventSuccess.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent, eventNameLabel: licenseCommandName}).Inc()
		}
	}
	if err != nil {
		slog.ErrorCtx(ctx, "failed sending license data: "+err.Error())
	}
}

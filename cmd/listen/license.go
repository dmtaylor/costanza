package listen

import (
	"context"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/dmtaylor/costanza/config"
)

const licenseCommandName = "license"

var commit = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}

	return ""
}()

var licenseSlashCommand = &discordgo.ApplicationCommand{
	Name:        licenseCommandName,
	Type:        discordgo.ChatApplicationCommand,
	Description: "Gets app info for costanza",
}

func (s *Server) license(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) {
	if s.m.enabled {
		start := time.Now()
		defer func() {
			s.m.eventDuration.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent, eventNameLabel: licenseCommandName}).Observe(time.Since(start).Seconds())
		}()
	}
	callStart := time.Now()
	err := sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "costanza " + config.VersionString + "." + commit + "\nLicensed with Apache 2.0\nFor more information & source code: https://github.com/dmtaylor/costanza",
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
		slog.ErrorContext(ctx, "failed sending license data: "+err.Error())
	}
}

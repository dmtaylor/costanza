package listen

import (
	"context"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"
)

const helpCommandName = "chelp"
const helpMessage string = `
Costanza commands:
` +
	"```" + `
/chelp:       this message.
/roll:        parse text as d-notation and evaluate expression.
/srroll:      parse text as d-notation, evaluate, and use result for Shadowrun roll.
/wodroll:     parse text as d-notation, evaluate, and use result for World of Darkness roll.
              Can be modified with '8again', '9again' and 'chance'. Rolls of < 1 dice are done as chance rolls.
/dhtest:      parse text as d-notation, evaluate, and use result for FF Warhammer 40k RPG roll (over-under on 1d100).
/weather:     get weather information for given location, or default
/leaderboard: print the leaderboard for the month so far for the given server, if configured
` +
	"```"

var helpSlashCommand = &discordgo.ApplicationCommand{
	Name:        helpCommandName,
	Type:        discordgo.ChatApplicationCommand,
	Description: "Get help info for costanza",
}

// help handler function for help messages
func (s *Server) help(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) {
	if s.m.enabled {
		start := time.Now()
		defer func() {
			s.m.eventDuration.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent, eventNameLabel: helpCommandName}).Observe(time.Since(start).Seconds())
		}()
	}
	slog.DebugContext(ctx, "running help command")
	callStart := time.Now()
	err := sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: helpMessage,
		},
	})
	if s.m.enabled {
		s.m.externalApiDuration.With(prometheus.Labels{eventNameLabel: helpCommandName, externalApiLabel: externalDiscordCallName}).Observe(time.Since(callStart).Seconds())
		if err != nil {
			s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent, eventNameLabel: helpCommandName, isTimeoutLabel: "false"}).Inc()
		} else {
			s.m.eventSuccess.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent, eventNameLabel: helpCommandName}).Inc()
		}
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed sending help data: "+err.Error())
	}
}

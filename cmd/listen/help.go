package listen

import (
	"context"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/dmtaylor/costanza/internal/util"
)

const helpCommandName = "chelp"
const helpMessage string = `
Costanza commands:
` +
	"```" + `
/chelp:   this message.
/roll:    parse text as d-notation and evaluate expression.
/srroll:  parse text as d-notation, evaluate, and use result for Shadowrun roll.
/wodroll: parse text as d-notation, evaluate, and use result for World of Darkness roll.
          Can be modified with '8again', '9again' and 'chance'. Rolls of < 1 dice are done as chance rolls.
/dhtest:  parse text as d-notation, evaluate, and use result for FF Warhammer 40k RPG roll (over-under on 1d100).
/weather: get weather information for given location, or default
` +
	"```"

var helpSlashCommand = &discordgo.ApplicationCommand{
	Name:        helpCommandName,
	Type:        discordgo.ChatApplicationCommand,
	Description: "Get help info for costanza",
}

// help handler function for help messages
func (s *Server) help(sess *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	if i.User != nil && i.User.Bot {
		return
	}
	if i.Member != nil && i.Member.User.Bot {
		return
	}
	if i.ApplicationCommandData().Name != helpCommandName {
		return
	}
	if s.m.enabled {
		start := time.Now()
		defer func() {
			s.m.eventDuration.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent, eventNameLabel: helpCommandName}).Observe(time.Since(start).Seconds())
		}()
	}
	ctx, cancel := util.ContextFromDiscordInteractionCreate(context.Background(), i, interactionTimeout)
	defer cancel()
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

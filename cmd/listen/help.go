package listen

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/exp/slog"
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
func help(sess *discordgo.Session, i *discordgo.InteractionCreate) {
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
	ctx, cancel := context.WithTimeout(context.Background(), interactionTimeout)
	defer cancel()
	ctx = context.WithValue(ctx, "interactionId", i.ID)
	ctx = context.WithValue(ctx, "guildId", i.GuildID)
	ctx = context.WithValue(ctx, "commandName", helpCommandName)
	slog.DebugCtx(ctx, "running help command")
	err := sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: helpMessage,
		},
	})
	if err != nil {
		slog.ErrorCtx(ctx, "failed sending help data: "+err.Error())
	}
}

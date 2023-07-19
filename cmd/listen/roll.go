package listen

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/exp/slog"

	"github.com/dmtaylor/costanza/internal/roller"
	"github.com/dmtaylor/costanza/internal/util"
)

const rollCommandName = "roll"
const shadowrunCommandName = "srroll"
const worldOfDarknessCommandName = "wodroll"
const darkHeresyTestCommandName = "dhtest"
const rollOptionName = "roll"

var rollCommands = map[string]bool{
	rollCommandName:            true,
	shadowrunCommandName:       true,
	worldOfDarknessCommandName: true,
	darkHeresyTestCommandName:  true,
}

var rollSlashCommand = &discordgo.ApplicationCommand{
	Name:        rollCommandName,
	Type:        discordgo.ChatApplicationCommand,
	Description: "Parse and execute d-notation roll",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        rollOptionName,
			Description: "Value to roll",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    true,
		},
	},
}

var shadowrunRollSlashCommand = &discordgo.ApplicationCommand{
	Name:        shadowrunCommandName,
	Type:        discordgo.ChatApplicationCommand,
	Description: "Parse and execute d-notation roll as a Shadowrun test",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        rollOptionName,
			Description: "Value to roll",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    true,
		},
	},
}

var worldOfDarknessCommand = &discordgo.ApplicationCommand{
	Name:        worldOfDarknessCommandName,
	Type:        discordgo.ChatApplicationCommand,
	Description: "Parse and execute d-notation roll as a World of Darkness test, including optional modifiers",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        rollOptionName,
			Description: "Value to roll",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    true,
		},
		{
			Name:        "chance",
			Description: "Is a chance roll",
			Type:        discordgo.ApplicationCommandOptionBoolean,
			Required:    false,
		},
		{
			Name:        "9again",
			Description: "Roll has 9-again modifier",
			Type:        discordgo.ApplicationCommandOptionBoolean,
			Required:    false,
		},
		{
			Name:        "8again",
			Description: "Roll has 8-again modifier",
			Type:        discordgo.ApplicationCommandOptionBoolean,
			Required:    false,
		},
	},
}

var darkHeresyTestSlashCommand = &discordgo.ApplicationCommand{
	Name:        darkHeresyTestCommandName,
	Type:        discordgo.ChatApplicationCommand,
	Description: "Parse and execute d-notation roll as a Dark Heresy skill test",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        rollOptionName,
			Description: "Value to roll",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    true,
		},
	},
}

// dispatchRollCommands Main entrypoint into handling roll commands. Reads the first word of the message content
// and calls the appropriate method for performing a roll. Update this to add additional message prefixes for additional
// roll types.
func (s *Server) dispatchRollCommands(sess *discordgo.Session, i *discordgo.InteractionCreate) {
	// Ensure we only get options from slash commands
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	if i.User != nil && i.User.Bot {
		return
	}
	if i.Member != nil && i.Member.User.Bot {
		return
	}
	cmdName := i.ApplicationCommandData().Name
	if _, ok := rollCommands[cmdName]; !ok { // stop running if not a roll command
		return
	}

	if s.m.enabled {
		start := time.Now()
		defer func() {
			s.m.eventDuration.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent, eventNameLabel: cmdName}).Observe(time.Since(start).Seconds())
		}()
	}
	ctx, cancel := util.ContextFromDiscordInteractionCreate(context.Background(), i, interactionTimeout)
	defer cancel()
	options := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(i.ApplicationCommandData().Options))
	for _, option := range i.ApplicationCommandData().Options {
		options[option.Name] = option
	}
	var rollInput string
	if o, ok := options[rollOptionName]; ok {
		rollInput = o.StringValue()
	}
	slog.DebugCtx(ctx, "starting roll", "roll", rollInput)

	var result string
	var err error
	switch cmdName {
	case rollCommandName:
		if rollInput == "" {
			if s.m.enabled {
				s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent, eventNameLabel: cmdName, isTimeoutLabel: "false"}).Inc()
			}
			slog.ErrorCtx(ctx, "missing roll input for interaction")
			return
		}
		result, err = s.doDNotationRoll(rollInput)
	case shadowrunCommandName:
		if rollInput == "" {
			if s.m.enabled {
				s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent, eventNameLabel: cmdName, isTimeoutLabel: "false"}).Inc()
			}
			slog.ErrorCtx(ctx, "missing roll input for interaction")
			return
		}
		result, err = s.doShadowrunRoll(rollInput)
	case worldOfDarknessCommandName:
		result, err = s.doWodRoll(rollInput, options) // pass in option set for WoD specific options
		if o, ok := options["chance"]; ok {
			if o.BoolValue() {
				rollInput = rollInput + " chance"
			}
		}
		if o, ok := options["8again"]; ok {
			if o.BoolValue() {
				rollInput = rollInput + " 8again"
			}
		}
		if o, ok := options["9again"]; ok {
			if o.BoolValue() {
				rollInput = rollInput + " 9again"
			}
		}
	case darkHeresyTestCommandName:
		result, err = s.doDHTestRoll(rollInput)
	default:
		if s.m.enabled {
			s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent, eventNameLabel: cmdName, isTimeoutLabel: "false"}).Inc()
		}
		slog.ErrorCtx(ctx, "invalid command name: "+cmdName)
		return
	}
	timeoutErr := util.CheckCtxTimeout(ctx)
	if err != nil {
		slog.ErrorCtx(ctx, "failed to process roll: "+err.Error(), "roll", rollInput)
		if timeoutErr != nil { // don't send response if context timed out
			if s.m.enabled {
				s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent, eventNameLabel: cmdName, isTimeoutLabel: "true"}).Inc()
			}
			slog.ErrorCtx(ctx, "context err: "+timeoutErr.Error())
			return
		}
		callStart := time.Now()
		err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("I was unable to handle your roll \"%s\". Why must there always be a problem?", rollInput),
			},
		})
		if s.m.enabled {
			s.m.externalApiDuration.With(prometheus.Labels{eventNameLabel: cmdName, externalApiLabel: externalDiscordCallName}).Observe(time.Since(callStart).Seconds())
		}
		if err != nil {
			if s.m.enabled {
				s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent, eventNameLabel: cmdName, isTimeoutLabel: "false"}).Inc()
			}
			slog.ErrorCtx(ctx, "failed to send interaction response: "+err.Error())
		}
		return
	}
	if timeoutErr != nil { // don't finish if context timed out
		if s.m.enabled {
			s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent, eventNameLabel: cmdName, isTimeoutLabel: "true"}).Inc()
		}
		slog.ErrorCtx(ctx, "context err: "+timeoutErr.Error())
		return
	}
	callStart := time.Now()
	err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%s → %s", rollInput, result),
		},
	})
	if s.m.enabled {
		s.m.externalApiDuration.With(prometheus.Labels{eventNameLabel: cmdName, externalApiLabel: externalDiscordCallName}).Observe(time.Since(callStart).Seconds())
	}
	if err != nil {
		if s.m.enabled {
			s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent, eventNameLabel: cmdName, isTimeoutLabel: "false"}).Inc()
		}
		slog.ErrorCtx(ctx, "failed to send interaction response: "+err.Error())
		return
	}
	slog.DebugCtx(ctx, "completed roll", "roll", rollInput)
	if s.m.enabled {
		s.m.eventSuccess.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent, eventNameLabel: cmdName}).Inc()
	}
}

func (s *Server) doDNotationRoll(input string) (string, error) {
	res, err := s.app.DNotationParser.DoParse(input)
	if err != nil {
		return "", fmt.Errorf("failed to parse roll: %w", err)
	}
	return fmt.Sprintf("%s = %d", res.StrValue, res.Value), nil
}

func (s *Server) doShadowrunRoll(input string) (string, error) {
	rollCount, err := s.app.DNotationParser.DoParse(input)
	if err != nil {
		return "", fmt.Errorf("failed to parse roll %s, %w", input, err)
	}
	params := roller.GetSrParams()
	rollResult, err := s.app.ThresholdRoller.DoThresholdRoll(rollCount.Value, roller.SrDieSides, params)
	if err != nil {
		return "", fmt.Errorf("failed to run threshold roll: %w", err)
	}
	rollRepr, err := rollResult.String()
	if err != nil {
		return "", fmt.Errorf("failed to get roll representation from %s: %w", input, err)
	}
	var res strings.Builder
	hitStr := "hit"
	if rollResult.Value() != 1 {
		hitStr = hitStr + "s"
	}
	_, err = res.WriteString(fmt.Sprintf("%s = %d %s", rollRepr, rollResult.Value(), hitStr))
	if err != nil {
		return "", fmt.Errorf("failed to write repr string %s to buffer: %w", rollRepr, err)
	}
	switch roller.GetGlitchStatus(rollResult) {
	case roller.SrGlitch:
		_, err = res.WriteString("\nYou glitched! I can't believe this! What was wrong with it? What didn't you like about it?")
	case roller.SrCritGlitch:
		_, err = res.WriteString("\nYou critically glitched! I don't want hope. Hope is killing me. My dream is to become hopeless. When you're hopeless, you don't care, and when you don't care, that indifference makes you attractive.")
	}
	if err != nil {
		return "", fmt.Errorf("failed to write optional glitch status to buffer: %w", err)
	}
	return res.String(), nil
}

func (s *Server) doWodRoll(input string, options map[string]*discordgo.ApplicationCommandInteractionDataOption) (string, error) {
	var isChance, isEightAgain, isNineAgain bool
	if o, ok := options["chance"]; ok {
		isChance = o.BoolValue()
	}
	if o, ok := options["8again"]; ok {
		isEightAgain = o.BoolValue()
	}
	if o, ok := options["9again"]; ok {
		isNineAgain = o.BoolValue()
	}
	params := roller.NewGetWodRollParams(isNineAgain, isEightAgain)
	if isChance {
		return s.doWodChanceRoll(params)
	} else {
		rollCount, err := s.app.DNotationParser.DoParse(input)
		if err != nil {
			return "", fmt.Errorf("failed to parse roll input %s: %w", input, err)
		}
		if rollCount.Value < 1 {
			return s.doWodChanceRoll(params)
		}
		roll, err := s.app.ThresholdRoller.DoThresholdRoll(rollCount.Value, roller.WodDieSides, params)
		if err != nil {
			return "", fmt.Errorf("failed to get wod threshold roll for %d dice: %w", rollCount.Value, err)
		}
		rollResStr, err := roll.String()
		if err != nil {
			return "", fmt.Errorf("failed to get representation for roll %v: %w", roll, err)
		}

		hitStr := "hit"
		if roll.Value() != 1 {
			hitStr = hitStr + "s"
		}
		result := fmt.Sprintf("%s = %d %s", rollResStr, roll.Value(), hitStr)
		if roll.Value() == 0 {
			result = result + "\nWould you like to critically fail?"
		}
		return result, nil
	}
}

func (s *Server) doWodChanceRoll(params roller.ThresholdParameters) (string, error) {
	roll, err := s.app.ThresholdRoller.DoThresholdRoll(1, roller.WodDieSides, params)
	if err != nil {
		return "", fmt.Errorf("failed to execute chance roll: %w", err)
	}
	rollResStr, err := roll.String()
	if err != nil {
		return "", fmt.Errorf("failed to get string representation of roll: %w", err)
	}
	result := fmt.Sprintf("%s = %d", rollResStr, roll.Value())
	if roll.Value() < 1 {
		result = result + "\nYou critically failed! Radiating waves of pain."
	}

	return result, nil
}

func (s *Server) doDHTestRoll(input string) (string, error) {
	threshold, err := s.app.DNotationParser.DoParse(input)
	if err != nil {
		return "", fmt.Errorf("failed to parse input %s: %w", input, err)
	}
	roll, err := s.app.DNotationParser.DoParse("1d100")
	if err != nil {
		return "", fmt.Errorf("failed to get dh test roll: %w", err)
	}
	if roll.Value > threshold.Value {
		return fmt.Sprintf("Rolled %s: you fail with %d degrees", roll.StrValue, (roll.Value-threshold.Value)/10), nil
	} else {
		return fmt.Sprintf("Rolled %s: you succeed with %d degrees", roll.StrValue, (threshold.Value-roll.Value)/10), nil
	}
}

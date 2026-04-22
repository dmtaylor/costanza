package listen

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/dmtaylor/costanza/config"
	"github.com/dmtaylor/costanza/internal/util"
)

const cursedChannelLogEventName = "cursed_channel"
const cursedWordLogEventName = "cursed_post"
const cursedAdminCommandName = "cursed-admin"
const cursedAdminListSubcommand = "list"
const cursedAdminAddSubcommand = "add"
const cursedAdminRemoveSubcommand = "remove"

var cursedChannelBaseLabels = prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent, eventNameLabel: cursedChannelLogEventName}
var cursedWordBaseLabels = prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent, eventNameLabel: cursedWordLogEventName}

var cursedAdminSlashCommand = &discordgo.ApplicationCommand{
	Name:        cursedAdminCommandName,
	Type:        discordgo.ChatApplicationCommand,
	Description: "Manage cursed words on the server",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        cursedAdminListSubcommand,
			Description: "List of cursed words on the server",
		},
		{
			Name:        cursedAdminAddSubcommand,
			Description: "Add to the cursed words on the server",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "word",
					Type:        discordgo.ApplicationCommandOptionString,
					Description: "Word to add to the cursed words on the server",
					Required:    true,
				},
			},
		},
		{
			Name:        cursedAdminRemoveSubcommand,
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Description: "Remove from the cursed words on the server",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "word",
					Type:        discordgo.ApplicationCommandOptionString,
					Description: "Word to remove from the cursed words on the server",
					Required:    true,
				},
			},
		},
	},
}

func (s *Server) processCursedAdminCommand(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) {
	var err error = nil
	if i.Member.Permissions&discordgo.PermissionManageGuild != discordgo.PermissionManageGuild {
		err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You must have permissions to manage the guild to run this command",
			},
		})
		if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Failed sending interaction response: %v", err), "error", err)
			return
		}
	}
	options := i.ApplicationCommandData().Options
	topName := options[0].Name
	if s.m.enabled {
		start := time.Now()
		defer func() {
			s.m.eventDuration.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent, eventNameLabel: cursedAdminCommandName + "." + topName}).Observe(time.Since(start).Seconds())
		}()
	}
	switch topName {
	case cursedAdminAddSubcommand:
		innerOptions := options[0].Options
		newValue := innerOptions[0].StringValue()
		slog.DebugContext(ctx, "Adding new word", "newValue", newValue)
		err = s.app.CursedStatsHandler.AddWordToCursedList(ctx, util.MustSnowflakeToInt(i.GuildID), newValue)
		if err != nil {
			if s.m.enabled {
				s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent, eventNameLabel: cursedAdminCommandName + "." + cursedAdminAddSubcommand, isTimeoutLabel: "false"}).Inc()
			}
			slog.ErrorContext(ctx, fmt.Sprintf("Error adding word to cursed list: %v", err), "error", err)
			err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Failed to add word to cursed list"),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Failed sending interaction response: %v", err), "error", err)
				return
			}
			return
		}
		err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Added word to cursed list"),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Failed sending interaction response: %v", err), "error", err)
			return
		}
	case cursedAdminRemoveSubcommand:
		innerOptions := options[0].Options
		valueToRemove := innerOptions[0].StringValue()
		slog.DebugContext(ctx, "Removing word from cursed list", "value", valueToRemove)
		err = s.app.CursedStatsHandler.RemoveWordFromCursedList(ctx, util.MustSnowflakeToInt(i.GuildID), valueToRemove)
		if err != nil {
			if s.m.enabled {
				s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent, eventNameLabel: cursedAdminCommandName + "." + cursedAdminRemoveSubcommand, isTimeoutLabel: "false"}).Inc()
			}
			slog.ErrorContext(ctx, fmt.Sprintf("Error adding word to cursed list: %v", err), "error", err)
			err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Failed to remove word from cursed list"),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Failed sending interaction response: %v", err), "error", err)
				return
			}
			return
		}
		err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Removed word from cursed list"),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Failed sending interaction response: %v", err), "error", err)
			return
		}
	case cursedAdminListSubcommand:
		wordList, err := s.app.CursedStatsHandler.GetCursedWordsList(ctx, util.MustSnowflakeToInt(i.GuildID))
		if err != nil {
			if s.m.enabled {
				s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent, eventNameLabel: cursedAdminCommandName + "." + cursedAdminListSubcommand, isTimeoutLabel: "false"}).Inc()
			}
			slog.ErrorContext(ctx, fmt.Sprintf("Error adding word to cursed list: %v", err), "error", err)
			err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Failed to get cursed word list"),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Failed sending interaction response: %v", err), "error", err)
				return
			}
			return
		}
		strBuilder := strings.Builder{}
		strBuilder.WriteString(fmt.Sprintf("Cursed words list:\n"))
		if len(wordList) > 0 {
			for _, word := range wordList {
				strBuilder.WriteString(fmt.Sprintf("* `%s`\n", word))
			}
		} else {
			strBuilder.WriteString("No words in the list")
		}
		err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: strBuilder.String(),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Failed sending interaction response: %v", err), "error", err)
			return
		}

	}
	if s.m.enabled {
		s.m.eventSuccess.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent, eventNameLabel: cursedAdminCommandName + "." + options[0].Name}).Inc()
	}
}

func (s *Server) logCursedChannelStat(ctx context.Context, sess *discordgo.Session, m *discordgo.MessageCreate) {
	if _, found := config.GlobalConfig.Discord.ListenChannelSet[m.GuildID]; !found {
		return
	}
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
	cursedChannels, err := s.app.CursedStatsHandler.GetCursedChannelList(ctx, guildId)
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

func (s *Server) logCursedPostStat(ctx context.Context, sess *discordgo.Session, m *discordgo.MessageCreate) {
	if _, found := config.GlobalConfig.Discord.ListenChannelSet[m.GuildID]; !found {
		return
	}
	var err error
	if s.m.enabled {
		start := time.Now()
		defer func() {
			s.m.eventDuration.With(cursedWordBaseLabels).Observe(time.Since(start).Seconds())
			if err != nil {
				var timeout = "false"
				if errors.Is(err, context.DeadlineExceeded) {
					timeout = "true"
				}
				s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent, eventNameLabel: cursedWordLogEventName, isTimeoutLabel: timeout}).Inc()
			} else {
				s.m.eventSuccess.With(cursedWordBaseLabels).Inc()
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
	cursedWords, err := s.app.CursedStatsHandler.GetCursedWordsList(ctx, guildId)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get cursed word list: "+err.Error())
		return
	}
	msg := strings.ToLower(m.Message.Content)
	count := 0
	for _, word := range cursedWords {
		count += strings.Count(msg, word)
	}
	if count > 0 {
		err = s.app.Stats.LogCursedPost(ctx, guildId, userId, time.Now().Format("2006-01"), count)
		if err != nil {
			slog.ErrorContext(ctx, "failed to update cursed post log: "+err.Error())
			return
		}
	}
}

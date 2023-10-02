package listen

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/dmtaylor/costanza/internal/util"
)

type DailyGamePlay struct {
	GuildId uint64
	UserId  uint64
	Tries   uint
	Win     bool
}

const dailyGameHandlerEventName = "dailyGameHandler"
const dailyGameReactionEventName = "dailyGameReaction"

var gamePattern = regexp.MustCompile(`(?s)(Framed|Tradle|Wordle|Heardle|GuessTheGame|Episode)\s+.*#?\d+.*[🟩⬛⬜🟥]`)
var wordleAndTradleCapturePattern = regexp.MustCompile(`(?s)#?(Tradle|Wordle)\s.*#?\d+\s+(\d+|X)/(\d+)`)

// dailyGameHandler performs handling of daily game events
func (s *Server) dailyGameHandler(sess *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == sess.State.User.ID {
		return
	}

	if s.m.enabled {
		start := time.Now()
		defer func() {
			s.m.eventDuration.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent, eventNameLabel: dailyGameHandlerEventName}).Observe(time.Since(start).Seconds())
		}()
	}
	ctx := util.ContextFromDiscordMessageCreate(context.Background(), m)

	if groups := gamePattern.FindStringSubmatch(m.Content); groups != nil {
		guildId, err := strconv.ParseUint(m.GuildID, 10, 64)
		if err != nil {
			if s.m.enabled {
				s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent, eventNameLabel: dailyGameHandlerEventName, isTimeoutLabel: "false"}).Inc()
			}
			slog.ErrorContext(ctx, "failed to parse guild id as uint64", "guildId", m.GuildID)
			return
		}
		userId, err := strconv.ParseUint(m.Author.ID, 10, 64)
		if err != nil {
			if s.m.enabled {
				s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent, eventNameLabel: dailyGameHandlerEventName, isTimeoutLabel: "false"}).Inc()
			}
			slog.ErrorContext(ctx, "failed to parse user id as uint64", "userId", m.Author.ID)
			return
		}
		slog.DebugContext(ctx, "matched game pattern", "message", m.Content, "userId", userId, "guildId", guildId)
		gameResult, err := createGameResult(guildId, userId, groups[1], m.Content)
		if err != nil {
			if s.m.enabled {
				s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent, eventNameLabel: dailyGameHandlerEventName, isTimeoutLabel: "false"}).Inc()
			}
			slog.ErrorContext(ctx, "failed to get game results", "error", err.Error())
			return
		}
		slog.DebugContext(ctx, "parsed game results", "gameResults", fmt.Sprintf("%+v", gameResult))
		var wg sync.WaitGroup
		var reactionError error
		if gameResult.Tries == 1 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				reactionError = s.doWinReaction(sess, m)
			}()
		}
		// TODO add stats here

		wg.Wait()
		if reactionError != nil {
			if s.m.enabled {
				s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent, eventNameLabel: dailyGameHandlerEventName, isTimeoutLabel: "false"}).Inc()
			}
			slog.ErrorContext(ctx, fmt.Sprintf("error adding reaction: %s", reactionError))
			return
		}

	}

	if s.m.enabled {
		s.m.eventSuccess.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent, eventNameLabel: dailyGameHandlerEventName}).Inc()
	}
}

func (s *Server) doWinReaction(sess *discordgo.Session, m *discordgo.MessageCreate) error {
	callStart := time.Now()
	err := sess.MessageReactionAdd(m.ChannelID, m.Message.ID, "💯")
	if err != nil {
		if s.m.enabled {
			s.m.externalApiDuration.With(prometheus.Labels{eventNameLabel: dailyGameReactionEventName, externalApiLabel: externalDiscordCallName}).Observe(time.Since(callStart).Seconds())
		}
		return fmt.Errorf("failed to add reaction: %w", err)
	}

	return nil
}

func isGameMessage(message string) bool {
	return gamePattern.MatchString(message)
}

func createGameResult(guildId, userId uint64, gameType, message string) (DailyGamePlay, error) {
	result := DailyGamePlay{
		guildId,
		userId,
		0,
		false,
	}
	switch gameType {
	case "Framed":
		fallthrough
	case "Heardle":
		fallthrough
	case "GuessTheGame":
		fallthrough
	case "Episode":
		for _, r := range []rune(message) {
			if r == '🟥' {
				result.Tries += 1
			} else if r == '🟩' {
				result.Tries += 1
				result.Win = true
				break
			}
		}
	case "Tradle":
		fallthrough
	case "Wordle":
		groups := wordleAndTradleCapturePattern.FindStringSubmatch(message)
		if groups == nil {
			return result, fmt.Errorf("invalid wordle/tradle match \"%s\"", message)
		}
		total, err := strconv.ParseUint(groups[3], 10, 32)
		if err != nil {
			return result, fmt.Errorf("failed parsing total: %w", err)
		}
		if groups[2] == "X" {
			result.Tries = uint(total)
		} else {
			guesses, err := strconv.ParseUint(groups[2], 10, 32)
			if err != nil {
				return result, fmt.Errorf("failed parsing guesses: %w", err)
			}
			result.Tries = uint(guesses)
			result.Win = true
		}
	default:
		return result, fmt.Errorf("invalid game type: %s", gameType)
	}

	return result, nil
}

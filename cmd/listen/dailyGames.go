package listen

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
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

const dailyWinReactMetricName = "dailyGameHandler"

var gamePattern = regexp.MustCompile(`(?s)(Framed|Tradle|Wordle|Heardle|GuessTheGame|Episode)\s+.*#?\d+.*[🟩⬛⬜🟥]`)
var wordleAndTradleCapturePattern = regexp.MustCompile(`(?s)#(Tradle|Wordle)\s.*#?\d+\s+(\d+|X)/(\d+)`)

// dailyGameHandler performs handling of
func (s *Server) dailyGameHandler(sess *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == sess.State.User.ID {
		return
	}

	if s.m.enabled {
		start := time.Now()
		defer func() {
			s.m.eventDuration.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent, eventNameLabel: dailyWinReactMetricName}).Observe(time.Since(start).Seconds())
		}()
	}
	ctx := util.ContextFromDiscordMessageCreate(context.Background(), m)

	for _, pattern := range s.dailyWinPatterns {
		if pattern.MatchString(m.Message.Content) {
			callStart := time.Now()
			err := sess.MessageReactionAdd(m.ChannelID, m.Message.ID, "💯")
			if err != nil {
				if s.m.enabled {
					s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent, eventNameLabel: dailyWinReactMetricName, isTimeoutLabel: "false"}).Inc()
					s.m.externalApiDuration.With(prometheus.Labels{eventNameLabel: dailyWinReactMetricName, externalApiLabel: externalDiscordCallName}).Observe(time.Since(callStart).Seconds())
				}
				slog.ErrorContext(ctx, fmt.Sprintf("error adding reaction: %s", err))
				return
			}
			if s.m.enabled {
				s.m.eventSuccess.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent, eventNameLabel: dailyWinReactMetricName}).Inc()
				s.m.externalApiDuration.With(prometheus.Labels{eventNameLabel: dailyWinReactMetricName, externalApiLabel: externalDiscordCallName}).Observe(time.Since(callStart).Seconds())
			}
			return
		}
	}
	if s.m.enabled {
		s.m.eventSuccess.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent, eventNameLabel: dailyWinReactMetricName}).Inc()
	}
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

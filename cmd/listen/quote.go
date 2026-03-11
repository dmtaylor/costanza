package listen

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/dmtaylor/costanza/internal/model"
)

const quoteEventName = "quote"

// echoQuote handler function for sending George Costanza quotes
func (s *Server) echoQuote(ctx context.Context, sess *discordgo.Session, m *discordgo.MessageCreate) {
	if s.m.enabled {
		start := time.Now()
		defer func() {
			s.m.eventDuration.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent, eventNameLabel: quoteEventName}).Observe(time.Since(start).Seconds())
		}()
	}

	for _, mentionedUser := range m.Mentions {
		if mentionedUser.ID == sess.State.User.ID {
			err := s.sendQuote(ctx, sess, m)
			if s.m.enabled {
				if err != nil {
					s.m.eventErrors.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent, eventNameLabel: quoteEventName, isTimeoutLabel: "false"}).Inc()
				} else {
					s.m.eventSuccess.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent, eventNameLabel: quoteEventName}).Inc()
				}
			}
			return
		}
	}
	if s.m.enabled {
		s.m.eventSuccess.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent, eventNameLabel: quoteEventName}).Inc()
	}
}

func (s *Server) sendQuote(ctx context.Context, sess *discordgo.Session, m *discordgo.MessageCreate) error {
	quoteData, err := s.app.Quotes.GetQuoteSql(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get quote: "+err.Error())
		_, err := sess.ChannelMessageSendReply(
			m.ChannelID,
			"I was unable to get a quote. Why must there always be a problem?",
			m.Reference(),
		)
		if err != nil {
			slog.ErrorContext(ctx, "error sending message: "+err.Error())
		}
		return err
	}
	switch quoteData.Type {
	case model.TextQuoteType:
		callStart := time.Now()
		_, err = sess.ChannelMessageSendReply(m.ChannelID, quoteData.Data, m.Reference())
		if s.m.enabled {
			s.m.externalApiDuration.With(prometheus.Labels{eventNameLabel: quoteEventName, externalApiLabel: externalDiscordCallName}).Observe(time.Since(callStart).Seconds())
		}
		if err != nil {
			slog.ErrorContext(ctx, "error sending message: "+err.Error())
			return err
		}
	case model.FileQuoteType:
		err = s.sendFileQuote(ctx, sess, m, quoteData)
		if err != nil {
			slog.ErrorContext(ctx, "error sending message: "+err.Error())
			return err
		}
	default:
		slog.ErrorContext(ctx, "invalid quote type in model", "quoteData", quoteData)
		return model.InvalidQuoteTypeError(quoteData.Type)
	}
	return nil
}

func (s *Server) sendFileQuote(ctx context.Context, sess *discordgo.Session, m *discordgo.MessageCreate, quoteEntry model.Quote) error {
	file, err := os.Open(quoteEntry.Data)
	if err != nil {
		return fmt.Errorf("failed to open attachment file: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			slog.ErrorContext(ctx, "failed to close attachment file", "err", err)
		}
	}(file)
	callStart := time.Now()
	_, err = sess.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Files: []*discordgo.File{{
			Name:        filepath.Base(quoteEntry.Data),
			ContentType: strings.Replace(".", "", filepath.Ext(quoteEntry.Data), 1),
			Reader:      file,
		}},
		Reference: m.Reference(),
	})
	if s.m.enabled {
		s.m.externalApiDuration.With(prometheus.Labels{eventNameLabel: quoteEventName, externalApiLabel: externalDiscordCallName}).Observe(time.Since(callStart).Seconds())
	}

	return err
}

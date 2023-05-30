package listen

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/exp/slog"

	"github.com/dmtaylor/costanza/internal/util"
)

const quoteEventName = "quote"

// echoQuote handler function for sending George Costanza quotes
func (s *Server) echoQuote(sess *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == sess.State.User.ID {
		return
	}
	if s.m.enabled {
		start := time.Now()
		defer func() {
			s.m.eventDuration.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent, eventNameLabel: quoteEventName}).Observe(time.Since(start).Seconds())
		}()
	}
	ctx := util.ContextFromDiscordMessageCreate(context.Background(), m)

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
	quote, err := s.app.Quotes.GetQuoteSql(ctx)
	if err != nil {
		slog.ErrorCtx(ctx, "failed to get quote: "+err.Error())
		_, err := sess.ChannelMessageSendReply(
			m.ChannelID,
			"I was unable to get a quote. Why must there always be a problem?",
			m.Reference(),
		)
		if err != nil {
			slog.ErrorCtx(ctx, "error sending message: ", err.Error())
		}
		return err
	}
	callStart := time.Now()
	_, err = sess.ChannelMessageSendReply(m.ChannelID, quote, m.Reference())
	if s.m.enabled {
		s.m.externalApiDuration.With(prometheus.Labels{eventNameLabel: quoteEventName, externalApiLabel: externalDiscordCallName}).Observe(time.Since(callStart).Seconds())
	}
	if err != nil {
		slog.ErrorCtx(ctx, "error sending message: "+err.Error())
		return err
	}
	return nil
}

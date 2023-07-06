package listen

import (
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// gatewayEventTypeLabel label name for the gateway event type in vector metrics
const gatewayEventTypeLabel = "gateway_event_type"

// eventNameLabel label name for the event that is happening
const eventNameLabel = "event_name"

// isTimeoutLabel used in measuring errors if the error results from a context timeout for interaction responses
const isTimeoutLabel = "is_timeout"

// externalApiLabel is the promethus label name for the destination of an external API call
const externalApiLabel = "external_api_dest"

// messageCreateGatewayEvent is the gateway event type for a received message
const messageCreateGatewayEvent = "messageCreate"

// interactionCreateGatewayEvent is the gateway event type for an interaction creation
const interactionCreateGatewayEvent = "interactionCreate"

// guildMemberAddGatewayEvent is the gateway event type when a user joins a guild
const guildMemberAddGatewayEvent = "guildMemberAdd"

// externalDiscordCallName used for external API calls to Discord
const externalDiscordCallName = "discord"

type metrics struct {
	enabled             bool
	eventReceives       *prometheus.CounterVec   // Number of events received by Discord gateway
	eventErrors         *prometheus.CounterVec   // Number of errors in event handlers. Logged by handler func
	eventsHandled       *prometheus.CounterVec   // Total number of gateway events handled
	eventSuccess        *prometheus.CounterVec   // Number of successful gateway events. Logged by handler func
	eventDuration       *prometheus.HistogramVec // Total duration of event handler function in seconds. Only log in handlers that are ran to avoid skewing with very short metrics
	externalApiDuration *prometheus.HistogramVec // Total duration of external API calls in seconds
}

func (s *Server) messageCreateMetricsMiddleware(f func(*discordgo.Session, *discordgo.MessageCreate)) func(*discordgo.Session, *discordgo.MessageCreate) {
	return func(sess *discordgo.Session, m *discordgo.MessageCreate) {
		if s.m.enabled {
			s.m.eventReceives.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent}).Inc()
			defer s.m.eventsHandled.With(prometheus.Labels{gatewayEventTypeLabel: messageCreateGatewayEvent}).Inc()
		}
		f(sess, m)
	}
}

func (s *Server) interactionCreateMetricsMiddleware(f func(*discordgo.Session, *discordgo.InteractionCreate)) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(sess *discordgo.Session, i *discordgo.InteractionCreate) {
		if s.m.enabled {
			s.m.eventReceives.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent}).Inc()
			defer s.m.eventsHandled.With(prometheus.Labels{gatewayEventTypeLabel: interactionCreateGatewayEvent}).Inc()
		}
		f(sess, i)
	}
}

func (s *Server) guildMemberAddMetricsMiddleware(f func(*discordgo.Session, *discordgo.GuildMemberAdd)) func(*discordgo.Session, *discordgo.GuildMemberAdd) {
	return func(sess *discordgo.Session, j *discordgo.GuildMemberAdd) {
		if s.m.enabled {
			s.m.eventReceives.With(prometheus.Labels{gatewayEventTypeLabel: "guildMemberAdd"}).Inc()
			defer s.m.eventsHandled.With(prometheus.Labels{gatewayEventTypeLabel: "guildMemberAdd"}).Inc()
		}
		f(sess, j)
	}
}

// setupMetrics configures prometheus metrics & modifies the Server object to support logging.
// Metrics should only be logged if this function has been run, and metrics.enabled should only be set to true here.
func (s *Server) setupMetrics() http.Handler {
	eventReceives := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costanza",
		Name:      "event_receives",
		Help:      "Total number of gateway events received",
	}, []string{gatewayEventTypeLabel})
	eventErrors := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costanza",
		Name:      "event_errors",
		Help:      "Number of errored events",
	}, []string{gatewayEventTypeLabel, eventNameLabel, isTimeoutLabel})
	eventsHandled := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costanza",
		Name:      "events_handled",
		Help:      "Total number of gateway events finished being handled",
	}, []string{gatewayEventTypeLabel})
	eventSuccesses := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costanza",
		Name:      "event_successes",
		Help:      "Number of successfully processed events",
	}, []string{gatewayEventTypeLabel, eventNameLabel})
	eventDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "costanza",
		Name:      "event_duration_seconds",
		Help:      "Processing time for non-ignored events",
		Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 0.75, 1.0, 1.25, 1.5, 2, 2.5, 5},
	}, []string{gatewayEventTypeLabel, eventNameLabel})
	externalApiDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "costanza",
		Name:      "external_api_time_seconds",
		Help:      "Time for external API calls",
		Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 0.75, 1.0, 1.25, 1.5, 2, 2.5, 5},
	}, []string{eventNameLabel, externalApiLabel})

	reg := prometheus.NewRegistry()

	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{Namespace: "costanza"}),
		eventReceives,
		eventErrors,
		eventsHandled,
		eventSuccesses,
		eventDuration,
		externalApiDuration,
	)

	s.m = metrics{
		enabled:             true,
		eventReceives:       eventReceives,
		eventErrors:         eventErrors,
		eventsHandled:       eventsHandled,
		eventSuccess:        eventSuccesses,
		eventDuration:       eventDuration,
		externalApiDuration: externalApiDuration,
	}
	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg})
}

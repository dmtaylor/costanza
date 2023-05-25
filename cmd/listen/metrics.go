package listen

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type metrics struct {
	enabled       bool
	eventReceives *prometheus.CounterVec   // Number of events received by Discord gateway
	eventErrors   *prometheus.CounterVec   // Number of errors in event handlers. Logged by handler func
	eventsHandled *prometheus.CounterVec   // Total number of gateway events handled
	eventSuccess  *prometheus.CounterVec   // Number of successful gateway events. Logged by handler func
	eventDuration *prometheus.HistogramVec // Total duration of event handler function in seconds. Only log in handlers that are ran
}

// TODO add metrics here

func (s *Server) setupMetrics() http.Handler {
	eventReceives := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costanza",
		Name:      "event_receives",
		Help:      "Total number of gateway events received",
	}, []string{"gateway_event_type"})
	eventErrors := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costanza",
		Name:      "event_errors",
		Help:      "Number of errored events",
	}, []string{"gateway_event_type", "event_name", "is_timeout"})
	eventsHandled := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costanza",
		Name:      "events_handled",
		Help:      "Total number of gateway events finished being handled",
	}, []string{"gateway_event_type"})
	eventSuccesses := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costanza",
		Name:      "event_successes",
		Help:      "Number of successfully processed events",
	}, []string{"gateway_event_type", "event_name"})
	eventDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "costanza",
		Name:      "event_duration_seconds",
		Help:      "Processing time for non-ignored events",
		Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 0.75, 1.0, 1.25, 1.5, 2, 2.5, 5},
	}, []string{"gateway_event_type", "event_name"})

	reg := prometheus.NewRegistry()

	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{Namespace: "costanza"}),
		eventReceives,
		eventErrors,
		eventsHandled,
		eventSuccesses,
		eventDuration,
	)

	s.m = metrics{
		enabled:       true,
		eventReceives: eventReceives,
		eventErrors:   eventErrors,
		eventsHandled: eventsHandled,
		eventSuccess:  eventSuccesses,
		eventDuration: eventDuration,
	}
	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg})
}

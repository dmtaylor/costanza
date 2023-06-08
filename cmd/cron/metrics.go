package cron

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const listenGuildIdLabel = "listenGuildId"

type metrics struct {
	enabled               bool
	failedReports         *prometheus.CounterVec
	successfulReports     *prometheus.CounterVec
	reportDurationSeconds *prometheus.HistogramVec
}

func (c *cronConfig) setupMetrics() http.Handler {

	failedReports := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costanza",
		Subsystem: "cron",
		Name:      "failed_reports",
		Help:      "Number of failed reports sent",
	}, []string{listenGuildIdLabel})
	successfulReports := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "costanza",
		Subsystem: "cron",
		Name:      "successful_reports",
		Help:      "Number of successful reports",
	}, []string{listenGuildIdLabel})
	reportDurationSeconds := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "costanza",
		Subsystem: "cron",
		Name:      "report_duration_seconds",
		Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 0.75, 1.0, 1.25, 1.5, 2, 2.5, 5},
	}, []string{listenGuildIdLabel})
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{Namespace: "costanza_cron"}),
		failedReports,
		successfulReports,
		reportDurationSeconds,
	)
	c.m = metrics{
		enabled:               true,
		failedReports:         failedReports,
		successfulReports:     successfulReports,
		reportDurationSeconds: reportDurationSeconds,
	}

	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg})
}

package config

import (
	"os"
	"strings"

	"golang.org/x/exp/slog"
)

func getLogLevel(lvl string) slog.Level {
	switch strings.ToLower(lvl) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func initializeLogger() {
	level := getLogLevel(GlobalConfig.Metrics.LogLevel)
	h := slog.HandlerOptions{Level: level}
	if level < slog.LevelInfo {
		h.AddSource = true
	}
	slog.SetDefault(slog.New(h.NewTextHandler(os.Stderr)).With("appname", GlobalConfig.Metrics.Appname))
}

package config

import (
	"log/slog"
	"os"
	"strings"
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
	h := &slog.HandlerOptions{Level: level}
	if level < slog.LevelInfo {
		h.AddSource = true
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, h)).With("appname", GlobalConfig.Metrics.Appname))
}

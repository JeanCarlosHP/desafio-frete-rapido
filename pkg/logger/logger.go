package logger

import (
	"log/slog"
	"os"

	"github.com/jeancarloshp/desafio-frete-rapido/pkg/config"
)

func New(config *config.Config) {
	var logger *slog.Logger

	handler := &slog.HandlerOptions{}

	setLoggingLevel(handler, config.LoggingLevel)

	if config.LoggingJSONFormat {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, handler))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, handler))
	}

	slog.SetDefault(logger)
}

func setLoggingLevel(handler *slog.HandlerOptions, level string) {
	switch level {
	case "DEBUG":
		handler.Level = slog.LevelDebug
	case "INFO":
		handler.Level = slog.LevelInfo
	case "WARNING":
		handler.Level = slog.LevelWarn
	case "ERROR":
		handler.Level = slog.LevelError
	}
}

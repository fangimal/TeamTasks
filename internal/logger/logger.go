package logger

import (
	"io"
	"log/slog"
	"os"

	"github.com/fangimal/TeamTask/internal/config"
)

func New(config config.LoggerConfig) *slog.Logger {
	options := &slog.HandlerOptions{
		Level: parseLevel(config.Level),
	}

	var handler slog.Handler
	output := io.Writer(os.Stdout)

	if config.Format == "text" {
		handler = slog.NewTextHandler(output, options)
	} else {
		handler = slog.NewJSONHandler(output, options)
	}

	return slog.New(handler)
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

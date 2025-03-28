package logger

import (
	"io"
	"log/slog"
)

type Logger struct {
	*slog.Logger
}

func New(out io.Writer, level slog.Level) *Logger {
	var handler slog.Handler

	options := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler = slog.NewTextHandler(out, options)

	return &Logger{
		Logger: slog.New(handler),
	}
}

package logger

import (
	"io"
	"log/slog"
)

func New(out io.Writer, level slog.Level) *slog.Logger {
	options := &slog.HandlerOptions{Level: level} // Используем переданный level
	handler := slog.NewTextHandler(out, options)
	return slog.New(handler)
}

package logging

import (
	"log/slog"
	"os"
)

// NewLogger returns a structured logger (JSON in production, human-readable in dev).
func NewLogger() *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(handler)
}

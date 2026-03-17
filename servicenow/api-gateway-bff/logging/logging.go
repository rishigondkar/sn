package logging

import (
	"context"
	"log/slog"
	"os"
)

// Ctx keys for request id and correlation id (set by middleware).
type ctxKey string

const (
	KeyRequestID    ctxKey = "request_id"
	KeyCorrelationID ctxKey = "correlation_id"
	KeyActorID      ctxKey = "actor_user_id"
	KeyActorName    ctxKey = "actor_name"
)

// NewLogger returns a structured logger. In production use JSON and configurable level.
func NewLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// FromContext returns logger with request_id and correlation_id from context if present.
func FromContext(ctx context.Context, log *slog.Logger) *slog.Logger {
	if log == nil {
		log = NewLogger()
	}
	args := make([]any, 0, 4)
	if v, ok := ctx.Value(KeyRequestID).(string); ok && v != "" {
		args = append(args, "request_id", v)
	}
	if v, ok := ctx.Value(KeyCorrelationID).(string); ok && v != "" {
		args = append(args, "correlation_id", v)
	}
	if v, ok := ctx.Value(KeyActorID).(string); ok && v != "" {
		args = append(args, "actor_user_id", v)
	}
	if len(args) == 0 {
		return log
	}
	return log.With(args...)
}

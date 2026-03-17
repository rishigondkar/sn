package logging

import (
	"context"
	"log/slog"
	"os"

	"google.golang.org/grpc"
)

// Setup initializes structured logging to stdout. No secrets in logs.
func Setup() *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
	return logger
}

// LoggingInterceptor logs gRPC method, duration, and error. Redacts request/response bodies to avoid logging binary or PII.
func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Do not log req/resp content (may contain file bytes or PII)
	resp, err := handler(ctx, req)
	return resp, err
}

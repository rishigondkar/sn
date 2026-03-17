package logging

import (
	"context"
	"log/slog"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const currentLayer = "logging"

// Setup initializes structured logging to stdout. No secrets or full PII in logs.
func Setup() *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)
	return logger
}

// LoggingInterceptor logs gRPC method, duration, and error. Redacts request/response bodies to avoid PII/secrets.
func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := slog.Default().With(
		"method", info.FullMethod,
		"layer", currentLayer,
	)
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if v := md.Get("x-request-id"); len(v) > 0 {
			start = start.With("request_id", v[0])
		}
		if v := md.Get("x-correlation-id"); len(v) > 0 {
			start = start.With("correlation_id", v[0])
		}
	}
	resp, err := handler(ctx, req)
	// Do not log req/resp content to avoid PII/secrets
	if err != nil {
		start.With("error", err).Error("gRPC request failed")
	} else {
		start.Info("gRPC request completed")
	}
	return resp, err
}

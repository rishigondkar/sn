package logging

import (
	"context"
	"log/slog"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// SetupLogging configures structured logging to stdout. Returns a logger and no cleanup needed.
func SetupLogging(_ context.Context) *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
	return logger
}

// LoggingInterceptor logs gRPC method, duration, and error. Does not log request/response bodies
// to avoid PII/secrets; only method name and metadata keys (not values) for attribution.
func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start)

	grpcAttrs := []any{"method", info.FullMethod, "duration_ms", duration}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if len(md.Get("x-request-id")) > 0 {
			grpcAttrs = append(grpcAttrs, "has_request_id", true)
		}
		if len(md.Get("x-correlation-id")) > 0 {
			grpcAttrs = append(grpcAttrs, "has_correlation_id", true)
		}
	}
	if err != nil {
		slog.Default().ErrorContext(ctx, "grpc request failed", slog.Any("error", err), slog.Group("grpc", grpcAttrs...))
		return resp, err
	}
	slog.Default().InfoContext(ctx, "grpc request", slog.Group("grpc", grpcAttrs...))
	return resp, nil
}

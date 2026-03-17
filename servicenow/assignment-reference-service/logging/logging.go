package logging

import (
	"context"
	"log/slog"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const layerKey = "layer"

// SetupLogging configures structured logging to stdout. Returns a logger and cleanup (no-op for now).
func SetupLogging(_ context.Context) (*slog.Logger, func()) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)
	return logger, func() {}
}

// LoggingInterceptor logs gRPC method, duration, and error. Does not log request/response bodies
// to avoid PII/secrets (P0: no secrets in logs).
func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	md, _ := metadata.FromIncomingContext(ctx)
	// Log only method and metadata keys (not values that might be tokens/PII)
	requestID := firstMeta(md, "x-request-id")
	correlationID := firstMeta(md, "x-correlation-id")

	resp, err := handler(ctx, req)
	dur := time.Since(start)
	if err != nil {
		slog.Error("grpc request failed",
			"method", info.FullMethod,
			"request_id", requestID,
			"correlation_id", correlationID,
			"duration_ms", dur.Milliseconds(),
			"error", err)
		return resp, err
	}
	slog.Info("grpc request",
		"method", info.FullMethod,
		"request_id", requestID,
		"correlation_id", correlationID,
		"duration_ms", dur.Milliseconds())
	return resp, nil
}

func firstMeta(md metadata.MD, key string) string {
	if v := md.Get(key); len(v) > 0 {
		return v[0]
	}
	return ""
}

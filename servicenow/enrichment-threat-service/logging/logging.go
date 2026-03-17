package logging

import (
	"context"
	"log/slog"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Setup initializes structured logging to stdout. No secrets in logs.
func Setup() *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	if os.Getenv("LOG_LEVEL") == "DEBUG" {
		opts.Level = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	slog.SetDefault(logger)
	return logger
}

// LoggingInterceptor logs gRPC requests: method, duration, error. Does not log request/response bodies (may contain PII/secrets).
func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	md, _ := metadata.FromIncomingContext(ctx)
	requestID := firstMeta(md, "x-request-id")
	correlationID := firstMeta(md, "x-correlation-id")

	resp, err := handler(ctx, req)
	dur := time.Since(start)
	if err != nil {
		st, _ := status.FromError(err)
		slog.Error("grpc request",
			"method", info.FullMethod,
			"duration_ms", dur.Milliseconds(),
			"code", st.Code(),
			"request_id", requestID,
			"correlation_id", correlationID,
		)
		return resp, err
	}
	slog.Info("grpc request",
		"method", info.FullMethod,
		"duration_ms", dur.Milliseconds(),
		"request_id", requestID,
		"correlation_id", correlationID,
	)
	return resp, err
}

func firstMeta(md metadata.MD, key string) string {
	if v := md.Get(key); len(v) > 0 {
		return v[0]
	}
	return ""
}

// RedactForLog returns a safe string for logging (e.g. truncate or redact PII). Used for error messages.
func RedactForLog(s string, maxLen int) string {
	if maxLen <= 0 {
		maxLen = 200
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

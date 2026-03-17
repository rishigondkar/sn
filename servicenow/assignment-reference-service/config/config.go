package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds service configuration from environment.
type Config struct {
	// DB connection string (e.g. postgres://user:pass@host:5432/dbname?sslmode=disable)
	DatabaseURL string
	// gRPC server listen address (default :50051)
	GRPCAddr string
	// HTTP server listen address for REST + health (default :8080)
	HTTPAddr string
	// Server read/write timeouts
	HTTPReadTimeout  time.Duration
	HTTPWriteTimeout time.Duration
	// Shutdown timeout for graceful drain
	ShutdownTimeout time.Duration
	// DB operation timeout
	DBTimeout time.Duration
	// AuditServiceAddr is the audit service gRPC address (e.g. localhost:50056); if set, publish audit events.
	AuditServiceAddr string
}

// Load loads configuration from environment.
func Load() *Config {
	c := &Config{
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://cursor:qwerty@localhost:5432/assignment_ref?sslmode=disable"),
		GRPCAddr:         getEnv("GRPC_ADDR", ":50054"), // default 50054 for single-host with gateway
		HTTPAddr:         getEnv("HTTP_ADDR", ":8080"),
		HTTPReadTimeout:   durationEnv("HTTP_READ_TIMEOUT", 15*time.Second),
		HTTPWriteTimeout:  durationEnv("HTTP_WRITE_TIMEOUT", 15*time.Second),
		ShutdownTimeout:   durationEnv("SHUTDOWN_TIMEOUT", 30*time.Second),
		DBTimeout:         durationEnv("DB_TIMEOUT", 10*time.Second),
		AuditServiceAddr:   getEnv("AUDIT_SERVICE_ADDR", ""),
	}
	return c
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func durationEnv(key string, d time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return d
}

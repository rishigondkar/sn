package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds service configuration from environment.
type Config struct {
	// DBURL is the PostgreSQL connection string (e.g. postgres://user:pass@host:5432/dbname).
	DBURL string
	// GRPCPort is the gRPC server port (default 50051).
	GRPCPort int
	// HTTPPort is the HTTP server port for REST and health (default 8080).
	HTTPPort int
	// AuditServiceAddr is the audit service gRPC address (e.g. localhost:50056); if set, publish audit events.
	AuditServiceAddr string
	// MaxPayloadBytes is the maximum allowed request body size for enrichment/threat payloads (default 1MB).
	MaxPayloadBytes int
	// ShutdownTimeout is how long to wait for graceful shutdown (default 30s).
	ShutdownTimeout time.Duration
	// ServerReadTimeout for HTTP/gRPC server (default 15s).
	ServerReadTimeout time.Duration
	// ServerWriteTimeout for HTTP server (default 15s).
	ServerWriteTimeout time.Duration
}

// Load reads configuration from environment.
func Load() *Config {
	c := &Config{
		DBURL:             getEnv("DATABASE_URL", "postgres://cursor:qwerty@localhost:5432/enrichment_threat?sslmode=disable"),
		GRPCPort:          getEnvInt("GRPC_PORT", 50053),
		HTTPPort:          getEnvInt("HTTP_PORT", 8080),
		AuditServiceAddr:  getEnv("AUDIT_SERVICE_ADDR", ""),
		MaxPayloadBytes:   getEnvInt("MAX_PAYLOAD_BYTES", 1024*1024),
		ShutdownTimeout:   30 * time.Second,
		ServerReadTimeout: 15 * time.Second,
		ServerWriteTimeout: 15 * time.Second,
	}
	if v := os.Getenv("SHUTDOWN_TIMEOUT_SEC"); v != "" {
		if sec, err := strconv.Atoi(v); err == nil && sec > 0 {
			c.ShutdownTimeout = time.Duration(sec) * time.Second
		}
	}
	return c
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}

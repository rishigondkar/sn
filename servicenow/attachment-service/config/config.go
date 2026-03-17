package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds service configuration from environment.
type Config struct {
	// Server
	GRPCPort string
	HTTPPort string

	// Database
	DatabaseURL string

	// Storage (object storage)
	StorageProvider string
	StorageBucket   string
	StorageTimeout  time.Duration
	StorageRetries  int

	// Audit (optional)
	AuditServiceAddr string

	// Limits and validation
	MaxFileSizeBytes int64
	AllowedTypes     []string // empty means allow all except DeniedTypes
	DeniedTypes      []string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// Load reads configuration from environment.
func Load() *Config {
	c := &Config{
		GRPCPort:         getEnv("GRPC_PORT", "50055"),
		HTTPPort:         getEnv("HTTP_PORT", "8080"),
		AuditServiceAddr: getEnv("AUDIT_SERVICE_ADDR", ""),
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://cursor:qwerty@localhost:5432/attachments?sslmode=disable"),
		StorageProvider: getEnv("STORAGE_PROVIDER", "s3"),
		StorageBucket:   getEnv("STORAGE_BUCKET", "attachments"),
		StorageTimeout:  getDurationEnv("STORAGE_TIMEOUT", 30*time.Second),
		StorageRetries:  getIntEnv("STORAGE_RETRIES", 3),
		MaxFileSizeBytes: getInt64Env("MAX_FILE_SIZE_BYTES", 50*1024*1024), // 50 MiB default
		ReadTimeout:     getDurationEnv("HTTP_READ_TIMEOUT", 15*time.Second),
		WriteTimeout:    getDurationEnv("HTTP_WRITE_TIMEOUT", 60*time.Second),
		ShutdownTimeout: getDurationEnv("SHUTDOWN_TIMEOUT", 30*time.Second),
	}
	c.AllowedTypes = getSliceEnv("ALLOWED_CONTENT_TYPES", nil)
	c.DeniedTypes = getSliceEnv("DENIED_CONTENT_TYPES", []string{})
	return c
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getIntEnv(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func getInt64Env(key string, def int64) int64 {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i
		}
	}
	return def
}

func getDurationEnv(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func getSliceEnv(key string, def []string) []string {
	if v := os.Getenv(key); v != "" {
		parts := strings.Split(v, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			if t := strings.TrimSpace(p); t != "" {
				out = append(out, t)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	return def
}

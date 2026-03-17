package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds application configuration from environment.
type Config struct {
	// Server
	GRPCPort int
	HTTPPort int

	// Database
	DatabaseURL string

	// Timeouts
	GRPCServerTimeout time.Duration
	HTTPServerTimeout  time.Duration
	DBTimeout          time.Duration

	// Consumer (event bus)
	ConsumerMaxRetries int
	ConsumerRetryDelay time.Duration
	ConsumerBatchSize  int
}

// Load reads configuration from environment variables.
func Load() *Config {
	c := &Config{
		GRPCPort:            getInt("GRPC_PORT", 50056), // default 50056 for single-host with gateway
		HTTPPort:             getInt("HTTP_PORT", 8080),
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://cursor:qwerty@localhost:5432/audit?sslmode=disable"),
		GRPCServerTimeout:    getDuration("GRPC_SERVER_TIMEOUT", 30*time.Second),
		HTTPServerTimeout:    getDuration("HTTP_SERVER_TIMEOUT", 30*time.Second),
		DBTimeout:            getDuration("DB_TIMEOUT", 10*time.Second),
		ConsumerMaxRetries:   getInt("CONSUMER_MAX_RETRIES", 3),
		ConsumerRetryDelay:   getDuration("CONSUMER_RETRY_DELAY", 2*time.Second),
		ConsumerBatchSize:    getInt("CONSUMER_BATCH_SIZE", 10),
	}
	return c
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultVal
}

func getDuration(key string, defaultVal time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultVal
}

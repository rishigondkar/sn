package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds service configuration from environment.
type Config struct {
	GRPCPort         int
	HTTPPort         int
	DBURL            string
	DBTimeout        time.Duration
	ServerTimeout    time.Duration
	ShutdownGrace    time.Duration
	EventBusTopic    string // e.g. soc-audit-events; optional for POC
	AuditServiceAddr string // e.g. localhost:50056; if set, case-service publishes audit events via gRPC
}

// Load reads configuration from environment.
func Load() *Config {
	c := &Config{
		GRPCPort:         intEnv("GRPC_PORT", 50051),
		HTTPPort:         intEnv("HTTP_PORT", 8080),
		DBURL:            getEnv("DATABASE_URL", "postgres://cursor:qwerty@localhost:5432/case_db?sslmode=disable"),
		DBTimeout:        10 * time.Second,
		ServerTimeout:    30 * time.Second,
		ShutdownGrace:    30 * time.Second,
		EventBusTopic:    os.Getenv("AUDIT_EVENT_TOPIC"),
		AuditServiceAddr: os.Getenv("AUDIT_SERVICE_ADDR"),
	}
	if t := os.Getenv("DB_TIMEOUT_SEC"); t != "" {
		if sec, err := strconv.Atoi(t); err == nil && sec > 0 {
			c.DBTimeout = time.Duration(sec) * time.Second
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

func intEnv(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultVal
}

package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds service configuration (env or file).
type Config struct {
	GRPCPort         int
	HTTPPort         int
	DB               DBConfig
	AuditServiceAddr string // e.g. localhost:50056; if set, publish audit events via gRPC
}

// DBConfig holds database connection settings.
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DSN returns the PostgreSQL connection string.
func (d DBConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode)
}

// Load reads configuration from environment.
func Load() *Config {
	return &Config{
		GRPCPort:         getEnvInt("GRPC_PORT", 50052),
		HTTPPort:         getEnvInt("HTTP_PORT", 8082),
		AuditServiceAddr: getEnv("AUDIT_SERVICE_ADDR", ""),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "cursor"),
			Password: getEnv("DB_PASSWORD", "qwerty"),
			DBName:   getEnv("DB_NAME", "alert_observable"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}
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

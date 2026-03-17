package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds application configuration.
type Config struct {
	HTTPPort        int
	ShutdownTimeout time.Duration

	// Downstream gRPC addresses (e.g. "case-service:50051")
	CaseServiceAddr        string
	ObservableServiceAddr  string
	EnrichmentServiceAddr  string
	ReferenceServiceAddr   string
	AttachmentServiceAddr  string
	AuditServiceAddr       string

	DownstreamTimeout time.Duration

	// CORS
	CORSAllowedOrigins []string
	CORSAllowedMethods []string

	// Auth (POC: trusted headers from upstream)
	TrustAuthHeaders bool
}

// Load reads config from environment with defaults.
func Load() (*Config, error) {
	cfg := &Config{
		HTTPPort:           8080,
		ShutdownTimeout:    30 * time.Second,
		DownstreamTimeout:  15 * time.Second,
		TrustAuthHeaders:   true,
		CORSAllowedOrigins: []string{"*"},
		CORSAllowedMethods: []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
	}

	if v := os.Getenv("HTTP_PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("HTTP_PORT: %w", err)
		}
		cfg.HTTPPort = p
	}

	if v := os.Getenv("SHUTDOWN_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("SHUTDOWN_TIMEOUT: %w", err)
		}
		cfg.ShutdownTimeout = d
	}

	if v := os.Getenv("DOWNSTREAM_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("DOWNSTREAM_TIMEOUT: %w", err)
		}
		cfg.DownstreamTimeout = d
	}

	cfg.CaseServiceAddr = getEnv("CASE_SERVICE_ADDR", "localhost:50051")
	cfg.ObservableServiceAddr = getEnv("OBSERVABLE_SERVICE_ADDR", "localhost:50052")
	cfg.EnrichmentServiceAddr = getEnv("ENRICHMENT_SERVICE_ADDR", "localhost:50053")
	cfg.ReferenceServiceAddr = getEnv("REFERENCE_SERVICE_ADDR", "localhost:50054")
	cfg.AttachmentServiceAddr = getEnv("ATTACHMENT_SERVICE_ADDR", "localhost:50055")
	cfg.AuditServiceAddr = getEnv("AUDIT_SERVICE_ADDR", "localhost:50056")

	if v := os.Getenv("CORS_ALLOWED_ORIGINS"); v != "" {
		cfg.CORSAllowedOrigins = strings.Split(v, ",")
	}
	if v := os.Getenv("CORS_ALLOWED_METHODS"); v != "" {
		cfg.CORSAllowedMethods = strings.Split(v, ",")
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

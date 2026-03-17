package service

import (
	"github.com/servicenow/enrichment-threat-service/config"
	"github.com/servicenow/enrichment-threat-service/repository"
)

// Service holds business logic dependencies.
type Service struct {
	Repo     *repository.Repository
	Audit    AuditEventPublisher
	Config   *config.Config
}

// NewService returns a Service with the given dependencies.
func NewService(repo *repository.Repository, audit AuditEventPublisher, cfg *config.Config) *Service {
	return &Service{Repo: repo, Audit: audit, Config: cfg}
}

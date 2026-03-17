package service

import (
	"github.com/servicenow/assignment-reference-service/repository"
)

// Service holds business logic dependencies.
type Service struct {
	Repo    *repository.Repository
	Audit   AuditPublisher
}

// New creates a Service with the given repository and optional audit publisher.
func New(repo *repository.Repository, audit AuditPublisher) *Service {
	if audit == nil {
		audit = NoopAuditPublisher{}
	}
	return &Service{Repo: repo, Audit: audit}
}

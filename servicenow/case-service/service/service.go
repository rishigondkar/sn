package service

import (
	"github.com/servicenow/case-service/audit"
	"github.com/servicenow/case-service/repository"
)

const currentLayer = "service"

// Service holds business logic dependencies.
type Service struct {
	Repo     *repository.Repository
	AuditPub audit.Publisher
}

// New creates a Service.
func New(repo *repository.Repository, auditPub audit.Publisher) *Service {
	return &Service{Repo: repo, AuditPub: auditPub}
}

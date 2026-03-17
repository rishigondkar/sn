package service

import (
	"github.com/servicenow/audit-service/repository"
)

const currentLayer = "service"

// Service holds business logic dependencies.
type Service struct {
	Repo *repository.Repository
}

// NewService returns a new Service.
func NewService(repo *repository.Repository) *Service {
	return &Service{Repo: repo}
}

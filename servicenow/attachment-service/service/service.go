package service

import (
	"context"
	"time"

	"github.com/soc-platform/attachment-service/audit"
	"github.com/soc-platform/attachment-service/domain"
	"github.com/soc-platform/attachment-service/storage"
)

const currentLayer = "service"

// AttachmentRepository abstracts attachment persistence for testability.
type AttachmentRepository interface {
	CreateAttachment(ctx context.Context, a *domain.Attachment) error
	GetByID(ctx context.Context, id string) (*domain.Attachment, error)
	ListByCaseID(ctx context.Context, caseID string, pageSize int32, pageToken string, includeDeleted bool) ([]*domain.Attachment, string, error)
	SoftDelete(ctx context.Context, id string, deletedAt time.Time) error
}

// Service holds business logic dependencies.
type Service struct {
	Repo     AttachmentRepository
	Storage  storage.Store
	Audit    audit.Publisher
	Bucket   string
	MaxBytes int64
	Allowed  []string
	Denied   []string
}

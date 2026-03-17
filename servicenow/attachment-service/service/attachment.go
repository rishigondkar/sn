package service

import (
	"context"
	"log/slog"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/soc-platform/attachment-service/audit"
	"github.com/soc-platform/attachment-service/domain"
	"github.com/soc-platform/attachment-service/storage"
)

// CreateAttachment uploads content to storage then persists metadata. On DB failure we attempt compensation (delete from storage).
func (s *Service) CreateAttachment(ctx context.Context, caseID, fileName, contentType, uploadedByUserID string, content []byte, requestID, correlationID, actorName string) (*domain.Attachment, error) {
	start := time.Now()
	method := "CreateAttachment"
	logs := []string{"starting CreateAttachment"}
	var err error
	defer func() {
		if err != nil {
			slog.Error("CreateAttachment failed", "layer", currentLayer, "method", method, "error", err, "logs", logs, "duration_ms", time.Since(start).Milliseconds())
		}
	}()

	if caseID == "" || uploadedByUserID == "" {
		err = ErrValidation{Field: "case_id or uploaded_by_user_id", Msg: "required"}
		return nil, err
	}
	if fileName == "" {
		err = ErrValidation{Field: "file_name", Msg: "required"}
		return nil, err
	}
	if int64(len(content)) > s.MaxBytes {
		err = ErrValidation{Field: "content", Msg: "file size exceeds limit"}
		return nil, err
	}
	if !s.contentTypeAllowed(contentType) {
		err = ErrValidation{Field: "content_type", Msg: "not allowed"}
		return nil, err
	}
	safeName := sanitizeFileName(fileName)
	if safeName == "" {
		safeName = "attachment"
	}
	logs = append(logs, "validated")

	attachmentID := uuid.New().String()
	now := time.Now().UTC()
	bucket := s.getBucket()
	storageKey := storage.KeyGenerator(caseID, attachmentID, safeName)

	// 1. Upload to storage first (transactional order: storage then DB; on DB failure we compensate by deleting object)
	if err = s.Storage.Put(ctx, storageKey, bucket, content); err != nil {
		return nil, err
	}
	logs = append(logs, "storage put done")

	a := &domain.Attachment{
		ID:               attachmentID,
		CaseID:           caseID,
		FileName:         safeName,
		FileSizeBytes:    int64(len(content)),
		ContentType:      contentType,
		StorageProvider:  "s3",
		StorageKey:       storageKey,
		StorageBucket:    bucket,
		UploadedByUserID: uploadedByUserID,
		UploadedAt:       now,
		IsDeleted:       false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err = s.Repo.CreateAttachment(ctx, a); err != nil {
		// Compensation: remove object so we don't leave orphaned blobs
		_ = s.Storage.Delete(ctx, storageKey, bucket)
		return nil, err
	}
	logs = append(logs, "db insert done")

	evt := audit.Event{
		EventID:       uuid.New().String(),
		EventType:     "attachment.uploaded",
		SourceService: "attachment-service",
		EntityType:    "attachment",
		EntityID:      attachmentID,
		CaseID:        caseID,
		Action:        "create",
		ActorUserID:   uploadedByUserID,
		ActorName:     actorName,
		RequestID:     requestID,
		CorrelationID: correlationID,
		AfterData:     map[string]interface{}{"file_name": safeName, "file_size_bytes": a.FileSizeBytes},
		OccurredAt:    now.Format(time.RFC3339),
	}
	_ = s.Audit.Publish(context.Background(), evt)
	return a, nil
}

// ListAttachmentsByCase returns attachments for a case (excluding deleted by default).
func (s *Service) ListAttachmentsByCase(ctx context.Context, caseID string, pageSize int32, pageToken string, includeDeleted bool) ([]*domain.Attachment, string, error) {
	return s.Repo.ListByCaseID(ctx, caseID, pageSize, pageToken, includeDeleted)
}

// DeleteAttachment soft-deletes metadata and deletes the object from storage, then publishes audit.
func (s *Service) DeleteAttachment(ctx context.Context, attachmentID, requestID, correlationID, actorUserID, actorName string) error {
	start := time.Now()
	method := "DeleteAttachment"
	logs := []string{"starting DeleteAttachment"}
	var err error
	defer func() {
		if err != nil {
			slog.Error("DeleteAttachment failed", "layer", currentLayer, "method", method, "error", err, "logs", logs, "duration_ms", time.Since(start).Milliseconds())
		}
	}()

	a, err := s.Repo.GetByID(ctx, attachmentID)
	if err != nil {
		return err
	}
	if a == nil {
		return ErrNotFound
	}
	if a.IsDeleted {
		return ErrNotFound
	}
	logs = append(logs, "found attachment")

	now := time.Now().UTC()
	if err = s.Repo.SoftDelete(ctx, attachmentID, now); err != nil {
		return err
	}
	logs = append(logs, "soft delete done")

	bucket := a.StorageBucket
	if bucket == "" {
		bucket = s.getBucket()
	}
	if err = s.Storage.Delete(ctx, a.StorageKey, bucket); err != nil && err != storage.ErrNotFound {
		// Log but don't fail; metadata is already soft-deleted
		slog.Warn("storage delete failed after soft delete", "attachment_id", attachmentID, "error", err)
	}
	logs = append(logs, "storage delete done")

	evt := audit.Event{
		EventID:       uuid.New().String(),
		EventType:     "attachment.deleted",
		SourceService: "attachment-service",
		EntityType:    "attachment",
		EntityID:      attachmentID,
		CaseID:        a.CaseID,
		Action:        "delete",
		ActorUserID:   actorUserID,
		ActorName:     actorName,
		RequestID:     requestID,
		CorrelationID: correlationID,
		OccurredAt:    now.Format(time.RFC3339),
	}
	_ = s.Audit.Publish(context.Background(), evt)
	return nil
}

// GetAttachment returns a single attachment by ID.
func (s *Service) GetAttachment(ctx context.Context, attachmentID string) (*domain.Attachment, error) {
	a, err := s.Repo.GetByID(ctx, attachmentID)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, ErrNotFound
	}
	return a, nil
}

func (s *Service) getBucket() string {
	if s.Bucket != "" {
		return s.Bucket
	}
	return "attachments"
}

func (s *Service) contentTypeAllowed(ct string) bool {
	if len(s.Denied) > 0 {
		lower := strings.ToLower(strings.TrimSpace(ct))
		for _, d := range s.Denied {
			if strings.HasPrefix(lower, strings.ToLower(d)) {
				return false
			}
		}
	}
	if len(s.Allowed) == 0 {
		return true
	}
	lower := strings.ToLower(strings.TrimSpace(ct))
	for _, a := range s.Allowed {
		if strings.HasPrefix(lower, strings.ToLower(a)) {
			return true
		}
	}
	return false
}

// sanitizeFileName removes path traversal and keeps a safe base name.
func sanitizeFileName(name string) string {
	base := path.Base(name)
	base = strings.TrimSpace(base)
	if base == "" || base == "." {
		return "attachment"
	}
	// Remove any remaining path-like or dangerous chars
	base = strings.ReplaceAll(base, "..", "")
	base = strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == 0 {
			return -1
		}
		return r
	}, base)
	if base == "" {
		return "attachment"
	}
	return base
}

package repository

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/soc-platform/attachment-service/domain"
)

const currentLayer = "repository"

// Repository handles persistence for attachments (metadata only).
type Repository struct {
	Pool *pgxpool.Pool
}

// New creates a Repository with the given pool.
func New(pool *pgxpool.Pool) *Repository {
	return &Repository{Pool: pool}
}

// CreateAttachment inserts an attachment row.
func (r *Repository) CreateAttachment(ctx context.Context, a *domain.Attachment) error {
	_, err := r.Pool.Exec(ctx, `
		INSERT INTO attachments (
			id, case_id, file_name, file_size_bytes, content_type,
			storage_provider, storage_key, storage_bucket,
			uploaded_by_user_id, uploaded_at, is_deleted, deleted_at,
			metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		a.ID, a.CaseID, a.FileName, a.FileSizeBytes, nullStr(a.ContentType),
		a.StorageProvider, a.StorageKey, nullStr(a.StorageBucket),
		a.UploadedByUserID, a.UploadedAt, a.IsDeleted, nullTimePtr(a.DeletedAt),
		jsonBOrNull(a.MetadataJSON), a.CreatedAt, a.UpdatedAt,
	)
	return err
}

// GetByID returns an attachment by ID, or nil if not found.
func (r *Repository) GetByID(ctx context.Context, id string) (*domain.Attachment, error) {
	var a domain.Attachment
	var contentType, storageBucket, metadataJSON *string
	var deletedAt *time.Time
	err := r.Pool.QueryRow(ctx, `
		SELECT id, case_id, file_name, file_size_bytes, content_type,
			storage_provider, storage_key, storage_bucket,
			uploaded_by_user_id, uploaded_at, is_deleted, deleted_at,
			metadata::text, created_at, updated_at
		FROM attachments WHERE id = $1`,
		id,
	).Scan(
		&a.ID, &a.CaseID, &a.FileName, &a.FileSizeBytes, &contentType,
		&a.StorageProvider, &a.StorageKey, &storageBucket,
		&a.UploadedByUserID, &a.UploadedAt, &a.IsDeleted, &deletedAt,
		&metadataJSON, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if contentType != nil {
		a.ContentType = *contentType
	}
	if storageBucket != nil {
		a.StorageBucket = *storageBucket
	}
	if metadataJSON != nil {
		a.MetadataJSON = *metadataJSON
	}
	a.DeletedAt = deletedAt
	return &a, nil
}

// ListByCaseID returns attachments for a case with optional pagination. By default excludes deleted.
func (r *Repository) ListByCaseID(ctx context.Context, caseID string, pageSize int32, pageToken string, includeDeleted bool) ([]*domain.Attachment, string, error) {
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := int64(0)
	if pageToken != "" {
		if v, err := strconv.ParseInt(pageToken, 10, 64); err == nil {
			offset = v
		}
	}
	cond := "case_id = $1"
	args := []interface{}{caseID}
	if !includeDeleted {
		cond += " AND is_deleted = false"
	}
	args = append(args, pageSize+1, offset)
	rows, err := r.Pool.Query(ctx, `
		SELECT id, case_id, file_name, file_size_bytes, content_type,
			storage_provider, storage_key, storage_bucket,
			uploaded_by_user_id, uploaded_at, is_deleted, deleted_at,
			metadata::text, created_at, updated_at
		FROM attachments WHERE `+cond+`
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3`,
		args...,
	)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()
	var list []*domain.Attachment
	var n int32
	for rows.Next() {
		n++
		if n > pageSize {
			nextToken := strconv.FormatInt(offset+int64(pageSize), 10)
			return list, nextToken, nil
		}
		var a domain.Attachment
		var contentType, storageBucket, metadataJSON *string
		var deletedAt *time.Time
		err := rows.Scan(
			&a.ID, &a.CaseID, &a.FileName, &a.FileSizeBytes, &contentType,
			&a.StorageProvider, &a.StorageKey, &storageBucket,
			&a.UploadedByUserID, &a.UploadedAt, &a.IsDeleted, &deletedAt,
			&metadataJSON, &a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, "", err
		}
		if contentType != nil {
			a.ContentType = *contentType
		}
		if storageBucket != nil {
			a.StorageBucket = *storageBucket
		}
		if metadataJSON != nil {
			a.MetadataJSON = *metadataJSON
		}
		a.DeletedAt = deletedAt
		list = append(list, &a)
	}
	return list, "", rows.Err()
}

// SoftDelete marks attachment as deleted and sets deleted_at.
func (r *Repository) SoftDelete(ctx context.Context, id string, deletedAt time.Time) error {
	_, err := r.Pool.Exec(ctx, `
		UPDATE attachments SET is_deleted = true, deleted_at = $2, updated_at = $2 WHERE id = $1`,
		id, deletedAt,
	)
	return err
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nullTimePtr(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}

func jsonBOrNull(s string) interface{} {
	if s == "" {
		return nil
	}
	var j json.RawMessage
	if json.Unmarshal([]byte(s), &j) != nil {
		return nil
	}
	return j
}

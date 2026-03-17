package repository

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

// UpsertEnrichmentResult inserts or updates by natural key (dedupe index).
func (r *Repository) UpsertEnrichmentResult(ctx context.Context, e *EnrichmentResult) error {
	query := `
INSERT INTO enrichment_results (
  id, case_id, observable_id, enrichment_type, source_name, source_record_id,
  status, summary, result_data, score, confidence, requested_at, received_at, expires_at,
  last_updated_by, created_at, updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
)
ON CONFLICT (
  COALESCE(case_id, '00000000-0000-0000-0000-000000000000'::uuid),
  COALESCE(observable_id, '00000000-0000-0000-0000-000000000000'::uuid),
  enrichment_type,
  source_name,
  COALESCE(source_record_id, '')
)
DO UPDATE SET
  id = EXCLUDED.id,
  status = EXCLUDED.status,
  summary = EXCLUDED.summary,
  result_data = EXCLUDED.result_data,
  score = EXCLUDED.score,
  confidence = EXCLUDED.confidence,
  requested_at = EXCLUDED.requested_at,
  received_at = EXCLUDED.received_at,
  expires_at = EXCLUDED.expires_at,
  last_updated_by = EXCLUDED.last_updated_by,
  updated_at = EXCLUDED.updated_at
`
	now := time.Now().UTC()
	_, err := r.Pool.Exec(ctx, query,
		e.ID, e.CaseID, e.ObservableID, e.EnrichmentType, e.SourceName, e.SourceRecordID,
		e.Status, e.Summary, e.ResultData, e.Score, e.Confidence,
		e.RequestedAt, e.ReceivedAt, e.ExpiresAt, e.LastUpdatedBy,
		now, now,
	)
	return err
}

// GetEnrichmentResultByID returns one row by id, or nil if not found.
func (r *Repository) GetEnrichmentResultByID(ctx context.Context, id string) (*EnrichmentResult, error) {
	query := `
SELECT id, case_id, observable_id, enrichment_type, source_name, source_record_id,
  status, summary, result_data, score, confidence, requested_at, received_at, expires_at,
  last_updated_by, created_at, updated_at
FROM enrichment_results WHERE id = $1
`
	var e EnrichmentResult
	err := r.Pool.QueryRow(ctx, query, id).Scan(
		&e.ID, &e.CaseID, &e.ObservableID, &e.EnrichmentType, &e.SourceName, &e.SourceRecordID,
		&e.Status, &e.Summary, &e.ResultData, &e.Score, &e.Confidence,
		&e.RequestedAt, &e.ReceivedAt, &e.ExpiresAt, &e.LastUpdatedBy,
		&e.CreatedAt, &e.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// ListEnrichmentResultsByCase returns results for a case, newest first, with optional filters.
func (r *Repository) ListEnrichmentResultsByCase(ctx context.Context, caseID string, f ListFilter) ([]*EnrichmentResult, string, error) {
	return r.listEnrichment(ctx, &caseID, nil, f)
}

// ListEnrichmentResultsByObservable returns results for an observable, newest first.
func (r *Repository) ListEnrichmentResultsByObservable(ctx context.Context, observableID string, f ListFilter) ([]*EnrichmentResult, string, error) {
	return r.listEnrichment(ctx, nil, &observableID, f)
}

func (r *Repository) listEnrichment(ctx context.Context, caseID, observableID *string, f ListFilter) ([]*EnrichmentResult, string, error) {
	pageSize := f.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := 0
	if f.PageToken != "" {
		if o, err := strconv.Atoi(f.PageToken); err == nil && o >= 0 {
			offset = o
		}
	}

	query := `
SELECT id, case_id, observable_id, enrichment_type, source_name, source_record_id,
  status, summary, result_data, score, confidence, requested_at, received_at, expires_at,
  last_updated_by, created_at, updated_at
FROM enrichment_results
WHERE 1=1
`
	args := []interface{}{}
	argNum := 1
	if caseID != nil {
		query += ` AND case_id = $` + strconv.Itoa(argNum)
		args = append(args, *caseID)
		argNum++
	}
	if observableID != nil {
		query += ` AND observable_id = $` + strconv.Itoa(argNum)
		args = append(args, *observableID)
		argNum++
	}
	if f.SourceName != nil {
		query += ` AND source_name = $` + strconv.Itoa(argNum)
		args = append(args, *f.SourceName)
		argNum++
	}
	if f.Type != nil {
		query += ` AND enrichment_type = $` + strconv.Itoa(argNum)
		args = append(args, *f.Type)
		argNum++
	}
	if f.ActiveOnly {
		query += ` AND (expires_at IS NULL OR expires_at > NOW())`
	}
	query += ` ORDER BY received_at DESC LIMIT $` + strconv.Itoa(argNum) + ` OFFSET $` + strconv.Itoa(argNum+1)
	args = append(args, pageSize+1, offset)

	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var items []*EnrichmentResult
	for rows.Next() {
		var e EnrichmentResult
		err := rows.Scan(
			&e.ID, &e.CaseID, &e.ObservableID, &e.EnrichmentType, &e.SourceName, &e.SourceRecordID,
			&e.Status, &e.Summary, &e.ResultData, &e.Score, &e.Confidence,
			&e.RequestedAt, &e.ReceivedAt, &e.ExpiresAt, &e.LastUpdatedBy,
			&e.CreatedAt, &e.UpdatedAt,
		)
		if err != nil {
			return nil, "", err
		}
		items = append(items, &e)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	var nextPageToken string
	if len(items) > int(pageSize) {
		items = items[:pageSize]
		nextPageToken = strconv.Itoa(offset + int(pageSize))
	}
	return items, nextPageToken, nil
}

// Ensure EnrichmentResult.ResultData is valid JSON for storage (already json.RawMessage).
var _ = json.Marshal

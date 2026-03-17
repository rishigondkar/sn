package repository

import (
	"context"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

// UpsertThreatLookupResult inserts or updates by natural key (dedupe index).
func (r *Repository) UpsertThreatLookupResult(ctx context.Context, t *ThreatLookupResult) error {
	query := `
INSERT INTO threat_lookup_results (
  id, case_id, observable_id, lookup_type, source_name, source_record_id,
  verdict, risk_score, confidence_score, tags, matched_indicators, summary,
  result_data, queried_at, received_at, expires_at, created_at, updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
)
ON CONFLICT (
  observable_id,
  lookup_type,
  source_name,
  COALESCE(source_record_id, '')
)
DO UPDATE SET
  id = EXCLUDED.id,
  case_id = EXCLUDED.case_id,
  verdict = EXCLUDED.verdict,
  risk_score = EXCLUDED.risk_score,
  confidence_score = EXCLUDED.confidence_score,
  tags = EXCLUDED.tags,
  matched_indicators = EXCLUDED.matched_indicators,
  summary = EXCLUDED.summary,
  result_data = EXCLUDED.result_data,
  queried_at = EXCLUDED.queried_at,
  received_at = EXCLUDED.received_at,
  expires_at = EXCLUDED.expires_at,
  updated_at = EXCLUDED.updated_at
`
	now := time.Now().UTC()
	_, err := r.Pool.Exec(ctx, query,
		t.ID, t.CaseID, t.ObservableID, t.LookupType, t.SourceName, t.SourceRecordID,
		t.Verdict, t.RiskScore, t.ConfidenceScore, t.Tags, t.MatchedIndicators, t.Summary,
		t.ResultData, t.QueriedAt, t.ReceivedAt, t.ExpiresAt,
		now, now,
	)
	return err
}

// GetThreatLookupResultByID returns one row by id, or nil if not found.
func (r *Repository) GetThreatLookupResultByID(ctx context.Context, id string) (*ThreatLookupResult, error) {
	query := `
SELECT id, case_id, observable_id, lookup_type, source_name, source_record_id,
  verdict, risk_score, confidence_score, tags, matched_indicators, summary,
  result_data, queried_at, received_at, expires_at, created_at, updated_at
FROM threat_lookup_results WHERE id = $1
`
	var t ThreatLookupResult
	err := r.Pool.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.CaseID, &t.ObservableID, &t.LookupType, &t.SourceName, &t.SourceRecordID,
		&t.Verdict, &t.RiskScore, &t.ConfidenceScore, &t.Tags, &t.MatchedIndicators, &t.Summary,
		&t.ResultData, &t.QueriedAt, &t.ReceivedAt, &t.ExpiresAt,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// ListThreatLookupResultsByCase returns results for a case, newest first.
func (r *Repository) ListThreatLookupResultsByCase(ctx context.Context, caseID string, f ListFilter) ([]*ThreatLookupResult, string, error) {
	return r.listThreatLookup(ctx, &caseID, nil, f)
}

// ListThreatLookupResultsByObservable returns results for an observable, newest first.
func (r *Repository) ListThreatLookupResultsByObservable(ctx context.Context, observableID string, f ListFilter) ([]*ThreatLookupResult, string, error) {
	return r.listThreatLookup(ctx, nil, &observableID, f)
}

func (r *Repository) listThreatLookup(ctx context.Context, caseID, observableID *string, f ListFilter) ([]*ThreatLookupResult, string, error) {
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
SELECT id, case_id, observable_id, lookup_type, source_name, source_record_id,
  verdict, risk_score, confidence_score, tags, matched_indicators, summary,
  result_data, queried_at, received_at, expires_at, created_at, updated_at
FROM threat_lookup_results
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
		query += ` AND lookup_type = $` + strconv.Itoa(argNum)
		args = append(args, *f.Type)
		argNum++
	}
	if f.Verdict != nil {
		query += ` AND verdict = $` + strconv.Itoa(argNum)
		args = append(args, *f.Verdict)
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

	var items []*ThreatLookupResult
	for rows.Next() {
		var t ThreatLookupResult
		err := rows.Scan(
			&t.ID, &t.CaseID, &t.ObservableID, &t.LookupType, &t.SourceName, &t.SourceRecordID,
			&t.Verdict, &t.RiskScore, &t.ConfidenceScore, &t.Tags, &t.MatchedIndicators, &t.Summary,
			&t.ResultData, &t.QueriedAt, &t.ReceivedAt, &t.ExpiresAt,
			&t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, "", err
		}
		items = append(items, &t)
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

// GetThreatLookupSummaryByObservable returns aggregated summary for an observable.
func (r *Repository) GetThreatLookupSummaryByObservable(ctx context.Context, observableID string) (*ThreatSummary, error) {
	query := `
SELECT
  observable_id,
  COUNT(*)::int AS total_count,
  MAX(verdict) AS highest_verdict,
  MAX(risk_score) AS max_risk_score,
  array_agg(DISTINCT source_name) AS source_names,
  MAX(received_at) AS latest_received_at
FROM threat_lookup_results
WHERE observable_id = $1
GROUP BY observable_id
`
	var ts ThreatSummary
	var sourceNames []string
	err := r.Pool.QueryRow(ctx, query, observableID).Scan(
		&ts.ObservableID,
		&ts.TotalCount,
		&ts.HighestVerdict,
		&ts.MaxRiskScore,
		&sourceNames,
		&ts.LatestReceivedAt,
	)
	if err == pgx.ErrNoRows {
		return &ThreatSummary{ObservableID: observableID, TotalCount: 0, LatestReceivedAt: time.Time{}}, nil
	}
	if err != nil {
		return nil, err
	}
	ts.SourceNames = sourceNames
	return &ts, nil
}

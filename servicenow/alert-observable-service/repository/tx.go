package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

// Tx is a transaction-scoped repository for use inside RunInTx.
type Tx struct {
	tx *sql.Tx
}

// RunInTx runs fn inside a transaction. Commit is called if fn returns nil; otherwise Rollback.
func (p *Postgres) RunInTx(ctx context.Context, fn func(ctx context.Context, tx *Tx) error) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	t := &Tx{tx: tx}
	err = fn(ctx, t)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (t *Tx) CreateObservable(ctx context.Context, o *Observable) error {
	_, err := t.tx.ExecContext(ctx, `INSERT INTO observables (id, observable_type, observable_value, normalized_value, first_seen_time, last_seen_time, created_at, updated_at, incident_count, finding, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		o.ID, o.ObservableType, o.ObservableValue, nullStr(o.NormalizedValue), nullTime(o.FirstSeenTime), nullTime(o.LastSeenTime), o.CreatedAt, o.UpdatedAt, o.IncidentCount, nullStr(o.Finding), nullStr(o.Notes))
	return err
}

func (t *Tx) GetObservableByTypeAndNormalized(ctx context.Context, observableType, normalizedValue string) (*Observable, error) {
	var o Observable
	err := t.tx.QueryRowContext(ctx, `SELECT id, observable_type, observable_value, normalized_value, first_seen_time, last_seen_time, created_at, updated_at, COALESCE(incident_count, 0), finding, notes
		FROM observables WHERE observable_type = $1 AND normalized_value = $2`, observableType, normalizedValue).Scan(
		&o.ID, &o.ObservableType, &o.ObservableValue, &o.NormalizedValue, &o.FirstSeenTime, &o.LastSeenTime, &o.CreatedAt, &o.UpdatedAt, &o.IncidentCount, &o.Finding, &o.Notes)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (t *Tx) GetObservableByID(ctx context.Context, id uuid.UUID) (*Observable, error) {
	var o Observable
	var normVal, finding, notes sql.NullString
	var firstSeen, lastSeen sql.NullTime
	err := t.tx.QueryRowContext(ctx, `SELECT id, observable_type, observable_value, normalized_value, first_seen_time, last_seen_time, created_at, updated_at, COALESCE(incident_count, 0), finding, notes
		FROM observables WHERE id = $1`, id).Scan(
		&o.ID, &o.ObservableType, &o.ObservableValue, &normVal, &firstSeen, &lastSeen, &o.CreatedAt, &o.UpdatedAt, &o.IncidentCount, &finding, &notes)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if normVal.Valid {
		o.NormalizedValue = &normVal.String
	}
	if firstSeen.Valid {
		o.FirstSeenTime = &firstSeen.Time
	}
	if lastSeen.Valid {
		o.LastSeenTime = &lastSeen.Time
	}
	if finding.Valid {
		o.Finding = &finding.String
	}
	if notes.Valid {
		o.Notes = &notes.String
	}
	return &o, nil
}

func (t *Tx) CreateCaseObservable(ctx context.Context, co *CaseObservable) error {
	_, err := t.tx.ExecContext(ctx, `INSERT INTO case_observables (id, case_id, observable_id, role_in_case, tracking_status, is_primary, accuracy, determination, impact, added_by_user_id, added_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		co.ID, co.CaseID, co.ObservableID, nullStr(co.RoleInCase), nullStr(co.TrackingStatus), co.IsPrimary,
		nullStr(co.Accuracy), nullStr(co.Determination), nullStr(co.Impact), nullUUID(co.AddedByUserID), co.AddedAt, co.UpdatedAt)
	return err
}

func (t *Tx) GetCaseObservableByCaseAndObservable(ctx context.Context, caseID, observableID uuid.UUID) (*CaseObservable, error) {
	var co CaseObservable
	err := t.tx.QueryRowContext(ctx, `SELECT id, case_id, observable_id, role_in_case, tracking_status, is_primary, accuracy, determination, impact, added_by_user_id, added_at, updated_at
		FROM case_observables WHERE case_id = $1 AND observable_id = $2`, caseID, observableID).Scan(
		&co.ID, &co.CaseID, &co.ObservableID, &co.RoleInCase, &co.TrackingStatus, &co.IsPrimary,
		&co.Accuracy, &co.Determination, &co.Impact, &co.AddedByUserID, &co.AddedAt, &co.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &co, nil
}

func (t *Tx) DeleteCaseObservableByCaseAndObservable(ctx context.Context, caseID, observableID uuid.UUID) error {
	_, err := t.tx.ExecContext(ctx, `DELETE FROM case_observables WHERE case_id = $1 AND observable_id = $2`, caseID, observableID)
	return err
}

func (t *Tx) GetObservableIDsByCaseID(ctx context.Context, caseID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := t.tx.QueryContext(ctx, `SELECT observable_id FROM case_observables WHERE case_id = $1`, caseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// AllCaseIDsByObservable returns all case IDs linked to this observable (for similar-incident computation).
func (t *Tx) AllCaseIDsByObservable(ctx context.Context, observableID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := t.tx.QueryContext(ctx, `SELECT case_id FROM case_observables WHERE observable_id = $1`, observableID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (t *Tx) UpsertSimilarIncident(ctx context.Context, s *SimilarIncident) error {
	_, err := t.tx.ExecContext(ctx, `INSERT INTO similar_security_incidents (id, case_id, similar_case_id, match_reason, shared_observable_count, shared_observable_ids, shared_observable_values, similarity_score, last_computed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (case_id, similar_case_id) DO UPDATE SET
			match_reason = EXCLUDED.match_reason,
			shared_observable_count = EXCLUDED.shared_observable_count,
			shared_observable_ids = EXCLUDED.shared_observable_ids,
			shared_observable_values = EXCLUDED.shared_observable_values,
			similarity_score = EXCLUDED.similarity_score,
			last_computed_at = EXCLUDED.last_computed_at,
			updated_at = EXCLUDED.updated_at`,
		s.ID, s.CaseID, s.SimilarCaseID, s.MatchReason, s.SharedObservableCount, s.SharedObservableIDs, nullJSONB(s.SharedObservableValues), nullFloat(s.SimilarityScore), s.LastComputedAt, s.CreatedAt, s.UpdatedAt)
	return err
}

func (t *Tx) DeleteSimilarIncidentsByCaseID(ctx context.Context, caseID uuid.UUID) error {
	_, err := t.tx.ExecContext(ctx, `DELETE FROM similar_security_incidents WHERE case_id = $1 OR similar_case_id = $1`, caseID)
	return err
}

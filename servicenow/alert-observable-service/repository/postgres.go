package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Postgres implements all repository interfaces.
type Postgres struct {
	db *sql.DB
}

// NewPostgres returns a new Postgres repository.
func NewPostgres(db *sql.DB) *Postgres {
	return &Postgres{db: db}
}

// AlertRule
func (p *Postgres) CreateAlertRule(ctx context.Context, r *AlertRule) error {
	_, err := p.db.ExecContext(ctx, `INSERT INTO alert_rules (id, rule_name, rule_type, source_system, external_rule_key, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		r.ID, r.RuleName, nullStr(r.RuleType), nullStr(r.SourceSystem), nullStr(r.ExternalRuleKey), nullStr(r.Description), r.IsActive, r.CreatedAt, r.UpdatedAt)
	return err
}

func (p *Postgres) GetAlertRuleByID(ctx context.Context, id uuid.UUID) (*AlertRule, error) {
	var r AlertRule
	err := p.db.QueryRowContext(ctx, `SELECT id, rule_name, rule_type, source_system, external_rule_key, description, is_active, created_at, updated_at
		FROM alert_rules WHERE id = $1`, id).Scan(
		&r.ID, &r.RuleName, &r.RuleType, &r.SourceSystem, &r.ExternalRuleKey, &r.Description, &r.IsActive, &r.CreatedAt, &r.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// Alert
func (p *Postgres) CreateAlert(ctx context.Context, a *Alert) error {
	_, err := p.db.ExecContext(ctx, `INSERT INTO alerts (id, case_id, alert_rule_id, source_system, source_alert_id, title, description, event_occurred_time, event_received_time, severity, raw_payload, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		a.ID, a.CaseID, nullUUID(a.AlertRuleID), a.SourceSystem, nullStr(a.SourceAlertID), nullStr(a.Title), nullStr(a.Description),
		nullTime(a.EventOccurredTime), nullTime(a.EventReceivedTime), nullStr(a.Severity), nullJSONB(a.RawPayload), a.CreatedAt, a.UpdatedAt)
	return err
}

func (p *Postgres) GetAlertByID(ctx context.Context, id uuid.UUID) (*Alert, error) {
	var a Alert
	err := p.db.QueryRowContext(ctx, `SELECT id, case_id, alert_rule_id, source_system, source_alert_id, title, description, event_occurred_time, event_received_time, severity, raw_payload, created_at, updated_at
		FROM alerts WHERE id = $1`, id).Scan(
		&a.ID, &a.CaseID, &a.AlertRuleID, &a.SourceSystem, &a.SourceAlertID, &a.Title, &a.Description,
		&a.EventOccurredTime, &a.EventReceivedTime, &a.Severity, &a.RawPayload, &a.CreatedAt, &a.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (p *Postgres) ListAlertsByCaseID(ctx context.Context, caseID uuid.UUID, opts ListOpts) ([]*Alert, string, error) {
	limit := opts.PageSize
	if limit <= 0 {
		limit = 50
	}
	offset := 0
	if opts.PageToken != "" {
		if o, err := strconv.Atoi(opts.PageToken); err == nil {
			offset = o
		}
	}
	rows, err := p.db.QueryContext(ctx, `SELECT id, case_id, alert_rule_id, source_system, source_alert_id, title, description, event_occurred_time, event_received_time, severity, raw_payload, created_at, updated_at
		FROM alerts WHERE case_id = $1 ORDER BY created_at LIMIT $2 OFFSET $3`, caseID, limit+1, offset)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()
	var list []*Alert
	for rows.Next() {
		var a Alert
		if err := rows.Scan(&a.ID, &a.CaseID, &a.AlertRuleID, &a.SourceSystem, &a.SourceAlertID, &a.Title, &a.Description,
			&a.EventOccurredTime, &a.EventReceivedTime, &a.Severity, &a.RawPayload, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, "", err
		}
		list = append(list, &a)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}
	nextToken := ""
	if len(list) > int(limit) {
		list = list[:limit]
		nextToken = strconv.Itoa(offset + int(limit))
	}
	return list, nextToken, nil
}

// Observable
func (p *Postgres) CreateObservable(ctx context.Context, o *Observable) error {
	_, err := p.db.ExecContext(ctx, `INSERT INTO observables (id, observable_type, observable_value, normalized_value, first_seen_time, last_seen_time, created_at, updated_at, incident_count, finding, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		o.ID, o.ObservableType, o.ObservableValue, nullStr(o.NormalizedValue), nullTime(o.FirstSeenTime), nullTime(o.LastSeenTime), o.CreatedAt, o.UpdatedAt,
		o.IncidentCount, nullStr(o.Finding), nullStr(o.Notes))
	return err
}

func (p *Postgres) GetObservableByID(ctx context.Context, id uuid.UUID) (*Observable, error) {
	var o Observable
	var normVal, finding, notes sql.NullString
	var firstSeen, lastSeen sql.NullTime
	err := p.db.QueryRowContext(ctx, `SELECT id, observable_type, observable_value, normalized_value, first_seen_time, last_seen_time, created_at, updated_at, COALESCE(incident_count, 0), finding, notes
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

func (p *Postgres) GetObservableByTypeAndNormalized(ctx context.Context, observableType, normalizedValue string) (*Observable, error) {
	var o Observable
	var normVal, finding, notes sql.NullString
	var firstSeen, lastSeen sql.NullTime
	err := p.db.QueryRowContext(ctx, `SELECT id, observable_type, observable_value, normalized_value, first_seen_time, last_seen_time, created_at, updated_at, COALESCE(incident_count, 0), finding, notes
		FROM observables WHERE observable_type = $1 AND normalized_value = $2`, observableType, normalizedValue).Scan(
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

func (p *Postgres) UpdateObservable(ctx context.Context, o *Observable) error {
	_, err := p.db.ExecContext(ctx, `UPDATE observables SET observable_type = $2, observable_value = $3, normalized_value = $4, finding = $5, notes = $6, updated_at = $7
		WHERE id = $1`,
		o.ID, o.ObservableType, o.ObservableValue, nullStr(o.NormalizedValue), nullStr(o.Finding), nullStr(o.Notes), o.UpdatedAt)
	return err
}

func (p *Postgres) ListObservables(ctx context.Context, opts ListOpts, searchQuery *string) ([]*Observable, string, error) {
	limit := opts.PageSize
	if limit <= 0 {
		limit = 50
	}
	offset := 0
	if opts.PageToken != "" {
		if o, err := strconv.Atoi(opts.PageToken); err == nil {
			offset = o
		}
	}
	query := `SELECT id, observable_type, observable_value, normalized_value, first_seen_time, last_seen_time, created_at, updated_at, COALESCE(incident_count, 0), finding, notes
		FROM observables`
	args := []interface{}{}
	argNum := 1
	if searchQuery != nil && *searchQuery != "" {
		query += ` WHERE (observable_value ILIKE $` + strconv.Itoa(argNum) + ` OR normalized_value ILIKE $` + strconv.Itoa(argNum) + `)`
		args = append(args, "%"+*searchQuery+"%")
		argNum++
	}
	query += ` ORDER BY created_at DESC LIMIT $` + strconv.Itoa(argNum) + ` OFFSET $` + strconv.Itoa(argNum+1)
	args = append(args, limit+1, offset)
	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()
	var list []*Observable
	for rows.Next() {
		var o Observable
		var normVal, finding, notes sql.NullString
		var firstSeen, lastSeen sql.NullTime
		if err := rows.Scan(&o.ID, &o.ObservableType, &o.ObservableValue, &normVal, &firstSeen, &lastSeen, &o.CreatedAt, &o.UpdatedAt, &o.IncidentCount, &finding, &notes); err != nil {
			return nil, "", err
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
		list = append(list, &o)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}
	nextToken := ""
	if len(list) > int(limit) {
		list = list[:limit]
		nextToken = strconv.Itoa(offset + int(limit))
	}
	return list, nextToken, nil
}

// CaseObservable
func (p *Postgres) CreateCaseObservable(ctx context.Context, co *CaseObservable) error {
	_, err := p.db.ExecContext(ctx, `INSERT INTO case_observables (id, case_id, observable_id, role_in_case, tracking_status, is_primary, accuracy, determination, impact, added_by_user_id, added_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		co.ID, co.CaseID, co.ObservableID, nullStr(co.RoleInCase), nullStr(co.TrackingStatus), co.IsPrimary,
		nullStr(co.Accuracy), nullStr(co.Determination), nullStr(co.Impact), nullUUID(co.AddedByUserID), co.AddedAt, co.UpdatedAt)
	return err
}

func (p *Postgres) UpdateCaseObservable(ctx context.Context, co *CaseObservable) error {
	_, err := p.db.ExecContext(ctx, `UPDATE case_observables SET role_in_case=$2, tracking_status=$3, is_primary=$4, accuracy=$5, determination=$6, impact=$7, updated_at=$8 WHERE id=$1`,
		co.ID, nullStr(co.RoleInCase), nullStr(co.TrackingStatus), co.IsPrimary, nullStr(co.Accuracy), nullStr(co.Determination), nullStr(co.Impact), co.UpdatedAt)
	return err
}

func (p *Postgres) GetCaseObservableByID(ctx context.Context, id uuid.UUID) (*CaseObservable, error) {
	var co CaseObservable
	err := p.db.QueryRowContext(ctx, `SELECT id, case_id, observable_id, role_in_case, tracking_status, is_primary, accuracy, determination, impact, added_by_user_id, added_at, updated_at
		FROM case_observables WHERE id = $1`, id).Scan(
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

func (p *Postgres) GetCaseObservableByCaseAndObservable(ctx context.Context, caseID, observableID uuid.UUID) (*CaseObservable, error) {
	var co CaseObservable
	err := p.db.QueryRowContext(ctx, `SELECT id, case_id, observable_id, role_in_case, tracking_status, is_primary, accuracy, determination, impact, added_by_user_id, added_at, updated_at
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

func (p *Postgres) DeleteCaseObservableByCaseAndObservable(ctx context.Context, caseID, observableID uuid.UUID) error {
	_, err := p.db.ExecContext(ctx, `DELETE FROM case_observables WHERE case_id = $1 AND observable_id = $2`, caseID, observableID)
	return err
}

func (p *Postgres) ListCaseObservablesByCaseID(ctx context.Context, caseID uuid.UUID, opts ListOpts, filterType, filterStatus *string) ([]*CaseObservable, string, error) {
	limit := opts.PageSize
	if limit <= 0 {
		limit = 50
	}
	offset := 0
	if opts.PageToken != "" {
		if o, err := strconv.Atoi(opts.PageToken); err == nil {
			offset = o
		}
	}
	query := `SELECT id, case_id, observable_id, role_in_case, tracking_status, is_primary, accuracy, determination, impact, added_by_user_id, added_at, updated_at
		FROM case_observables WHERE case_id = $1`
	args := []interface{}{caseID}
	argNum := 2
	if filterType != nil && *filterType != "" {
		query += ` AND observable_id IN (SELECT id FROM observables WHERE observable_type = $` + strconv.Itoa(argNum) + `)`
		args = append(args, *filterType)
		argNum++
	}
	if filterStatus != nil && *filterStatus != "" {
		query += ` AND tracking_status = $` + strconv.Itoa(argNum)
		args = append(args, *filterStatus)
		argNum++
	}
	query += ` ORDER BY added_at LIMIT $` + strconv.Itoa(argNum) + ` OFFSET $` + strconv.Itoa(argNum+1)
	args = append(args, limit+1, offset)
	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()
	var list []*CaseObservable
	for rows.Next() {
		var co CaseObservable
		if err := rows.Scan(&co.ID, &co.CaseID, &co.ObservableID, &co.RoleInCase, &co.TrackingStatus, &co.IsPrimary,
			&co.Accuracy, &co.Determination, &co.Impact, &co.AddedByUserID, &co.AddedAt, &co.UpdatedAt); err != nil {
			return nil, "", err
		}
		list = append(list, &co)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}
	nextToken := ""
	if len(list) > int(limit) {
		list = list[:limit]
		nextToken = strconv.Itoa(offset + int(limit))
	}
	return list, nextToken, nil
}

// ListCaseObservablesWithDetailsByCaseID returns case observables with observable type/value via JOIN.
func (p *Postgres) ListCaseObservablesWithDetailsByCaseID(ctx context.Context, caseID uuid.UUID, opts ListOpts, filterType, filterStatus *string) ([]*CaseObservableWithDetails, string, error) {
	limit := opts.PageSize
	if limit <= 0 {
		limit = 50
	}
	offset := 0
	if opts.PageToken != "" {
		if o, err := strconv.Atoi(opts.PageToken); err == nil {
			offset = o
		}
	}
	query := `SELECT co.id, co.case_id, co.observable_id, co.role_in_case, co.tracking_status, co.is_primary, co.accuracy, co.determination, co.impact, co.added_by_user_id, co.added_at, co.updated_at,
		o.observable_type, o.observable_value, COALESCE(o.normalized_value, ''),
		(SELECT COUNT(*)::int FROM case_observables c2 WHERE c2.observable_id = co.observable_id) AS incident_count
		FROM case_observables co
		INNER JOIN observables o ON o.id = co.observable_id
		WHERE co.case_id = $1`
	args := []interface{}{caseID}
	argNum := 2
	if filterType != nil && *filterType != "" {
		query += ` AND o.observable_type = $` + strconv.Itoa(argNum)
		args = append(args, *filterType)
		argNum++
	}
	if filterStatus != nil && *filterStatus != "" {
		query += ` AND co.tracking_status = $` + strconv.Itoa(argNum)
		args = append(args, *filterStatus)
		argNum++
	}
	query += ` ORDER BY co.added_at LIMIT $` + strconv.Itoa(argNum) + ` OFFSET $` + strconv.Itoa(argNum+1)
	args = append(args, limit+1, offset)
	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()
	var list []*CaseObservableWithDetails
	for rows.Next() {
		var d CaseObservableWithDetails
		if err := rows.Scan(&d.ID, &d.CaseID, &d.ObservableID, &d.RoleInCase, &d.TrackingStatus, &d.IsPrimary,
			&d.Accuracy, &d.Determination, &d.Impact, &d.AddedByUserID, &d.AddedAt, &d.UpdatedAt,
			&d.ObservableType, &d.ObservableValue, &d.NormalizedValue, &d.IncidentCount); err != nil {
			return nil, "", err
		}
		row := d
		list = append(list, &row)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}
	nextToken := ""
	if len(list) > int(limit) {
		list = list[:limit]
		nextToken = strconv.Itoa(offset + int(limit))
	}
	return list, nextToken, nil
}

func (p *Postgres) ListCaseIDsByObservable(ctx context.Context, observableID uuid.UUID, opts ListOpts) ([]uuid.UUID, string, error) {
	limit := opts.PageSize
	if limit <= 0 {
		limit = 100
	}
	offset := 0
	if opts.PageToken != "" {
		if o, err := strconv.Atoi(opts.PageToken); err == nil {
			offset = o
		}
	}
	rows, err := p.db.QueryContext(ctx, `SELECT case_id FROM case_observables WHERE observable_id = $1 ORDER BY case_id LIMIT $2 OFFSET $3`, observableID, limit+1, offset)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, "", err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}
	nextToken := ""
	if len(ids) > int(limit) {
		ids = ids[:limit]
		nextToken = strconv.Itoa(offset + int(limit))
	}
	return ids, nextToken, nil
}

func (p *Postgres) GetObservableIDsByCaseID(ctx context.Context, caseID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := p.db.QueryContext(ctx, `SELECT observable_id FROM case_observables WHERE case_id = $1`, caseID)
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

// ChildObservable
func (p *Postgres) CreateChildObservable(ctx context.Context, c *ChildObservable) error {
	_, err := p.db.ExecContext(ctx, `INSERT INTO child_observables (id, parent_observable_id, child_observable_id, relationship_type, relationship_direction, confidence, source_name, source_record_id, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		c.ID, c.ParentObservableID, c.ChildObservableID, c.RelationshipType, nullStr(c.RelationshipDirection), nullFloat(c.Confidence),
		nullStr(c.SourceName), nullStr(c.SourceRecordID), nullJSONB(c.Metadata), c.CreatedAt, c.UpdatedAt)
	return err
}

func (p *Postgres) GetChildObservableByParentChildType(ctx context.Context, parentID, childID uuid.UUID, relationshipType string) (*ChildObservable, error) {
	var c ChildObservable
	err := p.db.QueryRowContext(ctx, `SELECT id, parent_observable_id, child_observable_id, relationship_type, relationship_direction, confidence, source_name, source_record_id, metadata, created_at, updated_at
		FROM child_observables WHERE parent_observable_id = $1 AND child_observable_id = $2 AND relationship_type = $3`,
		parentID, childID, relationshipType).Scan(
		&c.ID, &c.ParentObservableID, &c.ChildObservableID, &c.RelationshipType, &c.RelationshipDirection, &c.Confidence,
		&c.SourceName, &c.SourceRecordID, &c.Metadata, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (p *Postgres) ListChildObservablesByParentID(ctx context.Context, parentID uuid.UUID, opts ListOpts) ([]*ChildObservable, string, error) {
	limit := opts.PageSize
	if limit <= 0 {
		limit = 50
	}
	offset := 0
	if opts.PageToken != "" {
		if o, err := strconv.Atoi(opts.PageToken); err == nil {
			offset = o
		}
	}
	rows, err := p.db.QueryContext(ctx, `SELECT id, parent_observable_id, child_observable_id, relationship_type, relationship_direction, confidence, source_name, source_record_id, metadata, created_at, updated_at
		FROM child_observables WHERE parent_observable_id = $1 ORDER BY created_at LIMIT $2 OFFSET $3`, parentID, limit+1, offset)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()
	var list []*ChildObservable
	for rows.Next() {
		var c ChildObservable
		if err := rows.Scan(&c.ID, &c.ParentObservableID, &c.ChildObservableID, &c.RelationshipType, &c.RelationshipDirection, &c.Confidence,
			&c.SourceName, &c.SourceRecordID, &c.Metadata, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, "", err
		}
		list = append(list, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}
	nextToken := ""
	if len(list) > int(limit) {
		list = list[:limit]
		nextToken = strconv.Itoa(offset + int(limit))
	}
	return list, nextToken, nil
}

// SimilarIncident
func (p *Postgres) UpsertSimilarIncident(ctx context.Context, s *SimilarIncident) error {
	_, err := p.db.ExecContext(ctx, `INSERT INTO similar_security_incidents (id, case_id, similar_case_id, match_reason, shared_observable_count, shared_observable_ids, shared_observable_values, similarity_score, last_computed_at, created_at, updated_at)
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

func (p *Postgres) DeleteSimilarIncidentByCaseAndSimilarCase(ctx context.Context, caseID, similarCaseID uuid.UUID) error {
	_, err := p.db.ExecContext(ctx, `DELETE FROM similar_security_incidents WHERE case_id = $1 AND similar_case_id = $2`, caseID, similarCaseID)
	return err
}

func (p *Postgres) ListSimilarIncidentsByCaseID(ctx context.Context, caseID uuid.UUID, opts ListOpts) ([]*SimilarIncident, string, error) {
	limit := opts.PageSize
	if limit <= 0 {
		limit = 50
	}
	offset := 0
	if opts.PageToken != "" {
		if o, err := strconv.Atoi(opts.PageToken); err == nil {
			offset = o
		}
	}
	rows, err := p.db.QueryContext(ctx, `SELECT id, case_id, similar_case_id, match_reason, shared_observable_count, shared_observable_ids, shared_observable_values, similarity_score, last_computed_at, created_at, updated_at
		FROM similar_security_incidents WHERE case_id = $1 ORDER BY similarity_score DESC NULLS LAST, last_computed_at DESC LIMIT $2 OFFSET $3`, caseID, limit+1, offset)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()
	var list []*SimilarIncident
	for rows.Next() {
		var s SimilarIncident
		if err := rows.Scan(&s.ID, &s.CaseID, &s.SimilarCaseID, &s.MatchReason, &s.SharedObservableCount, &s.SharedObservableIDs, &s.SharedObservableValues, &s.SimilarityScore, &s.LastComputedAt, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, "", err
		}
		list = append(list, &s)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}
	nextToken := ""
	if len(list) > int(limit) {
		list = list[:limit]
		nextToken = strconv.Itoa(offset + int(limit))
	}
	return list, nextToken, nil
}

// Helpers
func nullStr(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}

func nullUUID(u *uuid.UUID) interface{} {
	if u == nil {
		return nil
	}
	return *u
}

func nullTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}

func nullFloat(f *float64) interface{} {
	if f == nil {
		return nil
	}
	return *f
}

func nullJSONB(b []byte) interface{} {
	if b == nil {
		return nil
	}
	return b
}

// Check duplicate key error (PostgreSQL unique violation).
func isUniqueViolation(err error) bool {
	if e, ok := err.(*pq.Error); ok {
		return e.Code == "23505"
	}
	return false
}

// JSONB array of UUID strings for similar_security_incidents.
func marshalUUIDArray(ids []uuid.UUID) ([]byte, error) {
	strs := make([]string, len(ids))
	for i, id := range ids {
		strs[i] = id.String()
	}
	return json.Marshal(strs)
}

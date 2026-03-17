package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/servicenow/case-service/domain"
)

const currentLayer = "repository"

// Repository performs data access for cases and worknotes.
type Repository struct {
	pool *pgxpool.Pool
}

// New creates a Repository using the given pool.
func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// NextCaseNumber returns the next case number in format SIR000001. Must be called within a transaction that creates the case.
func (r *Repository) NextCaseNumber(ctx context.Context, tx pgx.Tx) (string, error) {
	var n int64
	err := tx.QueryRow(ctx, `SELECT nextval('case_number_seq')`).Scan(&n)
	if err != nil {
		return "", fmt.Errorf("case_number sequence: %w", err)
	}
	return fmt.Sprintf("SIR%06d", n), nil
}

// CreateCase inserts a case. Call NextCaseNumber in same TX and set c.CaseNumber before calling.
func (r *Repository) CreateCase(ctx context.Context, tx pgx.Tx, c *domain.Case) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO cases (
			id, case_number, short_description, description, state, priority, severity,
			opened_by_user_id, opened_time, event_occurred_time, event_received_time,
			affected_user_id, assigned_user_id, assignment_group_id, alert_rule_id,
			active_duration_seconds, accuracy, determination, impact,
			closure_code, closure_reason, closed_by_user_id, closed_time,
			is_active, version_no, created_at, updated_at,
			followup_time, category, subcategory, source, source_tool, source_tool_feature,
			configuration_item, soc_notes, next_steps, csirt_classification, soc_lead_user_id,
			requested_by_user_id, environment_level, environment_type, pdn, impacted_object,
			notification_time, mttr, reassignment_count, assigned_to_count, is_affected_user_vip,
			engineering_document, response_document
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27,
			$28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38,
			$39, $40, $41, $42, $43, $44, $45, $46, $47, $48, $49, $50
		)`,
		c.ID, c.CaseNumber, c.ShortDescription, nullStr(c.Description),
		c.State, c.Priority, c.Severity, c.OpenedByUserID, c.OpenedTime,
		nullTime(c.EventOccurredTime), nullTime(c.EventReceivedTime),
		nullStrPtr(c.AffectedUserID), nullStrPtr(c.AssignedUserID), nullStrPtr(c.AssignmentGroupID), nullStrPtr(c.AlertRuleID),
		c.ActiveDurationSeconds, nullStrPtr(c.Accuracy), nullStrPtr(c.Determination), nullStrPtr(c.Impact),
		nullStrPtr(c.ClosureCode), nullStrPtr(c.ClosureReason), nullStrPtr(c.ClosedByUserID), nullTimePtr(c.ClosedTime),
		c.IsActive, c.VersionNo, timeToTimestamptz(c.CreatedAt), timeToTimestamptz(c.UpdatedAt),
		nullTimePtr(c.FollowupTime), nullStrPtr(c.Category), nullStrPtr(c.Subcategory), nullStrPtr(c.Source), nullStrPtr(c.SourceTool), nullStrPtr(c.SourceToolFeature),
		nullStrPtr(c.ConfigurationItem), nullStrPtr(c.SOCNotes), nullStrPtr(c.NextSteps), nullStrPtr(c.CSIRTClassification), nullStrPtr(c.SOCLeadUserID),
		nullStrPtr(c.RequestedByUserID), nullStrPtr(c.EnvironmentLevel), nullStrPtr(c.EnvironmentType), nullStrPtr(c.PDN), nullStrPtr(c.ImpactedObject),
		nullTimePtr(c.NotificationTime), nullStrPtr(c.MTTR), c.ReassignmentCount, c.AssignedToCount, c.IsAffectedUserVIP,
		nullStrPtr(c.EngineeringDocument), nullStrPtr(c.ResponseDocument),
	)
	return err
}

// GetCaseByID returns a case by id, or nil if not found.
func (r *Repository) GetCaseByID(ctx context.Context, id string) (*domain.Case, error) {
	c := &domain.Case{}
	var desc *string
	var eventOccurred, eventReceived, closedTime, followupTime, notificationTime pgtype.Timestamptz
	var affectedUserID, assignedUserID, assignmentGroupID, alertRuleID, accuracy, determination, impact, closureCode, closureReason, closedByUserID *string
	var category, subcategory, source, sourceTool, sourceToolFeature, configurationItem, socNotes, nextSteps, csirtClassification, socLeadUserID *string
	var requestedByUserID, environmentLevel, environmentType, pdn, impactedObject, mttr, engineeringDocument, responseDocument *string
	err := r.pool.QueryRow(ctx, `
		SELECT id, case_number, short_description, description, state, priority, severity,
			opened_by_user_id, opened_time, event_occurred_time, event_received_time,
			affected_user_id, assigned_user_id, assignment_group_id, alert_rule_id,
			active_duration_seconds, accuracy, determination, impact,
			closure_code, closure_reason, closed_by_user_id, closed_time,
			is_active, version_no, created_at, updated_at,
			followup_time, category, subcategory, source, source_tool, source_tool_feature,
			configuration_item, soc_notes, next_steps, csirt_classification, soc_lead_user_id,
			requested_by_user_id, environment_level, environment_type, pdn, impacted_object,
			notification_time, mttr, reassignment_count, assigned_to_count, is_affected_user_vip,
			engineering_document, response_document
		FROM cases WHERE id = $1`,
		id,
	).Scan(
		&c.ID, &c.CaseNumber, &c.ShortDescription, &desc, &c.State, &c.Priority, &c.Severity,
		&c.OpenedByUserID, &c.OpenedTime, &eventOccurred, &eventReceived,
		&affectedUserID, &assignedUserID, &assignmentGroupID, &alertRuleID,
		&c.ActiveDurationSeconds, &accuracy, &determination, &impact,
		&closureCode, &closureReason, &closedByUserID, &closedTime,
		&c.IsActive, &c.VersionNo, &c.CreatedAt, &c.UpdatedAt,
		&followupTime, &category, &subcategory, &source, &sourceTool, &sourceToolFeature,
		&configurationItem, &socNotes, &nextSteps, &csirtClassification, &socLeadUserID,
		&requestedByUserID, &environmentLevel, &environmentType, &pdn, &impactedObject,
		&notificationTime, &mttr, &c.ReassignmentCount, &c.AssignedToCount, &c.IsAffectedUserVIP,
		&engineeringDocument, &responseDocument,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	c.Description = strVal(desc)
	c.EventOccurredTime = nullTimeToPtr(eventOccurred)
	c.EventReceivedTime = nullTimeToPtr(eventReceived)
	c.ClosedTime = nullTimeToPtr(closedTime)
	c.FollowupTime = nullTimeToPtr(followupTime)
	c.NotificationTime = nullTimeToPtr(notificationTime)
	c.AffectedUserID = affectedUserID
	c.AssignedUserID = assignedUserID
	c.AssignmentGroupID = assignmentGroupID
	c.AlertRuleID = alertRuleID
	c.Accuracy = accuracy
	c.Determination = determination
	c.Impact = impact
	c.ClosureCode = closureCode
	c.ClosureReason = closureReason
	c.ClosedByUserID = closedByUserID
	c.Category = category
	c.Subcategory = subcategory
	c.Source = source
	c.SourceTool = sourceTool
	c.SourceToolFeature = sourceToolFeature
	c.ConfigurationItem = configurationItem
	c.SOCNotes = socNotes
	c.NextSteps = nextSteps
	c.CSIRTClassification = csirtClassification
	c.SOCLeadUserID = socLeadUserID
	c.RequestedByUserID = requestedByUserID
	c.EnvironmentLevel = environmentLevel
	c.EnvironmentType = environmentType
	c.PDN = pdn
	c.ImpactedObject = impactedObject
	c.MTTR = mttr
	c.EngineeringDocument = engineeringDocument
	c.ResponseDocument = responseDocument
	return c, nil
}

// UpdateCase updates a case; expects version_no to match. Returns ErrVersionConflict if no row updated.
func (r *Repository) UpdateCase(ctx context.Context, c *domain.Case) error {
	cmd, err := r.pool.Exec(ctx, `
		UPDATE cases SET
			short_description = $2, description = $3, state = $4, priority = $5, severity = $6,
			event_occurred_time = $7, event_received_time = $8,
			affected_user_id = $9, assigned_user_id = $10, assignment_group_id = $11,
			active_duration_seconds = $12, accuracy = $13, determination = $14, impact = $15,
			closure_code = $16, closure_reason = $17, closed_by_user_id = $18, closed_time = $19,
			is_active = $20, version_no = version_no + 1, updated_at = $21,
			followup_time = $22, category = $23, subcategory = $24, source = $25, source_tool = $26, source_tool_feature = $27,
			configuration_item = $28, soc_notes = $29, next_steps = $30, csirt_classification = $31, soc_lead_user_id = $32,
			requested_by_user_id = $33, environment_level = $34, environment_type = $35, pdn = $36, impacted_object = $37,
			notification_time = $38, mttr = $39, reassignment_count = $40, assigned_to_count = $41, is_affected_user_vip = $42,
			engineering_document = $43, response_document = $44
		WHERE id = $1 AND version_no = $45`,
		c.ID, c.ShortDescription, nullStr(c.Description), c.State, c.Priority, c.Severity,
		nullTime(c.EventOccurredTime), nullTime(c.EventReceivedTime),
		nullStrPtr(c.AffectedUserID), nullStrPtr(c.AssignedUserID), nullStrPtr(c.AssignmentGroupID),
		c.ActiveDurationSeconds, nullStrPtr(c.Accuracy), nullStrPtr(c.Determination), nullStrPtr(c.Impact),
		nullStrPtr(c.ClosureCode), nullStrPtr(c.ClosureReason), nullStrPtr(c.ClosedByUserID), nullTimePtr(c.ClosedTime),
		c.IsActive, timeToTimestamptz(c.UpdatedAt),
		nullTimePtr(c.FollowupTime), nullStrPtr(c.Category), nullStrPtr(c.Subcategory), nullStrPtr(c.Source), nullStrPtr(c.SourceTool), nullStrPtr(c.SourceToolFeature),
		nullStrPtr(c.ConfigurationItem), nullStrPtr(c.SOCNotes), nullStrPtr(c.NextSteps), nullStrPtr(c.CSIRTClassification), nullStrPtr(c.SOCLeadUserID),
		nullStrPtr(c.RequestedByUserID), nullStrPtr(c.EnvironmentLevel), nullStrPtr(c.EnvironmentType), nullStrPtr(c.PDN), nullStrPtr(c.ImpactedObject),
		nullTimePtr(c.NotificationTime), nullStrPtr(c.MTTR), c.ReassignmentCount, c.AssignedToCount, c.IsAffectedUserVIP,
		nullStrPtr(c.EngineeringDocument), nullStrPtr(c.ResponseDocument),
		c.VersionNo,
	)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrVersionConflict
	}
	return nil
}

// ListCasesFilter for ListCases.
type ListCasesFilter struct {
	State              *string
	Priority           *string
	Severity           *string
	AssignmentGroupID  *string
	AssignedUserID     *string
	AffectedUserID     *string
	OpenedTimeAfter    *string // RFC3339 or comparable
	OpenedTimeBefore   *string
	PageSize           int32
	PageToken          string // offset-based for simplicity: numeric string
}

// ListCases returns cases ordered by opened_time DESC with filters and pagination.
func (r *Repository) ListCases(ctx context.Context, f ListCasesFilter) ([]*domain.Case, string, error) {
	pageSize := f.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := int64(0)
	if f.PageToken != "" {
		if n, err := strconv.ParseInt(f.PageToken, 10, 64); err == nil && n >= 0 {
			offset = n
		}
	}
	args := []interface{}{}
	argNum := 0
	addArg := func(v interface{}) {
		argNum++
		args = append(args, v)
	}
	q := `
		SELECT id, case_number, short_description, description, state, priority, severity,
			opened_by_user_id, opened_time, event_occurred_time, event_received_time,
			affected_user_id, assigned_user_id, assignment_group_id, alert_rule_id,
			active_duration_seconds, accuracy, determination, impact,
			closure_code, closure_reason, closed_by_user_id, closed_time,
			is_active, version_no, created_at, updated_at,
			followup_time, category, subcategory, source, source_tool, source_tool_feature,
			configuration_item, soc_notes, next_steps, csirt_classification, soc_lead_user_id,
			requested_by_user_id, environment_level, environment_type, pdn, impacted_object,
			notification_time, mttr, reassignment_count, assigned_to_count, is_affected_user_vip,
			engineering_document, response_document
		FROM cases WHERE 1=1`
	if f.State != nil && *f.State != "" {
		addArg(*f.State)
		q += fmt.Sprintf(" AND state = $%d", argNum)
	}
	if f.Priority != nil && *f.Priority != "" {
		addArg(*f.Priority)
		q += fmt.Sprintf(" AND priority = $%d", argNum)
	}
	if f.Severity != nil && *f.Severity != "" {
		addArg(*f.Severity)
		q += fmt.Sprintf(" AND severity = $%d", argNum)
	}
	if f.AssignmentGroupID != nil && *f.AssignmentGroupID != "" {
		addArg(*f.AssignmentGroupID)
		q += fmt.Sprintf(" AND assignment_group_id = $%d", argNum)
	}
	if f.AssignedUserID != nil && *f.AssignedUserID != "" {
		addArg(*f.AssignedUserID)
		q += fmt.Sprintf(" AND assigned_user_id = $%d", argNum)
	}
	if f.AffectedUserID != nil && *f.AffectedUserID != "" {
		addArg(*f.AffectedUserID)
		q += fmt.Sprintf(" AND affected_user_id = $%d", argNum)
	}
	if f.OpenedTimeAfter != nil && *f.OpenedTimeAfter != "" {
		addArg(*f.OpenedTimeAfter)
		q += fmt.Sprintf(" AND opened_time >= $%d", argNum)
	}
	if f.OpenedTimeBefore != nil && *f.OpenedTimeBefore != "" {
		addArg(*f.OpenedTimeBefore)
		q += fmt.Sprintf(" AND opened_time <= $%d", argNum)
	}
	q += " ORDER BY opened_time DESC"
	addArg(pageSize + 1)
	q += fmt.Sprintf(" LIMIT $%d OFFSET %d", argNum, offset)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()
	var list []*domain.Case
	for rows.Next() {
		c := &domain.Case{}
		var desc *string
		var eventOccurred, eventReceived, closedTime, followupTime, notificationTime pgtype.Timestamptz
		var affectedUserID, assignedUserID, assignmentGroupID, alertRuleID, accuracy, determination, impact, closureCode, closureReason, closedByUserID *string
		var category, subcategory, source, sourceTool, sourceToolFeature, configurationItem, socNotes, nextSteps, csirtClassification, socLeadUserID *string
		var requestedByUserID, environmentLevel, environmentType, pdn, impactedObject, mttr, engineeringDocument, responseDocument *string
		err := rows.Scan(
			&c.ID, &c.CaseNumber, &c.ShortDescription, &desc, &c.State, &c.Priority, &c.Severity,
			&c.OpenedByUserID, &c.OpenedTime, &eventOccurred, &eventReceived,
			&affectedUserID, &assignedUserID, &assignmentGroupID, &alertRuleID,
			&c.ActiveDurationSeconds, &accuracy, &determination, &impact,
			&closureCode, &closureReason, &closedByUserID, &closedTime,
			&c.IsActive, &c.VersionNo, &c.CreatedAt, &c.UpdatedAt,
			&followupTime, &category, &subcategory, &source, &sourceTool, &sourceToolFeature,
			&configurationItem, &socNotes, &nextSteps, &csirtClassification, &socLeadUserID,
			&requestedByUserID, &environmentLevel, &environmentType, &pdn, &impactedObject,
			&notificationTime, &mttr, &c.ReassignmentCount, &c.AssignedToCount, &c.IsAffectedUserVIP,
			&engineeringDocument, &responseDocument,
		)
		if err != nil {
			return nil, "", err
		}
		c.Description = strVal(desc)
		c.EventOccurredTime = nullTimeToPtr(eventOccurred)
		c.EventReceivedTime = nullTimeToPtr(eventReceived)
		c.ClosedTime = nullTimeToPtr(closedTime)
		c.FollowupTime = nullTimeToPtr(followupTime)
		c.NotificationTime = nullTimeToPtr(notificationTime)
		c.AffectedUserID = affectedUserID
		c.AssignedUserID = assignedUserID
		c.AssignmentGroupID = assignmentGroupID
		c.AlertRuleID = alertRuleID
		c.Accuracy = accuracy
		c.Determination = determination
		c.Impact = impact
		c.Category = category
		c.Subcategory = subcategory
		c.Source = source
		c.SourceTool = sourceTool
		c.SourceToolFeature = sourceToolFeature
		c.ConfigurationItem = configurationItem
		c.SOCNotes = socNotes
		c.NextSteps = nextSteps
		c.CSIRTClassification = csirtClassification
		c.SOCLeadUserID = socLeadUserID
		c.ClosureCode = closureCode
		c.ClosureReason = closureReason
		c.ClosedByUserID = closedByUserID
		c.ClosedTime = nullTimeToPtr(closedTime)
		c.RequestedByUserID = requestedByUserID
		c.EnvironmentLevel = environmentLevel
		c.EnvironmentType = environmentType
		c.PDN = pdn
		c.ImpactedObject = impactedObject
		c.MTTR = mttr
		c.EngineeringDocument = engineeringDocument
		c.ResponseDocument = responseDocument
		list = append(list, c)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}
	nextToken := ""
	if len(list) > int(pageSize) {
		list = list[:pageSize]
		nextToken = strconv.FormatInt(offset+int64(pageSize), 10)
	} else if len(list) == int(pageSize) {
		nextToken = strconv.FormatInt(offset+int64(pageSize), 10)
	}
	return list, nextToken, nil
}

// CreateWorknote inserts a worknote.
func (r *Repository) CreateWorknote(ctx context.Context, w *domain.Worknote) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO case_worknotes (id, case_id, note_text, note_type, created_by_user_id, created_at, updated_at, is_deleted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		w.ID, w.CaseID, w.NoteText, w.NoteType, w.CreatedByUserID, timeToTimestamptz(w.CreatedAt), nullTimePtrToTimestamptz(w.UpdatedAt), w.IsDeleted,
	)
	return err
}

// ListWorknotes returns worknotes for a case, ordered by created_at DESC (newest first), with pagination.
func (r *Repository) ListWorknotes(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*domain.Worknote, string, error) {
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := int64(0)
	if pageToken != "" {
		if n, err := strconv.ParseInt(pageToken, 10, 64); err == nil && n >= 0 {
			offset = n
		}
	}
	q := `
		SELECT id, case_id, note_text, note_type, created_by_user_id, created_at, updated_at, is_deleted
		FROM case_worknotes WHERE case_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, q, caseID, pageSize+1, offset)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()
	var list []*domain.Worknote
	for rows.Next() {
		w := &domain.Worknote{}
		var updatedAt pgtype.Timestamptz
		err := rows.Scan(&w.ID, &w.CaseID, &w.NoteText, &w.NoteType, &w.CreatedByUserID, &w.CreatedAt, &updatedAt, &w.IsDeleted)
		if err != nil {
			return nil, "", err
		}
		w.UpdatedAt = nullTimeToPtr(updatedAt)
		list = append(list, w)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}
	nextToken := ""
	if len(list) > int(pageSize) {
		list = list[:pageSize]
		nextToken = strconv.FormatInt(offset+int64(pageSize), 10)
	} else if len(list) == int(pageSize) {
		nextToken = strconv.FormatInt(offset+int64(pageSize), 10)
	}
	return list, nextToken, nil
}

// RunInTx runs fn inside a transaction. If fn returns error, transaction is rolled back.
func (r *Repository) RunInTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

var ErrVersionConflict = fmt.Errorf("optimistic concurrency conflict: version_no mismatch")

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nullStrPtr(s *string) interface{} {
	if s == nil || *s == "" {
		return nil
	}
	return *s
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func nullTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}

func nullTimePtr(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}

func nullTimeToPtr(nt pgtype.Timestamptz) *time.Time {
	if !nt.Valid {
		return nil
	}
	return &nt.Time
}

// timeToTimestamptz converts time.Time to pgtype.Timestamptz so pgx can encode it for timestamptz columns (avoids "cannot find encode plan" for zero or raw time).
func timeToTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// nullTimePtrToTimestamptz returns nil or pgtype.Timestamptz for optional timestamptz columns in Exec.
func nullTimePtrToTimestamptz(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

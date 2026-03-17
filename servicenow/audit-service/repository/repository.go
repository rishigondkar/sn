package repository

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

const currentLayer = "repository"

// AuditEventRow is the domain representation of a single audit event for repository use.
type AuditEventRow struct {
	ID               string
	EventID          string
	EventType        string
	SourceService    string
	EntityType       string
	EntityID         string
	ParentEntityType sql.NullString
	ParentEntityID   sql.NullString
	CaseID           sql.NullString
	ObservableID     sql.NullString
	Action           string
	ActorUserID      sql.NullString
	ActorName        sql.NullString
	RequestID        sql.NullString
	CorrelationID    sql.NullString
	ChangeSummary    sql.NullString
	BeforeData       []byte // JSONB as bytes
	AfterData        []byte
	Metadata         []byte
	OccurredAt       time.Time
	IngestedAt       time.Time
}

// ListFilter holds common list options (pagination and time range).
type ListFilter struct {
	PageSize       int32
	PageToken     string
	OccurredAfter  *time.Time
	OccurredBefore *time.Time
	OldestFirst    bool
}

// ListResult holds a page of audit events and the next page token.
type ListResult struct {
	Items         []*AuditEventRow
	NextPageToken string
}

// Repository provides data access for audit events.
type Repository struct {
	DB *sql.DB
}

// NewRepository returns a new Repository using the given DB.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

// InsertEvent inserts one audit event. It is idempotent: if event_id already exists, no error and no duplicate row.
func (r *Repository) InsertEvent(ctx context.Context, row *AuditEventRow) (inserted bool, err error) {
	query := `
	INSERT INTO audit_events (
		id, event_id, event_type, source_service, entity_type, entity_id,
		parent_entity_type, parent_entity_id, case_id, observable_id, action,
		actor_user_id, actor_name, request_id, correlation_id, change_summary,
		before_data, after_data, metadata, occurred_at, ingested_at
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21
	)
	ON CONFLICT (event_id) DO NOTHING
	`
	result, err := r.DB.ExecContext(ctx, query,
		row.ID, row.EventID, row.EventType, row.SourceService, row.EntityType, row.EntityID,
		nullString(row.ParentEntityType), nullUUID(row.ParentEntityID), nullUUID(row.CaseID), nullUUID(row.ObservableID),
		row.Action, nullUUIDSafe(row.ActorUserID), nullString(row.ActorName), nullString(row.RequestID), nullString(row.CorrelationID),
		nullString(row.ChangeSummary), nullJSON(row.BeforeData), nullJSON(row.AfterData), nullJSON(row.Metadata),
		row.OccurredAt, row.IngestedAt,
	)
	if err != nil {
		return false, err
	}
	n, _ := result.RowsAffected()
	return n == 1, nil
}

func nullString(n sql.NullString) interface{} {
	if n.Valid {
		return n.String
	}
	return nil
}

func nullUUID(n sql.NullString) interface{} {
	if n.Valid && n.String != "" {
		return n.String
	}
	return nil
}

// nullUUIDSafe returns the string for UUID columns only if it parses as a valid UUID; otherwise nil (so DB stores NULL). Use for actor_user_id when callers may send non-UUID placeholders like "frontend-user".
func nullUUIDSafe(n sql.NullString) interface{} {
	if !n.Valid || n.String == "" {
		return nil
	}
	if _, err := uuid.Parse(n.String); err != nil {
		return nil
	}
	return n.String
}

func nullJSON(b []byte) interface{} {
	if len(b) == 0 {
		return nil
	}
	return b
}

// ListByCaseID returns audit events for the given case_id with pagination and optional time filter.
func (r *Repository) ListByCaseID(ctx context.Context, caseID string, f ListFilter) (*ListResult, error) {
	return r.list(ctx, listParams{
		WhereColumn: "case_id",
		WhereValue:  caseID,
		Filter:      f,
	})
}

// ListByObservableID returns audit events for the given observable_id.
func (r *Repository) ListByObservableID(ctx context.Context, observableID string, f ListFilter) (*ListResult, error) {
	return r.list(ctx, listParams{
		WhereColumn: "observable_id",
		WhereValue:  observableID,
		Filter:      f,
	})
}

// ListByEntity returns audit events for the given entity_type and entity_id.
func (r *Repository) ListByEntity(ctx context.Context, entityType, entityID string, f ListFilter) (*ListResult, error) {
	return r.list(ctx, listParams{
		WhereColumn:  "entity_type",
		WhereValue:   entityType,
		WhereColumn2: "entity_id",
		WhereValue2:  entityID,
		Filter:       f,
	})
}

// ListByActorUserID returns audit events for the given actor_user_id.
func (r *Repository) ListByActorUserID(ctx context.Context, actorUserID string, f ListFilter) (*ListResult, error) {
	return r.list(ctx, listParams{
		WhereColumn: "actor_user_id",
		WhereValue:  actorUserID,
		Filter:      f,
	})
}

// ListByCorrelationID returns audit events for the given correlation_id.
func (r *Repository) ListByCorrelationID(ctx context.Context, correlationID string, f ListFilter) (*ListResult, error) {
	return r.list(ctx, listParams{
		WhereColumn: "correlation_id",
		WhereValue:  correlationID,
		Filter:      f,
	})
}

type listParams struct {
	WhereColumn  string
	WhereValue   string
	WhereColumn2 string
	WhereValue2  string
	Filter       ListFilter
}

func (r *Repository) list(ctx context.Context, p listParams) (*ListResult, error) {
	pageSize := p.Filter.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 100 {
		pageSize = 100
	}

	order := "DESC"
	if p.Filter.OldestFirst {
		order = "ASC"
	}

	// Keyset pagination: token is "occurred_at_unix|id"
	var cursorOccurred *time.Time
	var cursorID string
	if p.Filter.PageToken != "" {
		parseCursor(p.Filter.PageToken, &cursorOccurred, &cursorID)
	}

	whereClause := p.WhereColumn + " = $1"
	args := []interface{}{p.WhereValue}
	argNum := 2
	if p.WhereColumn2 != "" {
		whereClause += " AND " + p.WhereColumn2 + " = $" + strconv.Itoa(argNum)
		args = append(args, p.WhereValue2)
		argNum++
	}
	if p.Filter.OccurredAfter != nil {
		whereClause += " AND occurred_at >= $" + strconv.Itoa(argNum)
		args = append(args, *p.Filter.OccurredAfter)
		argNum++
	}
	if p.Filter.OccurredBefore != nil {
		whereClause += " AND occurred_at <= $" + strconv.Itoa(argNum)
		args = append(args, *p.Filter.OccurredBefore)
		argNum++
	}
	if cursorOccurred != nil && cursorID != "" {
		if p.Filter.OldestFirst {
			whereClause += " AND (occurred_at, id) > ($" + strconv.Itoa(argNum) + ", $" + strconv.Itoa(argNum+1) + ")"
		} else {
			whereClause += " AND (occurred_at, id) < ($" + strconv.Itoa(argNum) + ", $" + strconv.Itoa(argNum+1) + ")"
		}
		args = append(args, *cursorOccurred, cursorID)
		argNum += 2
	}

	limit := pageSize + 1
	args = append(args, limit)

	query := `
		SELECT id, event_id, event_type, source_service, entity_type, entity_id,
			parent_entity_type, parent_entity_id, case_id, observable_id, action,
			actor_user_id, actor_name, request_id, correlation_id, change_summary,
			before_data, after_data, metadata, occurred_at, ingested_at
		FROM audit_events
		WHERE ` + whereClause + `
		ORDER BY occurred_at ` + order + `, id ` + order + `
		LIMIT $` + strconv.Itoa(argNum)

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*AuditEventRow
	for rows.Next() {
		var row AuditEventRow
		var parentEntityType, parentEntityID, caseID, observableID sql.NullString
		var actorUserID, actorName, requestID, correlationID, changeSummary sql.NullString
		err := rows.Scan(
			&row.ID, &row.EventID, &row.EventType, &row.SourceService, &row.EntityType, &row.EntityID,
			&parentEntityType, &parentEntityID, &caseID, &observableID, &row.Action,
			&actorUserID, &actorName, &requestID, &correlationID, &changeSummary,
			&row.BeforeData, &row.AfterData, &row.Metadata, &row.OccurredAt, &row.IngestedAt,
		)
		if err != nil {
			return nil, err
		}
		row.ParentEntityType = parentEntityType
		row.ParentEntityID = parentEntityID
		row.CaseID = caseID
		row.ObservableID = observableID
		row.ActorUserID = actorUserID
		row.ActorName = actorName
		row.RequestID = requestID
		row.CorrelationID = correlationID
		row.ChangeSummary = changeSummary
		items = append(items, &row)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	var nextPageToken string
	if len(items) > int(pageSize) {
		last := items[pageSize]
		items = items[:pageSize]
		nextPageToken = makeCursor(last.OccurredAt, last.ID)
	}

	return &ListResult{Items: items, NextPageToken: nextPageToken}, nil
}

// makeCursor encodes occurred_at and id for keyset pagination (opaque page_token).
func makeCursor(occurredAt time.Time, id string) string {
	return occurredAt.UTC().Format(time.RFC3339Nano) + "|" + id
}

// parseCursor decodes page_token into occurred_at and id. Invalid tokens leave cursor nil/empty.
func parseCursor(token string, occurredAt **time.Time, id *string) {
	parts := strings.SplitN(token, "|", 2)
	if len(parts) != 2 {
		return
	}
	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return
	}
	*occurredAt = &t
	*id = parts[1]
}

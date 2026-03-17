package consumer

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/servicenow/audit-service/repository"
)

// AuditEnvelope is the platform contract audit event message body (JSON).
type AuditEnvelope struct {
	EventID          string                 `json:"event_id"`
	EventType        string                 `json:"event_type"`
	SourceService    string                 `json:"source_service"`
	EntityType       string                 `json:"entity_type"`
	EntityID         string                 `json:"entity_id"`
	ParentEntityType string                 `json:"parent_entity_type"`
	ParentEntityID   string                 `json:"parent_entity_id"`
	CaseID           string                 `json:"case_id"`
	ObservableID     string                 `json:"observable_id"`
	Action           string                 `json:"action"`
	ActorUserID      string                 `json:"actor_user_id"`
	ActorName        string                 `json:"actor_name"`
	RequestID        string                 `json:"request_id"`
	CorrelationID    string                 `json:"correlation_id"`
	ChangeSummary    string                 `json:"change_summary"`
	BeforeData       map[string]interface{} `json:"before_data"`
	AfterData        map[string]interface{} `json:"after_data"`
	Metadata         map[string]interface{} `json:"metadata"`
	OccurredAt       string                 `json:"occurred_at"`
}

// IncomingMessage represents one message from the event bus.
type IncomingMessage struct {
	Body []byte
	Ack  func()
	Nack func()
}

// MessageSource is the abstraction for subscribing to the audit event bus (e.g. SQS, Kafka).
type MessageSource interface {
	Receive(ctx context.Context) (*IncomingMessage, error)
}

// Consumer subscribes to the event bus, validates envelopes, and persists idempotently.
type Consumer struct {
	Repo   *repository.Repository
	Source MessageSource
	// MaxRetries is the number of processing attempts before Nack/DeadLetter.
	MaxRetries int
	// RetryDelay is the delay between retries.
	RetryDelay time.Duration
}

// ValidateEnvelope checks required fields of the audit event envelope. Returns a descriptive error if invalid.
func ValidateEnvelope(e *AuditEnvelope) error {
	if e.EventID == "" {
		return ErrMissingField("event_id")
	}
	if e.EventType == "" {
		return ErrMissingField("event_type")
	}
	if e.SourceService == "" {
		return ErrMissingField("source_service")
	}
	if e.EntityType == "" {
		return ErrMissingField("entity_type")
	}
	if e.EntityID == "" {
		return ErrMissingField("entity_id")
	}
	if e.Action == "" {
		return ErrMissingField("action")
	}
	if e.OccurredAt == "" {
		return ErrMissingField("occurred_at")
	}
	_, err := time.Parse(time.RFC3339, e.OccurredAt)
	if err != nil {
		_, err = time.Parse(time.RFC3339Nano, e.OccurredAt)
	}
	if err != nil {
		return ErrInvalidField{Field: "occurred_at", Err: err}
	}
	return nil
}

// ErrMissingField is returned when a required envelope field is missing.
type ErrMissingField string

func (f ErrMissingField) Error() string { return "missing required field: " + string(f) }

// ErrInvalidField is returned when a field value is invalid.
type ErrInvalidField struct {
	Field string
	Err   error
}

func (e ErrInvalidField) Error() string { return "invalid " + e.Field + ": " + e.Err.Error() }

// Run processes messages from the source until ctx is cancelled. Idempotent by event_id.
func (c *Consumer) Run(ctx context.Context) {
	maxRetries := c.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}
	retryDelay := c.RetryDelay
	if retryDelay <= 0 {
		retryDelay = 2 * time.Second
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		msg, err := c.Source.Receive(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			slog.Default().ErrorContext(ctx, "consumer receive error", slog.Any("error", err))
			continue
		}
		if msg == nil {
			continue
		}

		processed := c.processWithRetry(ctx, msg.Body, maxRetries, retryDelay)
		if processed {
			msg.Ack()
		} else {
			msg.Nack()
		}
	}
}

func (c *Consumer) processWithRetry(ctx context.Context, body []byte, maxRetries int, retryDelay time.Duration) bool {
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return false
			case <-time.After(retryDelay):
			}
		}
		err := c.processOne(ctx, body)
		if err == nil {
			return true
		}
		lastErr = err
		slog.Default().WarnContext(ctx, "consumer process attempt failed", slog.Int("attempt", attempt+1), slog.Any("error", err))
	}
	slog.Default().ErrorContext(ctx, "consumer processing failed after retries; message will be nack'd/dead-lettered", slog.Any("error", lastErr))
	return false
}

func (c *Consumer) processOne(ctx context.Context, body []byte) error {
	var envelope AuditEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return err
	}
	if err := ValidateEnvelope(&envelope); err != nil {
		return err
	}

	occurredAt, _ := time.Parse(time.RFC3339, envelope.OccurredAt)
	if occurredAt.IsZero() {
		occurredAt, _ = time.Parse(time.RFC3339Nano, envelope.OccurredAt)
	}
	ingestedAt := time.Now().UTC()

	row := envelopeToRow(&envelope, occurredAt, ingestedAt)
	inserted, err := c.Repo.InsertEvent(ctx, row)
	if err != nil {
		return err
	}
	if !inserted {
		// Idempotent: duplicate event_id, already stored
		slog.Default().DebugContext(ctx, "audit event already present", slog.String("event_id", envelope.EventID))
	}
	return nil
}

func envelopeToRow(e *AuditEnvelope, occurredAt, ingestedAt time.Time) *repository.AuditEventRow {
	row := &repository.AuditEventRow{
		ID:            uuid.New().String(),
		EventID:       e.EventID,
		EventType:     e.EventType,
		SourceService: e.SourceService,
		EntityType:    e.EntityType,
		EntityID:      e.EntityID,
		Action:        e.Action,
		OccurredAt:    occurredAt.UTC(),
		IngestedAt:    ingestedAt,
	}
	row.ParentEntityType = strPtr(e.ParentEntityType)
	row.ParentEntityID = strPtr(e.ParentEntityID)
	row.CaseID = strPtr(e.CaseID)
	row.ObservableID = strPtr(e.ObservableID)
	row.ActorUserID = strPtr(e.ActorUserID)
	row.ActorName = strPtr(e.ActorName)
	row.RequestID = strPtr(e.RequestID)
	row.CorrelationID = strPtr(e.CorrelationID)
	row.ChangeSummary = strPtr(e.ChangeSummary)
	row.BeforeData = jsonBytes(e.BeforeData)
	row.AfterData = jsonBytes(e.AfterData)
	row.Metadata = jsonBytes(e.Metadata)
	return row
}

func strPtr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func jsonBytes(m map[string]interface{}) []byte {
	if len(m) == 0 {
		return nil
	}
	b, _ := json.Marshal(m)
	return b
}

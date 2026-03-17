package audit

import "context"

// Event is the audit event envelope per platform contract (JSON on the wire).
type Event struct {
	EventID          string                 `json:"event_id"`
	EventType        string                 `json:"event_type"`
	SourceService    string                 `json:"source_service"`
	EntityType       string                 `json:"entity_type"`
	EntityID         string                 `json:"entity_id"`
	ParentEntityType string                 `json:"parent_entity_type,omitempty"`
	ParentEntityID   string                 `json:"parent_entity_id,omitempty"`
	CaseID           string                 `json:"case_id,omitempty"`
	Action           string                 `json:"action"`
	ActorUserID      string                 `json:"actor_user_id"`
	ActorName        string                 `json:"actor_name,omitempty"`
	RequestID        string                 `json:"request_id,omitempty"`
	CorrelationID    string                 `json:"correlation_id,omitempty"`
	ChangeSummary    string                 `json:"change_summary,omitempty"`
	BeforeData       map[string]interface{} `json:"before_data,omitempty"`
	AfterData        map[string]interface{} `json:"after_data,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	OccurredAt       string                 `json:"occurred_at"` // RFC3339
}

// Publisher publishes audit events to the event bus. Implementations must not block indefinitely.
type Publisher interface {
	Publish(ctx context.Context, evt Event) error
}

// NoopPublisher discards events (for tests or when audit is disabled).
type NoopPublisher struct{}

func (NoopPublisher) Publish(ctx context.Context, evt Event) error { return nil }

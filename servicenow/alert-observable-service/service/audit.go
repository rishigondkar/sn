package service

import "context"

// AuditEvent is the platform audit event envelope (JSON-serializable).
type AuditEvent struct {
	EventID           string                 `json:"event_id"`
	EventType         string                 `json:"event_type"`
	SourceService     string                 `json:"source_service"`
	EntityType        string                 `json:"entity_type"`
	EntityID          string                 `json:"entity_id"`
	ParentEntityType  string                 `json:"parent_entity_type,omitempty"`
	ParentEntityID    string                 `json:"parent_entity_id,omitempty"`
	CaseID            string                 `json:"case_id,omitempty"`
	ObservableID      string                 `json:"observable_id,omitempty"`
	Action            string                 `json:"action"`
	ActorUserID       string                 `json:"actor_user_id"`
	ActorName         string                 `json:"actor_name,omitempty"`
	RequestID         string                 `json:"request_id"`
	CorrelationID     string                 `json:"correlation_id"`
	ChangeSummary     string                 `json:"change_summary,omitempty"`
	BeforeData        map[string]interface{} `json:"before_data,omitempty"`
	AfterData         map[string]interface{} `json:"after_data,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	OccurredAt        string                 `json:"occurred_at"` // RFC3339
}

// AuditPublisher publishes audit events (e.g. to event bus). Publish only after successful commit.
type AuditPublisher interface {
	Publish(ctx context.Context, ev AuditEvent) error
}

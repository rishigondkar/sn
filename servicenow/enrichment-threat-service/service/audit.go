package service

import "context"

// AuditEventEnvelope matches the platform contract audit event shape (JSON).
type AuditEventEnvelope struct {
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
	ActorName         string                 `json:"actor_name"`
	RequestID         string                 `json:"request_id"`
	CorrelationID     string                 `json:"correlation_id"`
	ChangeSummary     string                 `json:"change_summary"`
	BeforeData        map[string]interface{} `json:"before_data,omitempty"`
	AfterData         map[string]interface{} `json:"after_data,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	OccurredAt        string                 `json:"occurred_at"` // RFC3339
}

// AuditEventPublisher publishes audit events to the event bus (e.g. SNS/SQS).
type AuditEventPublisher interface {
	Publish(ctx context.Context, event AuditEventEnvelope) error
}

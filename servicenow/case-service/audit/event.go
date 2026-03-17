package audit

import "time"

// Event is the audit event envelope per platform contract (JSON on the wire).
type Event struct {
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
	RequestID         string                 `json:"request_id,omitempty"`
	CorrelationID     string                 `json:"correlation_id,omitempty"`
	ChangeSummary     string                 `json:"change_summary,omitempty"`
	BeforeData        map[string]interface{} `json:"before_data,omitempty"`
	AfterData         map[string]interface{} `json:"after_data,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	OccurredAt        time.Time              `json:"occurred_at"`
}

const SourceService = "case-service"

// Event types for case service.
const (
	EventTypeCaseCreated    = "case.created"
	EventTypeCaseUpdated    = "case.updated"
	EventTypeCaseStateChanged = "case.state.changed"
	EventTypeCaseAssigned   = "case.assigned"
	EventTypeCaseClosed     = "case.closed"
	EventTypeWorknoteCreated = "worknote.created"
)

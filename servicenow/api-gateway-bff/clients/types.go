package clients

import "time"

// Internal DTOs used by the gateway. Map from downstream gRPC responses when real stubs are wired.

// Case represents case core data from Case Service.
type Case struct {
	ID                   string     `json:"id"`
	CaseNumber           string     `json:"case_number"`
	Title                string     `json:"title"`
	State                string     `json:"state"`
	Priority             string     `json:"priority"`
	Severity             string     `json:"severity,omitempty"`
	AssignedUserID       string     `json:"assigned_user_id,omitempty"`
	AssignmentGroupID    string     `json:"assignment_group_id,omitempty"`
	OpenedByUserID       string     `json:"opened_by_user_id,omitempty"`
	AffectedUserID       string     `json:"affected_user_id,omitempty"`
	Description          string     `json:"description,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	FollowupTime         *time.Time `json:"followup_time,omitempty"`
	Category             string     `json:"category,omitempty"`
	Subcategory          string     `json:"subcategory,omitempty"`
	Source               string     `json:"source,omitempty"`
	SourceTool           string     `json:"source_tool,omitempty"`
	SourceToolFeature    string     `json:"source_tool_feature,omitempty"`
	ConfigurationItem    string     `json:"configuration_item,omitempty"`
	SOCNotes             string     `json:"soc_notes,omitempty"`
	NextSteps            string     `json:"next_steps,omitempty"`
	CSIRTClassification  string     `json:"csirt_classification,omitempty"`
	SOCLeadUserID        string     `json:"soc_lead_user_id,omitempty"`
	RequestedByUserID    string     `json:"requested_by_user_id,omitempty"`
	EnvironmentLevel     string     `json:"environment_level,omitempty"`
	EnvironmentType      string     `json:"environment_type,omitempty"`
	PDN                  string     `json:"pdn,omitempty"`
	ImpactedObject       string     `json:"impacted_object,omitempty"`
	NotificationTime     *time.Time `json:"notification_time,omitempty"`
	OpenedTime           *time.Time `json:"opened_time,omitempty"`
	EventOccurredTime    *time.Time `json:"event_occurred_time,omitempty"`
	MTTR                 string     `json:"mttr,omitempty"`
	ReassignmentCount    int32      `json:"reassignment_count,omitempty"`
	AssignedToCount      int32      `json:"assigned_to_count,omitempty"`
	IsAffectedUserVIP    bool       `json:"is_affected_user_vip,omitempty"`
	EngineeringDocument  string     `json:"engineering_document,omitempty"`
	ResponseDocument     string     `json:"response_document,omitempty"`
	Accuracy             string     `json:"accuracy,omitempty"`
	Determination        string     `json:"determination,omitempty"`
	ClosureReason        string     `json:"closure_reason,omitempty"`
	ClosedByUserID       string     `json:"closed_by_user_id,omitempty"`
	ClosedTime           *time.Time `json:"closed_time,omitempty"`
}

// Worknote from Case Service.
type Worknote struct {
	ID        string    `json:"id"`
	CaseID    string    `json:"case_id"`
	Content   string    `json:"content"`
	CreatedBy string    `json:"created_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Alert from Alert & Observable Service.
type Alert struct {
	ID         string    `json:"id"`
	CaseID     string    `json:"case_id"`
	Summary    string    `json:"summary,omitempty"`
	Severity   string    `json:"severity,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// Observable from Alert & Observable Service.
type Observable struct {
	ID             string    `json:"id"`
	CaseID         string    `json:"case_id"`
	Type           string    `json:"type"`
	Value          string    `json:"value"`
	TrackingStatus string    `json:"tracking_status,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at,omitempty"`
	Determination  string    `json:"finding,omitempty"`
	IncidentCount  int       `json:"incident_count,omitempty"`
	Notes          string    `json:"notes,omitempty"`
}

// ChildObservable summary for graph.
type ChildObservable struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Value    string `json:"value"`
	ParentID string `json:"parent_id,omitempty"`
}

// SimilarIncident from Alert & Observable Service (enriched with similar case info in orchestrator).
type SimilarIncident struct {
	ID                     string `json:"id"`
	CaseID                 string `json:"case_id"`
	SimilarCaseID          string `json:"similar_case_id,omitempty"`
	Summary                string `json:"summary,omitempty"`
	Similarity             string `json:"similarity,omitempty"`
	SharedObservableValues string `json:"shared_observable_values,omitempty"` // common observable value(s), for display
	// Enriched by orchestrator from case service:
	SimilarCaseNumber   string `json:"similar_case_number,omitempty"`
	SimilarCaseTitle    string `json:"similar_case_title,omitempty"`
	SimilarCaseCreatedAt string `json:"similar_case_created_at,omitempty"`
}

// EnrichmentResult from Enrichment & Threat Lookup Service.
type EnrichmentResult struct {
	ID          string    `json:"id"`
	ObservableID string   `json:"observable_id"`
	CaseID      string   `json:"case_id"`
	Source      string   `json:"source,omitempty"`
	Result      string   `json:"result,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// ThreatLookupResult from Enrichment & Threat Lookup Service.
type ThreatLookupResult struct {
	ID           string    `json:"id"`
	ObservableID string    `json:"observable_id"`
	CaseID       string    `json:"case_id"`
	Provider     string    `json:"provider,omitempty"`
	Verdict      string    `json:"verdict,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// Attachment from Attachment Service.
type Attachment struct {
	ID        string    `json:"id"`
	CaseID    string    `json:"case_id"`
	FileName  string    `json:"file_name"`
	SizeBytes int64     `json:"size_bytes"`
	CreatedAt time.Time `json:"created_at"`
}

// AuditEvent from Audit Service.
type AuditEvent struct {
	EventID       string    `json:"event_id"`
	EventType     string    `json:"event_type"`
	EntityType    string    `json:"entity_type"`
	EntityID      string    `json:"entity_id"`
	Action        string    `json:"action"`
	ActorUserID   string    `json:"actor_user_id,omitempty"`
	ActorName     string    `json:"actor_name,omitempty"`
	ChangeSummary string    `json:"change_summary,omitempty"`
	OccurredAt    time.Time `json:"occurred_at"`
}

// User from Assignment & Reference Service.
type User struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name,omitempty"`
	Email       string `json:"email,omitempty"`
}

// Group from Assignment & Reference Service.
type Group struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

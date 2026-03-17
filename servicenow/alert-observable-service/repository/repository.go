package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AlertRule row.
type AlertRule struct {
	ID              uuid.UUID
	RuleName        string
	RuleType        *string
	SourceSystem    *string
	ExternalRuleKey *string
	Description     *string
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Alert row.
type Alert struct {
	ID                uuid.UUID
	CaseID            uuid.UUID
	AlertRuleID       *uuid.UUID
	SourceSystem      string
	SourceAlertID     *string
	Title             *string
	Description       *string
	EventOccurredTime *time.Time
	EventReceivedTime *time.Time
	Severity          *string
	RawPayload        []byte // JSONB
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// Observable row.
type Observable struct {
	ID              uuid.UUID
	ObservableType  string
	ObservableValue string
	NormalizedValue *string
	FirstSeenTime   *time.Time
	LastSeenTime    *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	IncidentCount   int     // number of cases linking this observable
	Finding         *string // -- None --, Unknown, Malicious, Suspicious, Clean
	Notes           *string // free-form notes (editable text, not work notes)
}

// CaseObservable row.
type CaseObservable struct {
	ID               uuid.UUID
	CaseID           uuid.UUID
	ObservableID     uuid.UUID
	RoleInCase       *string
	TrackingStatus   *string
	IsPrimary        bool
	Accuracy         *string
	Determination    *string
	Impact           *string
	AddedByUserID    *uuid.UUID
	AddedAt          time.Time
	UpdatedAt        time.Time
}

// CaseObservableWithDetails is a case observable with joined observable type/value and incident count for display.
type CaseObservableWithDetails struct {
	CaseObservable
	ObservableType   string
	ObservableValue  string
	NormalizedValue  string
	IncidentCount    int
}

// ChildObservable row.
type ChildObservable struct {
	ID                   uuid.UUID
	ParentObservableID   uuid.UUID
	ChildObservableID    uuid.UUID
	RelationshipType     string
	RelationshipDirection *string
	Confidence           *float64
	SourceName           *string
	SourceRecordID       *string
	Metadata             []byte // JSONB
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// SimilarIncident row.
type SimilarIncident struct {
	ID                     uuid.UUID
	CaseID                 uuid.UUID
	SimilarCaseID          uuid.UUID
	MatchReason            string
	SharedObservableCount  int
	SharedObservableIDs    []byte // JSONB array of UUIDs
	SharedObservableValues []byte // JSONB optional
	SimilarityScore        *float64
	LastComputedAt         time.Time
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// ListOpts for pagination.
type ListOpts struct {
	PageSize  int32
	PageToken string // offset or cursor
}

// AlertRuleRepo interface.
type AlertRuleRepo interface {
	Create(ctx context.Context, r *AlertRule) error
	GetByID(ctx context.Context, id uuid.UUID) (*AlertRule, error)
}

// AlertRepo interface.
type AlertRepo interface {
	Create(ctx context.Context, a *Alert) error
	GetByID(ctx context.Context, id uuid.UUID) (*Alert, error)
	ListByCaseID(ctx context.Context, caseID uuid.UUID, opts ListOpts) ([]*Alert, string, error)
}

// ObservableRepo interface.
type ObservableRepo interface {
	Create(ctx context.Context, o *Observable) error
	GetByID(ctx context.Context, id uuid.UUID) (*Observable, error)
	GetByTypeAndNormalized(ctx context.Context, observableType, normalizedValue string) (*Observable, error)
	Update(ctx context.Context, o *Observable) error
	List(ctx context.Context, opts ListOpts, searchQuery *string) ([]*Observable, string, error)
}

// CaseObservableRepo interface.
type CaseObservableRepo interface {
	Create(ctx context.Context, co *CaseObservable) error
	Update(ctx context.Context, co *CaseObservable) error
	GetByID(ctx context.Context, id uuid.UUID) (*CaseObservable, error)
	GetByCaseAndObservable(ctx context.Context, caseID, observableID uuid.UUID) (*CaseObservable, error)
	ListByCaseID(ctx context.Context, caseID uuid.UUID, opts ListOpts, filterType, filterStatus *string) ([]*CaseObservable, string, error)
	// ListCaseObservablesWithDetailsByCaseID returns case observables with observable type/value from JOIN.
	ListCaseObservablesWithDetailsByCaseID(ctx context.Context, caseID uuid.UUID, opts ListOpts, filterType, filterStatus *string) ([]*CaseObservableWithDetails, string, error)
	// CaseIDsByObservable returns case IDs that have this observable linked.
	CaseIDsByObservable(ctx context.Context, observableID uuid.UUID, opts ListOpts) ([]uuid.UUID, string, error)
	// ObservablesByCaseID returns observable IDs linked to the given case (for similar-incident computation).
	ObservableIDsByCaseID(ctx context.Context, caseID uuid.UUID) ([]uuid.UUID, error)
	// DeleteByCaseAndObservable removes the case_observable link.
	DeleteByCaseAndObservable(ctx context.Context, caseID, observableID uuid.UUID) error
}

// ChildObservableRepo interface.
type ChildObservableRepo interface {
	Create(ctx context.Context, c *ChildObservable) error
	GetByParentChildType(ctx context.Context, parentID, childID uuid.UUID, relationshipType string) (*ChildObservable, error)
	ListByParentID(ctx context.Context, parentID uuid.UUID, opts ListOpts) ([]*ChildObservable, string, error)
}

// SimilarIncidentRepo interface.
type SimilarIncidentRepo interface {
	Upsert(ctx context.Context, s *SimilarIncident) error
	DeleteByCaseAndSimilarCase(ctx context.Context, caseID, similarCaseID uuid.UUID) error
	ListByCaseID(ctx context.Context, caseID uuid.UUID, opts ListOpts) ([]*SimilarIncident, string, error)
}

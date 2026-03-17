package clients

import (
	"context"
	"time"
)

// CaseCommandService is the gRPC client contract for case write operations.
// Implement with generated stub when Case Service proto is vendored.
type CaseCommandService interface {
	CreateCase(ctx context.Context, req *CreateCaseRequest) (*Case, error)
	UpdateCase(ctx context.Context, caseID string, req *UpdateCaseRequest) (*Case, error)
	AddWorknote(ctx context.Context, caseID string, req *AddWorknoteRequest) (*Worknote, error)
	AssignCase(ctx context.Context, caseID string, req *AssignCaseRequest) error
	CloseCase(ctx context.Context, caseID string, req *CloseCaseRequest) error
	LinkObservable(ctx context.Context, caseID string, req *LinkObservableRequest) error
}

// CaseQueryService is the gRPC client contract for case read operations.
type CaseQueryService interface {
	GetCase(ctx context.Context, caseID string) (*Case, error)
	ListCases(ctx context.Context, pageSize int32, pageToken string) ([]*Case, string, error)
	ListWorknotes(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*Worknote, string, error)
}

// CreateCaseRequest for CreateCase.
type CreateCaseRequest struct {
	Title                string
	Description          string
	Priority             string
	Severity             string
	AffectedUserID       string
	AssignedUserID       string
	AssignmentGroupID    string
	FollowupTime         *time.Time
	Category             string
	Subcategory          string
	Source               string
	SourceTool           string
	SourceToolFeature    string
	ConfigurationItem    string
	SOCNotes             string
	NextSteps            string
	CSIRTClassification  string
	SOCLeadUserID        string
	NotificationTime     *time.Time
	IsAffectedUserVIP    bool
	RequestedByUserID    string
	EnvironmentLevel     string
	EnvironmentType      string
	PDN                  string
	ImpactedObject       string
	MTTR                 string
	EngineeringDocument  string
	ResponseDocument     string
}

// UpdateCaseRequest for UpdateCase (PATCH).
type UpdateCaseRequest struct {
	Title                *string
	Description          *string
	State                *string
	Priority             *string
	Severity             *string
	AffectedUserID       *string
	AssignedUserID       *string
	AssignmentGroupID    *string
	FollowupTime         *time.Time
	EventOccurredTime    *time.Time
	Category             *string
	Subcategory          *string
	Source               *string
	SourceTool           *string
	SourceToolFeature    *string
	ConfigurationItem    *string
	SOCNotes             *string
	NextSteps            *string
	CSIRTClassification  *string
	SOCLeadUserID        *string
	NotificationTime     *time.Time
	IsAffectedUserVIP    *bool
	RequestedByUserID    *string
	EnvironmentLevel     *string
	EnvironmentType      *string
	PDN                  *string
	ImpactedObject       *string
	MTTR                 *string
	ReassignmentCount    *int32
	AssignedToCount      *int32
	EngineeringDocument  *string
	ResponseDocument     *string
	Accuracy             *string
	Determination        *string
	ClosureReason        *string
}

// AddWorknoteRequest for AddWorknote.
type AddWorknoteRequest struct {
	Content string
}

// AssignCaseRequest for AssignCase.
type AssignCaseRequest struct {
	AssignedUserID    string
	AssignmentGroupID string
}

// CloseCaseRequest for CloseCase.
type CloseCaseRequest struct {
	Resolution string
}

// LinkObservableRequest for linking an observable to a case.
type LinkObservableRequest struct {
	ObservableID string
}

package clients

import "context"

// ObservableCommandService for observable write operations.
type ObservableCommandService interface {
	LinkObservableToCase(ctx context.Context, caseID, observableID string) error
	CreateAndLinkObservable(ctx context.Context, caseID, observableType, observableValue string) error
	CreateOrGetObservable(ctx context.Context, observableType, observableValue string) (*Observable, bool, error)
	UpdateObservable(ctx context.Context, observableID string, req *UpdateObservableRequest) (*Observable, error)
	UnlinkObservableFromCase(ctx context.Context, caseID, observableID string) error
}

// UpdateObservableRequest for PATCH observable (optional fields).
type UpdateObservableRequest struct {
	ObservableValue *string `json:"observable_value,omitempty"`
	ObservableType  *string `json:"observable_type,omitempty"`
	Finding         *string `json:"finding,omitempty"`
	Notes           *string `json:"notes,omitempty"`
}

// ObservableQueryService for observable read operations.
type ObservableQueryService interface {
	GetObservable(ctx context.Context, observableID string) (*Observable, error)
	ListObservables(ctx context.Context, search string, pageSize int32, pageToken string) ([]*Observable, string, error)
	ListCaseAlerts(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*Alert, string, error)
	ListCaseObservables(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*Observable, string, error)
	ListChildObservables(ctx context.Context, observableID string, pageSize int32, pageToken string) ([]*ChildObservable, string, error)
	ListSimilarIncidents(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*SimilarIncident, string, error)
}

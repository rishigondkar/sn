package orchestrator

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/servicenow/api-gateway/clients"
)

// failingCaseQuery returns error on GetCase to simulate required section failure
type failingCaseQuery struct{}

func (f *failingCaseQuery) GetCase(ctx context.Context, caseID string) (*clients.Case, error) {
	return nil, errBadRequest
}

func (f *failingCaseQuery) ListCases(ctx context.Context, pageSize int32, pageToken string) ([]*clients.Case, string, error) {
	return nil, "", nil
}

func (f *failingCaseQuery) ListWorknotes(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*clients.Worknote, string, error) {
	return nil, "", nil
}

// partialFailCaseQuery returns case but one of the list calls could fail in real scenario
type partialFailCaseQuery struct {
	worknotesErr error
}

func (p *partialFailCaseQuery) GetCase(ctx context.Context, caseID string) (*clients.Case, error) {
	return &clients.Case{ID: caseID, CaseNumber: "C-1", Title: "T", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}, nil
}

func (p *partialFailCaseQuery) ListCases(ctx context.Context, pageSize int32, pageToken string) ([]*clients.Case, string, error) {
	return nil, "", nil
}

func (p *partialFailCaseQuery) ListWorknotes(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*clients.Worknote, string, error) {
	if p.worknotesErr != nil {
		return nil, "", p.worknotesErr
	}
	return []*clients.Worknote{}, "", nil
}

type stubObsQuery struct{}
func (s *stubObsQuery) GetObservable(ctx context.Context, observableID string) (*clients.Observable, error) {
	return nil, nil
}
func (s *stubObsQuery) ListObservables(ctx context.Context, search string, pageSize int32, pageToken string) ([]*clients.Observable, string, error) {
	return nil, "", nil
}
func (s *stubObsQuery) ListCaseAlerts(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*clients.Alert, string, error) {
	return nil, "", nil
}
func (s *stubObsQuery) ListCaseObservables(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*clients.Observable, string, error) {
	return nil, "", nil
}
func (s *stubObsQuery) ListChildObservables(ctx context.Context, observableID string, pageSize int32, pageToken string) ([]*clients.ChildObservable, string, error) {
	return nil, "", nil
}
func (s *stubObsQuery) ListSimilarIncidents(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*clients.SimilarIncident, string, error) {
	return nil, "", nil
}

type stubEnrichment struct{}
func (s *stubEnrichment) ListEnrichmentResultsByCase(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*clients.EnrichmentResult, string, error) {
	return nil, "", nil
}

type stubThreatLookup struct{}
func (s *stubThreatLookup) ListThreatLookupResultsByCase(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*clients.ThreatLookupResult, string, error) {
	return nil, "", nil
}
func (s *stubThreatLookup) ListThreatLookupsByObservable(ctx context.Context, observableID string, pageSize int32, pageToken string) ([]*clients.ThreatLookupResult, string, error) {
	return nil, "", nil
}

type stubRef struct{}
func (s *stubRef) GetUser(ctx context.Context, userID string) (*clients.User, error) { return nil, nil }
func (s *stubRef) GetGroup(ctx context.Context, groupID string) (*clients.Group, error) { return nil, nil }
func (s *stubRef) ListUsers(ctx context.Context, pageSize int32, pageToken string) ([]*clients.User, string, error) {
	return nil, "", nil
}
func (s *stubRef) ListGroups(ctx context.Context, pageSize int32, pageToken string) ([]*clients.Group, string, error) {
	return nil, "", nil
}

type stubAttachment struct{}
func (s *stubAttachment) ListAttachmentsByCase(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*clients.Attachment, string, error) {
	return nil, "", nil
}

type stubAttachmentCmd struct{}
func (s *stubAttachmentCmd) CreateAttachment(ctx context.Context, req *clients.CreateAttachmentRequest) (*clients.Attachment, error) {
	return nil, nil
}

type stubAudit struct{}
func (s *stubAudit) ListAuditEventsByCase(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*clients.AuditEvent, string, error) {
	return nil, "", nil
}

type stubCaseCmd struct{}
func (s *stubCaseCmd) CreateCase(ctx context.Context, req *clients.CreateCaseRequest) (*clients.Case, error) { return nil, nil }
func (s *stubCaseCmd) UpdateCase(ctx context.Context, caseID string, req *clients.UpdateCaseRequest) (*clients.Case, error) { return nil, nil }
func (s *stubCaseCmd) AddWorknote(ctx context.Context, caseID string, req *clients.AddWorknoteRequest) (*clients.Worknote, error) { return nil, nil }
func (s *stubCaseCmd) AssignCase(ctx context.Context, caseID string, req *clients.AssignCaseRequest) error { return nil }
func (s *stubCaseCmd) CloseCase(ctx context.Context, caseID string, req *clients.CloseCaseRequest) error { return nil }
func (s *stubCaseCmd) LinkObservable(ctx context.Context, caseID string, req *clients.LinkObservableRequest) error { return nil }

type stubObsCmd struct{}
func (s *stubObsCmd) LinkObservableToCase(ctx context.Context, caseID, observableID string) error { return nil }
func (s *stubObsCmd) UnlinkObservableFromCase(ctx context.Context, caseID, observableID string) error { return nil }
func (s *stubObsCmd) CreateAndLinkObservable(ctx context.Context, caseID, observableType, observableValue string) error { return nil }
func (s *stubObsCmd) CreateOrGetObservable(ctx context.Context, observableType, observableValue string) (*clients.Observable, bool, error) {
	return &clients.Observable{ID: "stub-id", Type: observableType, Value: observableValue}, true, nil
}
func (s *stubObsCmd) UpdateObservable(ctx context.Context, observableID string, req *clients.UpdateObservableRequest) (*clients.Observable, error) {
	return nil, nil
}

func TestGetCaseDetail_RequiredFailure(t *testing.T) {
	orch := New(
		&stubCaseCmd{},
		&failingCaseQuery{},
		&stubObsCmd{},
		&stubObsQuery{},
		&stubEnrichment{},
		&stubThreatLookup{},
		&stubRef{},
		&stubAttachment{},
		&stubAttachmentCmd{},
		&stubAudit{},
	)
	_, err := orch.GetCaseDetail(context.Background(), "case-1", nil)
	if err != errBadRequest {
		t.Errorf("expected errBadRequest when GetCase fails, got %v", err)
	}
}

func TestGetCaseDetail_Success(t *testing.T) {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	orch := New(
		&stubCaseCmd{},
		&partialFailCaseQuery{},
		&stubObsCmd{},
		&stubObsQuery{},
		&stubEnrichment{},
		&stubThreatLookup{},
		&stubRef{},
		&stubAttachment{},
		&stubAttachmentCmd{},
		&stubAudit{},
	)
	detail, err := orch.GetCaseDetail(context.Background(), "case-1", log.Info)
	if err != nil {
		t.Fatal(err)
	}
	if detail.Case == nil || detail.Case.ID != "case-1" {
		t.Errorf("expected case id case-1, got %+v", detail.Case)
	}
}

func TestGetCaseDetail_EmptyCaseID(t *testing.T) {
	orch := New(
		&stubCaseCmd{},
		&partialFailCaseQuery{},
		&stubObsCmd{},
		&stubObsQuery{},
		&stubEnrichment{},
		&stubThreatLookup{},
		&stubRef{},
		&stubAttachment{},
		&stubAttachmentCmd{},
		&stubAudit{},
	)
	_, err := orch.GetCaseDetail(context.Background(), "", nil)
	if err != errBadRequest {
		t.Errorf("expected errBadRequest for empty caseId, got %v", err)
	}
}

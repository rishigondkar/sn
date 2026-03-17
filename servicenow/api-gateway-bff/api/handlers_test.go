package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/servicenow/api-gateway/clients"
	"github.com/servicenow/api-gateway/orchestrator"
)

// mockCaseCommand implements CaseCommandService for tests
type mockCaseCommand struct {
	createCaseErr error
	createCase    *clients.Case
}

func (m *mockCaseCommand) CreateCase(ctx context.Context, req *clients.CreateCaseRequest) (*clients.Case, error) {
	if m.createCaseErr != nil {
		return nil, m.createCaseErr
	}
	if m.createCase != nil {
		return m.createCase, nil
	}
	return &clients.Case{ID: "test-id", CaseNumber: "CASE-001", Title: req.Title}, nil
}

func (m *mockCaseCommand) UpdateCase(ctx context.Context, caseID string, req *clients.UpdateCaseRequest) (*clients.Case, error) {
	return nil, nil
}
func (m *mockCaseCommand) AddWorknote(ctx context.Context, caseID string, req *clients.AddWorknoteRequest) (*clients.Worknote, error) {
	return nil, nil
}
func (m *mockCaseCommand) AssignCase(ctx context.Context, caseID string, req *clients.AssignCaseRequest) error {
	return nil
}
func (m *mockCaseCommand) CloseCase(ctx context.Context, caseID string, req *clients.CloseCaseRequest) error {
	return nil
}
func (m *mockCaseCommand) LinkObservable(ctx context.Context, caseID string, req *clients.LinkObservableRequest) error {
	return nil
}

// mockCaseQuery implements CaseQueryService
type mockCaseQuery struct {
	getCaseErr error
	getCase    *clients.Case
}

func (m *mockCaseQuery) GetCase(ctx context.Context, caseID string) (*clients.Case, error) {
	if m.getCaseErr != nil {
		return nil, m.getCaseErr
	}
	if m.getCase != nil {
		return m.getCase, nil
	}
	return &clients.Case{ID: caseID, CaseNumber: "CASE-001", Title: "Test"}, nil
}

func (m *mockCaseQuery) ListCases(ctx context.Context, pageSize int32, pageToken string) ([]*clients.Case, string, error) {
	return nil, "", nil
}

func (m *mockCaseQuery) ListWorknotes(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*clients.Worknote, string, error) {
	return nil, "", nil
}

func stubOrchestrator() *orchestrator.Orchestrator {
	return orchestrator.New(
		&mockCaseCommand{},
		&mockCaseQuery{},
		&mockObservableCommand{},
		&mockObservableQuery{},
		&mockEnrichmentQuery{},
		&mockThreatLookupQuery{},
		&mockReferenceQuery{},
		&mockAttachmentQuery{},
		&mockAttachmentCommand{},
		&mockAuditQuery{},
	)
}

type mockObservableCommand struct{}
func (m *mockObservableCommand) LinkObservableToCase(ctx context.Context, caseID, observableID string) error { return nil }
func (m *mockObservableCommand) UnlinkObservableFromCase(ctx context.Context, caseID, observableID string) error { return nil }
func (m *mockObservableCommand) CreateAndLinkObservable(ctx context.Context, caseID, observableType, observableValue string) error { return nil }
func (m *mockObservableCommand) CreateOrGetObservable(ctx context.Context, observableType, observableValue string) (*clients.Observable, bool, error) {
	return &clients.Observable{ID: "mock-id", Type: observableType, Value: observableValue}, true, nil
}
func (m *mockObservableCommand) UpdateObservable(ctx context.Context, observableID string, req *clients.UpdateObservableRequest) (*clients.Observable, error) {
	return nil, nil
}

type mockObservableQuery struct{}
func (m *mockObservableQuery) GetObservable(ctx context.Context, observableID string) (*clients.Observable, error) {
	return nil, nil
}
func (m *mockObservableQuery) ListObservables(ctx context.Context, search string, pageSize int32, pageToken string) ([]*clients.Observable, string, error) {
	return nil, "", nil
}
func (m *mockObservableQuery) ListCaseAlerts(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*clients.Alert, string, error) {
	return nil, "", nil
}
func (m *mockObservableQuery) ListCaseObservables(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*clients.Observable, string, error) {
	return nil, "", nil
}
func (m *mockObservableQuery) ListChildObservables(ctx context.Context, observableID string, pageSize int32, pageToken string) ([]*clients.ChildObservable, string, error) {
	return nil, "", nil
}
func (m *mockObservableQuery) ListSimilarIncidents(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*clients.SimilarIncident, string, error) {
	return nil, "", nil
}

type mockEnrichmentQuery struct{}
func (m *mockEnrichmentQuery) ListEnrichmentResultsByCase(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*clients.EnrichmentResult, string, error) {
	return nil, "", nil
}

type mockThreatLookupQuery struct{}
func (m *mockThreatLookupQuery) ListThreatLookupResultsByCase(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*clients.ThreatLookupResult, string, error) {
	return nil, "", nil
}
func (m *mockThreatLookupQuery) ListThreatLookupsByObservable(ctx context.Context, observableID string, pageSize int32, pageToken string) ([]*clients.ThreatLookupResult, string, error) {
	return nil, "", nil
}

type mockReferenceQuery struct{}
func (m *mockReferenceQuery) GetUser(ctx context.Context, userID string) (*clients.User, error) { return nil, nil }
func (m *mockReferenceQuery) GetGroup(ctx context.Context, groupID string) (*clients.Group, error) { return nil, nil }
func (m *mockReferenceQuery) ListUsers(ctx context.Context, pageSize int32, pageToken string) ([]*clients.User, string, error) {
	return nil, "", nil
}
func (m *mockReferenceQuery) ListGroups(ctx context.Context, pageSize int32, pageToken string) ([]*clients.Group, string, error) {
	return nil, "", nil
}

type mockAttachmentQuery struct{}
func (m *mockAttachmentQuery) ListAttachmentsByCase(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*clients.Attachment, string, error) {
	return nil, "", nil
}

type mockAttachmentCommand struct{}
func (m *mockAttachmentCommand) CreateAttachment(ctx context.Context, req *clients.CreateAttachmentRequest) (*clients.Attachment, error) {
	return nil, nil
}

type mockAuditQuery struct{}
func (m *mockAuditQuery) ListAuditEventsByCase(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*clients.AuditEvent, string, error) {
	return nil, "", nil
}

func TestHealth(t *testing.T) {
	h := NewHandler(stubOrchestrator())
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("health: got status %d", rec.Code)
	}
	body := rec.Body.String()
	if body != `{"status":"ok"}` && body != `{"status":"ok"}`+"\n" {
		t.Errorf("health: got body %q", body)
	}
}

func TestCreateCase_Validation(t *testing.T) {
	h := NewHandler(stubOrchestrator())
	body := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cases", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("CreateCase empty title: got %d", rec.Code)
	}
	var errResp RESTError
	if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
		t.Fatal(err)
	}
	if errResp.Error.Code != CodeValidationError {
		t.Errorf("expected code %s, got %s", CodeValidationError, errResp.Error.Code)
	}
}

func TestCreateCase_Success(t *testing.T) {
	h := NewHandler(stubOrchestrator())
	body := bytes.NewBufferString(`{"title":"Test Case","priority":"P2"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cases", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Errorf("CreateCase success: got %d", rec.Code)
	}
	var c clients.Case
	if err := json.NewDecoder(rec.Body).Decode(&c); err != nil {
		t.Fatal(err)
	}
	if c.Title != "Test Case" {
		t.Errorf("expected title Test Case, got %s", c.Title)
	}
}

func TestGetCase_NotFound(t *testing.T) {
	getCaseErr := errors.New("not found")
	orch := orchestrator.New(
		&mockCaseCommand{},
		&mockCaseQuery{getCaseErr: getCaseErr},
		&mockObservableCommand{},
		&mockObservableQuery{},
		&mockEnrichmentQuery{},
		&mockThreatLookupQuery{},
		&mockReferenceQuery{},
		&mockAttachmentQuery{},
		&mockAttachmentCommand{},
		&mockAuditQuery{},
	)
	h := NewHandler(orch)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/cases/abc-123", nil)
	rec := httptest.NewRecorder()
	h.Router().ServeHTTP(rec, req)
	// Currently we map all errors to 500 unless they are errNotFound
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("GetCase downstream error: got %d", rec.Code)
	}
}

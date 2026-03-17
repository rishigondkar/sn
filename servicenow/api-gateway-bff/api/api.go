package api

import (
	"net/http"

	"github.com/servicenow/api-gateway/orchestrator"
)

// Handler holds the orchestrator and optional logger for request-scoped logging.
type Handler struct {
	Orch *orchestrator.Orchestrator
}

// NewHandler returns the API handler.
func NewHandler(orch *orchestrator.Orchestrator) *Handler {
	return &Handler{Orch: orch}
}

// Router returns the HTTP mux for /api/v1 and /health.
func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()

	// Health (no /api/v1 prefix)
	mux.HandleFunc("GET /health", h.Health)

	// API v1
	v1 := http.NewServeMux()
	v1.HandleFunc("POST /cases", h.CreateCase)
	v1.HandleFunc("GET /cases", h.ListCases)
	v1.HandleFunc("GET /cases/{caseId}", h.GetCase)
	v1.HandleFunc("PATCH /cases/{caseId}", h.UpdateCase)
	v1.HandleFunc("POST /cases/{caseId}/worknotes", h.AddWorknote)
	v1.HandleFunc("GET /cases/{caseId}/worknotes", h.ListWorknotes)
	v1.HandleFunc("POST /cases/{caseId}/assign", h.AssignCase)
	v1.HandleFunc("POST /cases/{caseId}/close", h.CloseCase)
	v1.HandleFunc("POST /cases/{caseId}/observables", h.LinkObservable)
	v1.HandleFunc("GET /cases/{caseId}/observables", h.ListCaseObservables)
	v1.HandleFunc("DELETE /cases/{caseId}/observables/{observableId}", h.UnlinkObservable)
	v1.HandleFunc("GET /cases/{caseId}/alerts", h.ListCaseAlerts)
	v1.HandleFunc("GET /cases/{caseId}/enrichment-results", h.ListEnrichmentResults)
	v1.HandleFunc("POST /enrichment-results", h.CreateEnrichmentResult)
	v1.HandleFunc("GET /cases/{caseId}/attachments", h.ListCaseAttachments)
	v1.HandleFunc("GET /cases/{caseId}/audit-events", h.ListCaseAuditEvents)
	v1.HandleFunc("GET /cases/{caseId}/detail", h.GetCaseDetail)

	v1.HandleFunc("GET /observables/classify", h.ClassifyObservable)
	v1.HandleFunc("GET /observables", h.ListObservables)
	v1.HandleFunc("POST /observables", h.CreateObservable)
	v1.HandleFunc("GET /observables/{observableId}", h.GetObservable)
	v1.HandleFunc("PATCH /observables/{observableId}", h.UpdateObservable)
	v1.HandleFunc("GET /observables/{observableId}/threat-lookups", h.ListThreatLookupsByObservable)

	v1.HandleFunc("POST /attachments", h.CreateAttachment)

	v1.HandleFunc("GET /reference/users", h.ListUsers)
	v1.HandleFunc("GET /reference/groups", h.ListGroups)

	v1.HandleFunc("POST /seed/presentation", h.SeedPresentation)

	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", v1))

	return mux
}

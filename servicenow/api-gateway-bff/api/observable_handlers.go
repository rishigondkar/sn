package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/org/alert-observable-service/service"
	"github.com/servicenow/api-gateway/clients"
)

func (h *Handler) LinkObservable(w http.ResponseWriter, r *http.Request) {
	caseID := r.PathValue("caseId")
	if caseID == "" {
		writeValidationError(w, r, "caseId", "required")
		return
	}
	var req struct {
		ObservableID    string `json:"observable_id"`
		ObservableType  string `json:"observable_type"`
		ObservableValue string `json:"observable_value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeValidationError(w, r, "body", "invalid JSON")
		return
	}
	obsType := strings.TrimSpace(req.ObservableType)
	obsValue := strings.TrimSpace(req.ObservableValue)
	if obsType != "" || obsValue != "" {
		if obsType == "" || obsValue == "" {
			writeValidationError(w, r, "observable_type and observable_value", "both required for create")
			return
		}
		obsType = strings.ToLower(obsType)
		err := h.Orch.ObsCmd.CreateAndLinkObservable(r.Context(), caseID, obsType, obsValue)
		if err != nil {
			writeDownstreamError(w, r, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if req.ObservableID == "" {
		writeValidationError(w, r, "observable_id", "required when observable_type/observable_value not provided")
		return
	}
	err := h.Orch.ObsCmd.LinkObservableToCase(r.Context(), caseID, strings.TrimSpace(req.ObservableID))
	if err != nil {
		writeDownstreamError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListObservables returns all observables with optional search (GET /observables?search=...&page_size=...&page_token=...).
func (h *Handler) ListObservables(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	pageSize, pageToken := parsePagination(r)
	list, nextToken, err := h.Orch.ObsQuery.ListObservables(r.Context(), search, pageSize, pageToken)
	if err != nil {
		writeDownstreamError(w, r, err)
		return
	}
	resp := struct {
		Observables   []*clients.Observable `json:"observables"`
		NextPageToken string                `json:"next_page_token,omitempty"`
	}{Observables: list, NextPageToken: nextToken}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// UnlinkObservable removes an observable from a case (DELETE /cases/{caseId}/observables/{observableId}).
func (h *Handler) UnlinkObservable(w http.ResponseWriter, r *http.Request) {
	caseID := r.PathValue("caseId")
	observableID := r.PathValue("observableId")
	if caseID == "" || observableID == "" {
		writeValidationError(w, r, "caseId and observableId", "required")
		return
	}
	if err := h.Orch.ObsCmd.UnlinkObservableFromCase(r.Context(), caseID, observableID); err != nil {
		writeDownstreamError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ClassifyObservable returns the suggested observable type and normalized value for a raw value (GET ?value=...).
func (h *Handler) ClassifyObservable(w http.ResponseWriter, r *http.Request) {
	value := strings.TrimSpace(r.URL.Query().Get("value"))
	if value == "" {
		writeValidationError(w, r, "value", "required")
		return
	}
	obsType, normalized := service.ClassifyObservableValue(value)
	resp := struct {
		ObservableType  string `json:"observable_type"`
		NormalizedValue string `json:"normalized_value"`
	}{ObservableType: obsType, NormalizedValue: normalized}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// CreateObservable creates a standalone observable (or returns existing by type+normalized value). POST body: { observable_type, observable_value }.
func (h *Handler) CreateObservable(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ObservableType  string `json:"observable_type"`
		ObservableValue string `json:"observable_value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeValidationError(w, r, "body", "invalid JSON")
		return
	}
	obsType := strings.TrimSpace(req.ObservableType)
	obsValue := strings.TrimSpace(req.ObservableValue)
	if obsType == "" || obsValue == "" {
		writeValidationError(w, r, "observable_type and observable_value", "required")
		return
	}
	obsType = strings.ToLower(obsType)
	observable, _, err := h.Orch.ObsCmd.CreateOrGetObservable(r.Context(), obsType, obsValue)
	if err != nil {
		writeDownstreamError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(observable)
}

func (h *Handler) ListCaseObservables(w http.ResponseWriter, r *http.Request) {
	caseID := r.PathValue("caseId")
	if caseID == "" {
		writeValidationError(w, r, "caseId", "required")
		return
	}
	pageSize, pageToken := parsePagination(r)
	list, nextToken, err := h.Orch.ObsQuery.ListCaseObservables(r.Context(), caseID, pageSize, pageToken)
	if err != nil {
		writeDownstreamError(w, r, err)
		return
	}
	resp := struct {
		Observables    []*clients.Observable `json:"observables"`
		NextPageToken string                 `json:"next_page_token,omitempty"`
	}{Observables: list, NextPageToken: nextToken}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) ListCaseAlerts(w http.ResponseWriter, r *http.Request) {
	caseID := r.PathValue("caseId")
	if caseID == "" {
		writeValidationError(w, r, "caseId", "required")
		return
	}
	pageSize, pageToken := parsePagination(r)
	list, nextToken, err := h.Orch.ObsQuery.ListCaseAlerts(r.Context(), caseID, pageSize, pageToken)
	if err != nil {
		writeDownstreamError(w, r, err)
		return
	}
	resp := struct {
		Alerts         []*clients.Alert `json:"alerts"`
		NextPageToken  string          `json:"next_page_token,omitempty"`
	}{Alerts: list, NextPageToken: nextToken}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) ListEnrichmentResults(w http.ResponseWriter, r *http.Request) {
	caseID := r.PathValue("caseId")
	if caseID == "" {
		writeValidationError(w, r, "caseId", "required")
		return
	}
	pageSize, pageToken := parsePagination(r)
	list, nextToken, err := h.Orch.Enrichment.ListEnrichmentResultsByCase(r.Context(), caseID, pageSize, pageToken)
	if err != nil {
		writeDownstreamError(w, r, err)
		return
	}
	resp := struct {
		EnrichmentResults []*clients.EnrichmentResult `json:"enrichment_results"`
		NextPageToken     string                      `json:"next_page_token,omitempty"`
	}{EnrichmentResults: list, NextPageToken: nextToken}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// CreateEnrichmentResult proxies an enrichment result creation request to the enrichment-threat-service REST API.
// It expects the request body to match enrichment-threat-service's POST /api/v1/enrichment-results shape.
func (h *Handler) CreateEnrichmentResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Read body once so we can forward it.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeDownstreamError(w, r, err)
		return
	}
	_ = r.Body.Close()

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, "http://localhost:8086/api/v1/enrichment-results", bytes.NewReader(body))
	if err != nil {
		writeDownstreamError(w, r, err)
		return
	}
	// Preserve content type and tracing/identity headers.
	if ct := r.Header.Get("Content-Type"); ct != "" {
		req.Header.Set("Content-Type", ct)
	} else {
		req.Header.Set("Content-Type", "application/json")
	}
	if v := r.Header.Get("X-User-Id"); v != "" {
		req.Header.Set("X-User-Id", v)
	}
	if v := r.Header.Get("X-Request-Id"); v != "" {
		req.Header.Set("X-Request-Id", v)
	}
	if v := r.Header.Get("X-Request-ID"); v != "" {
		req.Header.Set("X-Request-ID", v)
	}
	if v := r.Header.Get("X-Correlation-Id"); v != "" {
		req.Header.Set("X-Correlation-Id", v)
	}
	if v := r.Header.Get("X-Correlation-ID"); v != "" {
		req.Header.Set("X-Correlation-ID", v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		writeDownstreamError(w, r, err)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

// GetObservable returns a single observable by ID (for the observable detail page).
func (h *Handler) GetObservable(w http.ResponseWriter, r *http.Request) {
	observableID := r.PathValue("observableId")
	if observableID == "" {
		writeValidationError(w, r, "observableId", "required")
		return
	}
	obs, err := h.Orch.ObsQuery.GetObservable(r.Context(), observableID)
	if err != nil {
		writeDownstreamError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(obs)
}

// UpdateObservable updates an observable (PATCH). Body: { observable_value?, observable_type?, finding?, notes? }.
func (h *Handler) UpdateObservable(w http.ResponseWriter, r *http.Request) {
	observableID := r.PathValue("observableId")
	if observableID == "" {
		writeValidationError(w, r, "observableId", "required")
		return
	}
	var body clients.UpdateObservableRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeValidationError(w, r, "body", "invalid JSON")
		return
	}
	obs, err := h.Orch.ObsCmd.UpdateObservable(r.Context(), observableID, &body)
	if err != nil {
		writeDownstreamError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(obs)
}

func (h *Handler) ListThreatLookupsByObservable(w http.ResponseWriter, r *http.Request) {
	observableID := r.PathValue("observableId")
	if observableID == "" {
		writeValidationError(w, r, "observableId", "required")
		return
	}
	pageSize, pageToken := parsePagination(r)
	list, nextToken, err := h.Orch.ThreatLookup.ListThreatLookupsByObservable(r.Context(), observableID, pageSize, pageToken)
	if err != nil {
		writeDownstreamError(w, r, err)
		return
	}
	resp := struct {
		ThreatLookups  []*clients.ThreatLookupResult `json:"threat_lookups"`
		NextPageToken  string                         `json:"next_page_token,omitempty"`
	}{ThreatLookups: list, NextPageToken: nextToken}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

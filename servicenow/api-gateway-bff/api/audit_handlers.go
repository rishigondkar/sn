package api

import (
	"encoding/json"
	"net/http"

	"github.com/servicenow/api-gateway/clients"
)

func (h *Handler) ListCaseAuditEvents(w http.ResponseWriter, r *http.Request) {
	caseID := r.PathValue("caseId")
	if caseID == "" {
		writeValidationError(w, r, "caseId", "required")
		return
	}
	pageSize, pageToken := parsePagination(r)
	list, nextToken, err := h.Orch.Audit.ListAuditEventsByCase(r.Context(), caseID, pageSize, pageToken)
	if err != nil {
		writeDownstreamError(w, r, err)
		return
	}
	resp := struct {
		AuditEvents   []*clients.AuditEvent `json:"audit_events"`
		NextPageToken string                `json:"next_page_token,omitempty"`
	}{AuditEvents: list, NextPageToken: nextToken}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

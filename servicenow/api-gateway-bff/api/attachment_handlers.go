package api

import (
	"encoding/json"
	"net/http"

	"github.com/servicenow/api-gateway/clients"
)

func (h *Handler) CreateAttachment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CaseID      string `json:"case_id"`
		FileName    string `json:"file_name"`
		SizeBytes   int64  `json:"size_bytes"`
		ContentType string `json:"content_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeValidationError(w, r, "body", "invalid JSON")
		return
	}
	if req.CaseID == "" || req.FileName == "" {
		writeValidationError(w, r, "case_id or file_name", "required")
		return
	}
	att, err := h.Orch.AttachmentCmd.CreateAttachment(r.Context(), &clients.CreateAttachmentRequest{
		CaseID:      req.CaseID,
		FileName:    req.FileName,
		SizeBytes:   req.SizeBytes,
		ContentType: req.ContentType,
	})
	if err != nil {
		writeDownstreamError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(att)
}

func (h *Handler) ListCaseAttachments(w http.ResponseWriter, r *http.Request) {
	caseID := r.PathValue("caseId")
	if caseID == "" {
		writeValidationError(w, r, "caseId", "required")
		return
	}
	pageSize, pageToken := parsePagination(r)
	list, nextToken, err := h.Orch.Attachment.ListAttachmentsByCase(r.Context(), caseID, pageSize, pageToken)
	if err != nil {
		writeDownstreamError(w, r, err)
		return
	}
	resp := struct {
		Attachments   []*clients.Attachment `json:"attachments"`
		NextPageToken string                `json:"next_page_token,omitempty"`
	}{Attachments: list, NextPageToken: nextToken}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

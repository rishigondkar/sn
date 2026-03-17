package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/soc-platform/attachment-service/domain"
	"github.com/soc-platform/attachment-service/service"
)

// Health returns 200 when the process is up.
func (a *API) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// CreateAttachmentRequest is the JSON body for POST /api/v1/attachments.
type CreateAttachmentRequest struct {
	CaseID           string `json:"case_id"`
	FileName         string `json:"file_name"`
	ContentType      string `json:"content_type"`
	UploadedByUserID string `json:"uploaded_by_user_id"`
	Content          []byte `json:"content"` // base64 when sent as JSON; or use multipart in real impl
}

// CreateAttachment handles POST /api/v1/attachments.
func (a *API) CreateAttachment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}
	requestID := r.Header.Get("X-Request-Id")
	correlationID := r.Header.Get("X-Correlation-Id")
	actorName := r.Header.Get("X-Actor-Name")

	var req CreateAttachmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid JSON body")
		return
	}
	if req.UploadedByUserID == "" {
		req.UploadedByUserID = r.Header.Get("X-User-Id")
	}

	att, err := a.Service.CreateAttachment(r.Context(),
		req.CaseID, req.FileName, req.ContentType, req.UploadedByUserID, req.Content,
		requestID, correlationID, actorName,
	)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         att.ID,
		"case_id":    att.CaseID,
		"file_name":  att.FileName,
		"file_size_bytes": att.FileSizeBytes,
		"content_type": att.ContentType,
		"uploaded_by_user_id": att.UploadedByUserID,
		"uploaded_at": att.UploadedAt.Format("2006-01-02T15:04:05Z07:00"),
		"created_at":  att.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// ListAttachmentsByCase handles GET /api/v1/cases/{caseId}/attachments.
func (a *API) ListAttachmentsByCase(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}
	caseID := r.PathValue("caseId")
	if caseID == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "caseId required")
		return
	}
	pageSize := int32(50)
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if n, err := strconv.Atoi(ps); err == nil && n > 0 {
			if n > 100 {
				n = 100
			}
			pageSize = int32(n)
		}
	}
	pageToken := r.URL.Query().Get("page_token")
	includeDeleted := r.URL.Query().Get("include_deleted") == "true"

	items, nextToken, err := a.Service.ListAttachmentsByCase(r.Context(), caseID, pageSize, pageToken, includeDeleted)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	out := make([]map[string]interface{}, len(items))
	for i, att := range items {
		out[i] = attachmentToMap(att)
	}
	res := map[string]interface{}{"items": out}
	if nextToken != "" {
		res["next_page_token"] = nextToken
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(res)
}

// DeleteAttachment handles DELETE /api/v1/attachments/{attachmentId}.
func (a *API) DeleteAttachment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}
	attachmentID := r.PathValue("attachmentId")
	if attachmentID == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "attachmentId required")
		return
	}
	userID := r.Header.Get("X-User-Id")
	requestID := r.Header.Get("X-Request-Id")
	correlationID := r.Header.Get("X-Correlation-Id")
	actorName := r.Header.Get("X-Actor-Name")

	err := a.Service.DeleteAttachment(r.Context(), attachmentID, requestID, correlationID, userID, actorName)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func attachmentToMap(att *domain.Attachment) map[string]interface{} {
	m := map[string]interface{}{
		"id":                   att.ID,
		"case_id":              att.CaseID,
		"file_name":            att.FileName,
		"file_size_bytes":      att.FileSizeBytes,
		"content_type":         att.ContentType,
		"storage_provider":     att.StorageProvider,
		"uploaded_by_user_id":  att.UploadedByUserID,
		"uploaded_at":         att.UploadedAt.Format("2006-01-02T15:04:05Z07:00"),
		"is_deleted":          att.IsDeleted,
		"created_at":          att.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"updated_at":          att.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if att.DeletedAt != nil {
		m["deleted_at"] = att.DeletedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	if att.MetadataJSON != "" {
		m["metadata_json"] = att.MetadataJSON
	}
	return m
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch e := err.(type) {
	case service.ErrValidation:
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", e.Error())
		return
	}
	if err == service.ErrNotFound {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "attachment not found")
		return
	}
	writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
}

func writeError(w http.ResponseWriter, code int, errCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    errCode,
			"message": message,
		},
	})
}

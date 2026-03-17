package api

import (
	"encoding/json"
	"net/http"
)

// RESTError is the shared platform error envelope (00_platform_integration_contract).
type RESTError struct {
	Error struct {
		Code         string          `json:"code"`
		Message      string          `json:"message"`
		Details      []DetailItem    `json:"details,omitempty"`
		RequestID    string          `json:"requestId,omitempty"`
		CorrelationID string         `json:"correlationId,omitempty"`
	} `json:"error"`
}

type DetailItem struct {
	Field string `json:"field"`
	Issue string `json:"issue"`
}

// WriteError writes the REST error envelope and sets status code.
func WriteError(w http.ResponseWriter, status int, code, message, requestID, correlationID string, details []DetailItem) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(RESTError{
		Error: struct {
			Code          string       `json:"code"`
			Message       string       `json:"message"`
			Details       []DetailItem `json:"details,omitempty"`
			RequestID     string       `json:"requestId,omitempty"`
			CorrelationID string       `json:"correlationId,omitempty"`
		}{
			Code:          code,
			Message:       message,
			Details:       details,
			RequestID:     requestID,
			CorrelationID: correlationID,
		},
	})
}

// Error codes per platform contract
const (
	CodeValidationError   = "VALIDATION_ERROR"
	CodeNotFound          = "NOT_FOUND"
	CodeUnauthorized      = "UNAUTHORIZED"
	CodeForbidden         = "FORBIDDEN"
	CodeConflict          = "CONFLICT"
	CodeUnavailable       = "UNAVAILABLE"
	CodeInternal          = "INTERNAL_ERROR"
)

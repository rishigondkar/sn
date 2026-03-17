package api

import (
	"encoding/json"
	"net/http"
)

// RESTError is the platform contract error response shape.
type RESTError struct {
	Error struct {
		Code          string        `json:"code"`
		Message       string        `json:"message"`
		Details       []DetailEntry `json:"details,omitempty"`
		RequestID     string        `json:"requestId,omitempty"`
		CorrelationID string        `json:"correlationId,omitempty"`
	} `json:"error"`
}

type DetailEntry struct {
	Field string `json:"field"`
	Issue string `json:"issue"`
}

func WriteError(w http.ResponseWriter, code, message, requestID, correlationID string, details []DetailEntry, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(RESTError{
		Error: struct {
			Code          string        `json:"code"`
			Message       string        `json:"message"`
			Details       []DetailEntry `json:"details,omitempty"`
			RequestID     string        `json:"requestId,omitempty"`
			CorrelationID string        `json:"correlationId,omitempty"`
		}{
			Code:          code,
			Message:       message,
			Details:       details,
			RequestID:     requestID,
			CorrelationID: correlationID,
		},
	})
}

func WriteValidationError(w http.ResponseWriter, message string, details []DetailEntry, requestID, correlationID string) {
	WriteError(w, "VALIDATION_ERROR", message, requestID, correlationID, details, http.StatusBadRequest)
}

func WriteNotFound(w http.ResponseWriter, message, requestID, correlationID string) {
	WriteError(w, "NOT_FOUND", message, requestID, correlationID, nil, http.StatusNotFound)
}

func WriteInternal(w http.ResponseWriter, requestID, correlationID string) {
	WriteError(w, "INTERNAL_ERROR", "An internal error occurred", requestID, correlationID, nil, http.StatusInternalServerError)
}

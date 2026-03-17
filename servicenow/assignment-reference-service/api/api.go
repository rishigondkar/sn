package api

import (
	"encoding/json"
	"net/http"

	"github.com/servicenow/assignment-reference-service/service"
)

// RESTError is the platform error response shape.
type RESTError struct {
	Error struct {
		Code         string          `json:"code"`
		Message      string          `json:"message"`
		Details      []RESTErrorDetail `json:"details,omitempty"`
		RequestID    string          `json:"requestId,omitempty"`
		CorrelationID string         `json:"correlationId,omitempty"`
	} `json:"error"`
}

type RESTErrorDetail struct {
	Field string `json:"field"`
	Issue string `json:"issue"`
}

// Server holds REST dependencies.
type Server struct {
	Service *service.Service
}

// NewServer creates an API server with the given service.
func NewServer(svc *service.Service) *Server {
	return &Server{Service: svc}
}

// Handler returns the HTTP handler for REST + health. Health is at GET /health.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /api/v1/reference/users", s.listUsers)
	mux.HandleFunc("GET /api/v1/reference/groups", s.listGroups)
	mux.HandleFunc("GET /api/v1/reference/groups/{groupId}/members", s.listGroupMembers)
	return mux
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func writeRESTError(w http.ResponseWriter, code, message, requestID, correlationID string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(RESTError{
		Error: struct {
			Code          string            `json:"code"`
			Message       string            `json:"message"`
			Details       []RESTErrorDetail `json:"details,omitempty"`
			RequestID     string            `json:"requestId,omitempty"`
			CorrelationID string            `json:"correlationId,omitempty"`
		}{
			Code:          code,
			Message:       message,
			RequestID:     requestID,
			CorrelationID: correlationID,
		},
	})
}

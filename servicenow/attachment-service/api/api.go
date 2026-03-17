package api

import (
	"net/http"

	"github.com/soc-platform/attachment-service/service"
)

// API holds REST dependencies.
type API struct {
	Service *service.Service
}

// New returns a new API.
func New(svc *service.Service) *API {
	return &API{Service: svc}
}

// Handler returns the HTTP router for REST and health.
func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()
	// Health (required)
	mux.HandleFunc("GET /health", a.Health)
	// REST under /api/v1
	mux.HandleFunc("POST /api/v1/attachments", a.CreateAttachment)
	mux.HandleFunc("GET /api/v1/cases/{caseId}/attachments", a.ListAttachmentsByCase)
	mux.HandleFunc("DELETE /api/v1/attachments/{attachmentId}", a.DeleteAttachment)
	return mux
}

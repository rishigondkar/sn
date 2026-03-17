package api

import (
	"net/http"

	"github.com/servicenow/case-service/service"
)

// API provides HTTP handlers for health and optional REST.
type API struct {
	Service *service.Service
}

// New creates an API.
func New(svc *service.Service) *API {
	return &API{Service: svc}
}

// Handler returns the root http.Handler with /api/v1 and /health.
func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", a.Health)
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", a.v1Routes()))
	return mux
}

func (a *API) v1Routes() http.Handler {
	mux := http.NewServeMux()
	// Cases
	mux.HandleFunc("POST /cases", a.CreateCase)
	mux.HandleFunc("GET /cases/{id}", a.GetCase)
	mux.HandleFunc("GET /cases", a.ListCases)
	mux.HandleFunc("PATCH /cases/{id}", a.PatchCase)
	mux.HandleFunc("POST /cases/{id}/assign", a.AssignCase)
	mux.HandleFunc("POST /cases/{id}/close", a.CloseCase)
	// Worknotes
	mux.HandleFunc("POST /cases/{id}/worknotes", a.AddWorknote)
	mux.HandleFunc("GET /cases/{id}/worknotes", a.ListWorknotes)
	return mux
}

package api

import (
	"context"
	"net/http"
	"time"

	"github.com/servicenow/enrichment-threat-service/service"
)

// API holds REST dependencies.
type API struct {
	Service *service.Service
}

// NewAPI returns an API with the given Service.
func NewAPI(svc *service.Service) *API {
	return &API{Service: svc}
}

// Handler returns the HTTP handler for REST and health. Use with http.Server.
func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()

	// Health (platform: required)
	mux.HandleFunc("GET /health", a.Health)

	// Enrichment results
	mux.HandleFunc("POST /api/v1/enrichment-results", a.PostEnrichmentResult)
	mux.HandleFunc("PUT /api/v1/enrichment-results/{id}", a.PutEnrichmentResult)
	mux.HandleFunc("GET /api/v1/cases/{caseId}/enrichment-results", a.GetEnrichmentResultsByCase)
	mux.HandleFunc("GET /api/v1/observables/{observableId}/enrichment-results", a.GetEnrichmentResultsByObservable)

	// Threat lookups
	mux.HandleFunc("POST /api/v1/threat-lookups", a.PostThreatLookup)
	mux.HandleFunc("PUT /api/v1/threat-lookups/{id}", a.PutThreatLookup)
	mux.HandleFunc("GET /api/v1/cases/{caseId}/threat-lookups", a.GetThreatLookupsByCase)
	mux.HandleFunc("GET /api/v1/observables/{observableId}/threat-lookups", a.GetThreatLookupsByObservable)

	// Request ID / timeout middleware
	return requestIDMiddleware(timeoutMiddleware(mux, 30*time.Second))
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = r.Header.Get("X-Request-ID")
		}
		correlationID := r.Header.Get("X-Correlation-Id")
		if correlationID == "" {
			correlationID = r.Header.Get("X-Correlation-ID")
		}
		ctx := context.WithValue(r.Context(), ctxKeyRequestID, requestID)
		ctx = context.WithValue(ctx, ctxKeyCorrelationID, correlationID)
		ctx = context.WithValue(ctx, ctxKeyUserID, r.Header.Get("X-User-Id"))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type ctxKey string

const (
	ctxKeyRequestID    ctxKey = "request_id"
	ctxKeyCorrelationID ctxKey = "correlation_id"
	ctxKeyUserID       ctxKey = "user_id"
)

func timeoutMiddleware(next http.Handler, timeout time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *API) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

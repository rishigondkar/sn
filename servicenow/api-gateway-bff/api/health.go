package api

import (
	"net/http"
)

// Health returns 200 when the gateway is up. Optionally returns 503 if critical downstream is unreachable (not implemented in stub).
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

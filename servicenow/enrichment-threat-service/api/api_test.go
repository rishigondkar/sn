package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/servicenow/enrichment-threat-service/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealth(t *testing.T) {
	api := NewAPI(service.NewService(nil, nil, nil))
	handler := api.Handler()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "ok")
}

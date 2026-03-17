package handler

import (
	"github.com/servicenow/enrichment-threat-service/proto/enrichment_threat_service"
	"github.com/servicenow/enrichment-threat-service/service"
)

// Handler implements EnrichmentThreatService gRPC server.
type Handler struct {
	enrichment_threat_service.UnimplementedEnrichmentThreatServiceServer
	Service *service.Service
}

// NewHandler returns a Handler with the given Service.
func NewHandler(svc *service.Service) *Handler {
	return &Handler{Service: svc}
}

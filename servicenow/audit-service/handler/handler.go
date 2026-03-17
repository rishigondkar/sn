package handler

import (
	"github.com/servicenow/audit-service/proto/audit_service"
	"github.com/servicenow/audit-service/service"
)

const currentLayer = "handler"

// Handler implements the gRPC AuditService (query APIs only).
type Handler struct {
	audit_service.UnimplementedAuditServiceServer
	Service *service.Service
}

// NewHandler returns a new Handler.
func NewHandler(svc *service.Service) *Handler {
	return &Handler{Service: svc}
}

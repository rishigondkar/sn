package handler

import (
	"github.com/servicenow/case-service/proto/case_service"
	"github.com/servicenow/case-service/service"
)

const currentLayer = "handler"

// Handler implements CaseService gRPC server.
type Handler struct {
	case_service.UnimplementedCaseServiceServer
	Service *service.Service
}

// New creates a Handler.
func New(svc *service.Service) *Handler {
	return &Handler{Service: svc}
}

package handler

import (
	pb "github.com/servicenow/assignment-reference-service/proto/assignment_reference_service"
	"github.com/servicenow/assignment-reference-service/service"
)

// Handler implements AssignmentReferenceService gRPC server.
type Handler struct {
	pb.UnimplementedAssignmentReferenceServiceServer
	Service *service.Service
}

// New creates a gRPC handler with the given service.
func New(svc *service.Service) *Handler {
	return &Handler{Service: svc}
}

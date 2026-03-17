package handler

import (
	"context"

	"github.com/org/alert-observable-service/service"
	pb "github.com/org/alert-observable-service/proto/alert_observable_service"
	"google.golang.org/grpc/metadata"
)

// Server implements AlertObservableService gRPC server.
type Server struct {
	pb.UnimplementedAlertObservableServiceServer
	svc *service.Service
}

// NewServer returns a new gRPC server that delegates to the service.
func NewServer(svc *service.Service) *Server {
	return &Server{svc: svc}
}

// actorFromContext extracts x-user-id, x-request-id, x-correlation-id from gRPC metadata.
func actorFromContext(ctx context.Context) service.Actor {
	md, _ := metadata.FromIncomingContext(ctx)
	get := func(key string) string {
		if v := md.Get(key); len(v) > 0 {
			return v[0]
		}
		return ""
	}
	return service.Actor{
		UserID:        get("x-user-id"),
		UserName:      get("x-user-name"),
		RequestID:     get("x-request-id"),
		CorrelationID: get("x-correlation-id"),
	}
}

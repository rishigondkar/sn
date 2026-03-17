package handler

import (
	"context"

	"google.golang.org/grpc/metadata"

	"github.com/servicenow/assignment-reference-service/service"
)

// actorMetadataKeys per platform contract (gRPC).
const (
	metaUserID        = "x-user-id"
	metaRequestID    = "x-request-id"
	metaCorrelationID = "x-correlation-id"
)

// actorFromContext extracts actor and tracing metadata from gRPC context.
func actorFromContext(ctx context.Context) service.Actor {
	md, _ := metadata.FromIncomingContext(ctx)
	get := func(key string) string {
		if v := md.Get(key); len(v) > 0 {
			return v[0]
		}
		return ""
	}
	return service.Actor{
		UserID:        get(metaUserID),
		RequestID:    get(metaRequestID),
		CorrelationID: get(metaCorrelationID),
	}
}

package handler

import (
	"context"

	"google.golang.org/grpc/metadata"

	"github.com/servicenow/case-service/service"
)

// actorMetadata keys per platform contract.
const (
	metaUserID        = "x-user-id"
	metaRequestID     = "x-request-id"
	metaCorrelationID = "x-correlation-id"
)

func actorFromContext(ctx context.Context) *service.ActorMeta {
	meta := &service.ActorMeta{}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return meta
	}
	if v := md.Get(metaUserID); len(v) > 0 {
		meta.UserID = v[0]
	}
	if v := md.Get(metaRequestID); len(v) > 0 {
		meta.RequestID = v[0]
	}
	if v := md.Get(metaCorrelationID); len(v) > 0 {
		meta.CorrelationID = v[0]
	}
	return meta
}

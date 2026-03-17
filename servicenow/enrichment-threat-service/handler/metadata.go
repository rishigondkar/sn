package handler

import (
	"context"

	"github.com/servicenow/enrichment-threat-service/service"
	"google.golang.org/grpc/metadata"
)

func actorFromContext(ctx context.Context) service.ActorInfo {
	md, _ := metadata.FromIncomingContext(ctx)
	get := func(key string) string {
		if v := md.Get(key); len(v) > 0 {
			return v[0]
		}
		return ""
	}
	return service.ActorInfo{
		UserID:        get("x-user-id"),
		Name:          get("x-actor-name"),
		RequestID:     get("x-request-id"),
		CorrelationID: get("x-correlation-id"),
	}
}

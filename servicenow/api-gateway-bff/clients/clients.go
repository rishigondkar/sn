package clients

import "context"

// RequestContext carries actor and tracing metadata for downstream gRPC calls.
type RequestContext struct {
	ActorUserID   string
	ActorName     string
	RequestID     string
	CorrelationID string
}

// FromContext builds RequestContext from context values (set by middleware).
func FromContext(ctx context.Context) RequestContext {
	var rc RequestContext
	if v, ok := ctx.Value("actor_user_id").(string); ok {
		rc.ActorUserID = v
	}
	if v, ok := ctx.Value("actor_name").(string); ok {
		rc.ActorName = v
	}
	if v, ok := ctx.Value("request_id").(string); ok {
		rc.RequestID = v
	}
	if v, ok := ctx.Value("correlation_id").(string); ok {
		rc.CorrelationID = v
	}
	return rc
}

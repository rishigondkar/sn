package service

import "context"

// NoopAudit is a no-op AuditPublisher (e.g. when event bus is not configured).
type NoopAudit struct{}

func (NoopAudit) Publish(ctx context.Context, ev AuditEvent) error {
	return nil
}

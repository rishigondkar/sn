package service

import "context"

// StubAuditPublisher is a no-op publisher for development; replace with real event bus in production.
type StubAuditPublisher struct{}

func (s *StubAuditPublisher) Publish(ctx context.Context, event AuditEventEnvelope) error {
	return nil
}

package audit

import "context"

// Publisher publishes audit events (e.g. to event bus). After successful commit, call Publish.
// For transactional outbox, events are written to outbox table in same TX and relay publishes later.
type Publisher interface {
	Publish(ctx context.Context, evt *Event) error
}

// NoopPublisher does not send events; use when no bus is configured (e.g. local dev).
type NoopPublisher struct{}

func (NoopPublisher) Publish(ctx context.Context, evt *Event) error { return nil }

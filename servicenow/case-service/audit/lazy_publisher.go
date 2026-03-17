package audit

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LazyPublisher connects to the audit service on first Publish (or reconnects after failure).
// Use when the audit service may not be ready at case-service startup.
type LazyPublisher struct {
	addr   string
	mu     sync.Mutex
	client *GRPCPublisher
}

// NewLazyPublisher returns a publisher that dials addr on first use (or after reconnect).
func NewLazyPublisher(addr string) *LazyPublisher {
	return &LazyPublisher{addr: addr}
}

// auditPublishTimeout limits how long we wait for the audit service so we don't block forever.
const auditPublishTimeout = 10 * time.Second

// Publish sends the event. If not connected or connection failed, tries to connect then send.
// On connection error (e.g. dead connection after seed burst), clears the client and retries once with a new connection.
func (p *LazyPublisher) Publish(ctx context.Context, evt *Event) error {
	if evt == nil {
		return nil
	}
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		client, err := p.getOrConnect(ctx)
		if err != nil {
			lastErr = err
			slog.Default().ErrorContext(ctx, "audit lazy publish", slog.String("event_id", evt.EventID), slog.Int("attempt", attempt+1), slog.Any("error", err))
			continue
		}
		timeoutCtx, cancel := context.WithTimeout(ctx, auditPublishTimeout)
		err = client.Publish(timeoutCtx, evt)
		cancel()
		if err == nil {
			return nil
		}
		lastErr = err
		if isConnectionError(err) {
			p.mu.Lock()
			if p.client != nil {
				_ = p.client.Close()
				p.client = nil
			}
			p.mu.Unlock()
			slog.Default().WarnContext(ctx, "audit publish connection error, retrying with new connection", slog.String("event_id", evt.EventID), slog.Int("attempt", attempt+1), slog.Any("error", err))
			continue
		}
		slog.Default().ErrorContext(ctx, "audit publish", slog.String("event_id", evt.EventID), slog.Any("error", err))
		return err
	}
	slog.Default().ErrorContext(ctx, "audit publish failed after retry", slog.String("event_id", evt.EventID), slog.Any("error", lastErr))
	return lastErr
}

func (p *LazyPublisher) getOrConnect(ctx context.Context) (*GRPCPublisher, error) {
	p.mu.Lock()
	client := p.client
	p.mu.Unlock()
	if client != nil {
		return client, nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.client != nil {
		return p.client, nil
	}
	c, err := NewGRPCPublisher(ctx, p.addr)
	if err != nil {
		return nil, err
	}
	p.client = c
	return c, nil
}

func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	s, ok := status.FromError(err)
	if !ok {
		return true
	}
	switch s.Code() {
	case codes.Unavailable, codes.DeadlineExceeded, codes.Canceled:
		return true
	default:
		return false
	}
}

var _ Publisher = (*LazyPublisher)(nil)

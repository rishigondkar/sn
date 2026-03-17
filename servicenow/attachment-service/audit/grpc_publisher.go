package audit

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	pb "github.com/servicenow/audit-service/proto/audit_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GRPCPublisher publishes audit events to the Audit Service via gRPC. Connects lazily and reconnects on failure.
type GRPCPublisher struct {
	addr   string
	mu     sync.Mutex
	client pb.AuditServiceClient
	conn   *grpc.ClientConn
}

// NewGRPCPublisher returns a publisher that dials addr on first use.
func NewGRPCPublisher(addr string) *GRPCPublisher {
	return &GRPCPublisher{addr: addr}
}

const auditPublishTimeout = 10 * time.Second

// Publish sends the event to the Audit Service. Implements Publisher.
// On connection error, clears the client and retries once with a new connection.
func (p *GRPCPublisher) Publish(ctx context.Context, evt Event) error {
	if evt.EventID == "" {
		return nil
	}
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		client, err := p.getOrConnect(ctx)
		if err != nil {
			lastErr = err
			slog.Default().ErrorContext(ctx, "audit publish", slog.String("event_id", evt.EventID), slog.Int("attempt", attempt+1), slog.Any("error", err))
			continue
		}
		timeoutCtx, cancel := context.WithTimeout(ctx, auditPublishTimeout)
		req := eventToProto(evt)
		_, err = client.IngestEvent(timeoutCtx, &pb.IngestEventRequest{Event: req})
		cancel()
		if err == nil {
			return nil
		}
		lastErr = err
		if isConnError(err) {
			p.mu.Lock()
			if p.conn != nil {
				_ = p.conn.Close()
				p.conn, p.client = nil, nil
			}
			p.mu.Unlock()
			slog.Default().WarnContext(ctx, "audit ingest connection error, retrying", slog.String("event_id", evt.EventID), slog.Int("attempt", attempt+1), slog.Any("error", err))
			continue
		}
		slog.Default().ErrorContext(ctx, "audit ingest", slog.String("event_id", evt.EventID), slog.Any("error", err))
		return err
	}
	slog.Default().ErrorContext(ctx, "audit ingest failed after retry", slog.String("event_id", evt.EventID), slog.Any("error", lastErr))
	return lastErr
}

func (p *GRPCPublisher) getOrConnect(ctx context.Context) (pb.AuditServiceClient, error) {
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
	conn, err := grpc.NewClient(p.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	p.conn = conn
	p.client = pb.NewAuditServiceClient(conn)
	return p.client, nil
}

func eventToProto(e Event) *pb.AuditEvent {
	occurredAt := time.Now().UTC()
	if e.OccurredAt != "" {
		if t, err := time.Parse(time.RFC3339, e.OccurredAt); err == nil {
			occurredAt = t
		}
	}
	p := &pb.AuditEvent{
		EventId:       e.EventID,
		EventType:     e.EventType,
		SourceService: e.SourceService,
		EntityType:    e.EntityType,
		EntityId:      e.EntityID,
		Action:        e.Action,
		ActorUserId:   e.ActorUserID,
		OccurredAt:    timestamppb.New(occurredAt),
	}
	if e.ParentEntityType != "" {
		p.ParentEntityType = e.ParentEntityType
	}
	if e.ParentEntityID != "" {
		p.ParentEntityId = e.ParentEntityID
	}
	if e.CaseID != "" {
		p.CaseId = e.CaseID
	}
	if e.ActorName != "" {
		p.ActorName = e.ActorName
	}
	if e.RequestID != "" {
		p.RequestId = e.RequestID
	}
	if e.CorrelationID != "" {
		p.CorrelationId = e.CorrelationID
	}
	if e.ChangeSummary != "" {
		p.ChangeSummary = e.ChangeSummary
	}
	if len(e.AfterData) > 0 {
		b, _ := json.Marshal(e.AfterData)
		p.AfterDataJson = string(b)
	}
	if len(e.BeforeData) > 0 {
		b, _ := json.Marshal(e.BeforeData)
		p.BeforeDataJson = string(b)
	}
	if len(e.Metadata) > 0 {
		b, _ := json.Marshal(e.Metadata)
		p.MetadataJson = string(b)
	}
	return p
}

func isConnError(err error) bool {
	if err == nil {
		return false
	}
	s, ok := status.FromError(err)
	if !ok {
		return true
	}
	return s.Code() == codes.Unavailable || s.Code() == codes.DeadlineExceeded || s.Code() == codes.Canceled
}

var _ Publisher = (*GRPCPublisher)(nil)

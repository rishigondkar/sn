package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	pb "github.com/servicenow/audit-service/proto/audit_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GRPCAuditPublisher publishes audit events to the Audit Service via gRPC. Connects lazily and reconnects on failure.
// Implements AuditPublisher (Publish has no context; uses context.Background() for gRPC).
type GRPCAuditPublisher struct {
	addr   string
	mu     sync.Mutex
	client pb.AuditServiceClient
	conn   *grpc.ClientConn
}

// NewGRPCAuditPublisher returns a publisher that dials addr on first use.
func NewGRPCAuditPublisher(addr string) *GRPCAuditPublisher {
	return &GRPCAuditPublisher{addr: addr}
}

// Publish sends the event to the Audit Service. Implements AuditPublisher.
func (p *GRPCAuditPublisher) Publish(event AuditEvent) error {
	ctx := context.Background()
	client, err := p.getOrConnect(ctx)
	if err != nil {
		slog.Default().ErrorContext(ctx, "audit publish", slog.String("event_id", event.EventID), slog.Any("error", err))
		return err
	}
	req := auditEventToProto(event)
	_, err = client.IngestEvent(ctx, &pb.IngestEventRequest{Event: req})
	if err != nil {
		if isConnError(err) {
			p.mu.Lock()
			if p.conn != nil {
				_ = p.conn.Close()
				p.conn, p.client = nil, nil
			}
			p.mu.Unlock()
		}
		slog.Default().ErrorContext(ctx, "audit ingest", slog.String("event_id", event.EventID), slog.Any("error", err))
		return err
	}
	return nil
}

func (p *GRPCAuditPublisher) getOrConnect(ctx context.Context) (pb.AuditServiceClient, error) {
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

func auditEventToProto(e AuditEvent) *pb.AuditEvent {
	p := &pb.AuditEvent{
		EventId:       e.EventID,
		EventType:     e.EventType,
		SourceService: e.SourceService,
		EntityType:    e.EntityType,
		EntityId:      e.EntityID,
		Action:        e.Action,
		ActorUserId:   e.ActorUserID,
		OccurredAt:    timestamppb.New(e.OccurredAt),
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
	if e.ObservableID != "" {
		p.ObservableId = e.ObservableID
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

var _ AuditPublisher = (*GRPCAuditPublisher)(nil)

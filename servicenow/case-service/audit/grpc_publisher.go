package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	pb "github.com/servicenow/audit-service/proto/audit_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GRPCPublisher publishes audit events to the Audit Service via gRPC IngestEvent.
type GRPCPublisher struct {
	client pb.AuditServiceClient
	conn   *grpc.ClientConn
}

// NewGRPCPublisher dials the audit service at addr and returns a Publisher. Caller should close when done.
func NewGRPCPublisher(ctx context.Context, addr string) (*GRPCPublisher, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("audit service dial: %w", err)
	}
	client := pb.NewAuditServiceClient(conn)
	return &GRPCPublisher{client: client, conn: conn}, nil
}

// Close closes the gRPC connection.
func (p *GRPCPublisher) Close() error {
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

// Publish sends the event to the Audit Service. Implements audit.Publisher.
func (p *GRPCPublisher) Publish(ctx context.Context, evt *Event) error {
	if evt == nil {
		return nil
	}
	req := eventToProto(evt)
	_, err := p.client.IngestEvent(ctx, &pb.IngestEventRequest{Event: req})
	if err != nil {
		slog.Default().ErrorContext(ctx, "audit publish", slog.String("event_id", evt.EventID), slog.Any("error", err))
		return err
	}
	return nil
}

func eventToProto(e *Event) *pb.AuditEvent {
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
	if len(e.BeforeData) > 0 {
		b, _ := json.Marshal(e.BeforeData)
		p.BeforeDataJson = string(b)
	}
	if len(e.AfterData) > 0 {
		b, _ := json.Marshal(e.AfterData)
		p.AfterDataJson = string(b)
	}
	if len(e.Metadata) > 0 {
		b, _ := json.Marshal(e.Metadata)
		p.MetadataJson = string(b)
	}
	return p
}

var _ Publisher = (*GRPCPublisher)(nil)

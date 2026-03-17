package clients

import (
	"context"
	"fmt"
	"time"

	pb "github.com/servicenow/audit-service/proto/audit_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ AuditQueryService = (*auditQueryGRPC)(nil)

type auditQueryGRPC struct {
	client pb.AuditServiceClient
}

// NewAuditClients dials the Audit Service and returns an AuditQueryService.
func NewAuditClients(addr string, timeout time.Duration) (AuditQueryService, func(), error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("audit service dial: %w", err)
	}
	client := pb.NewAuditServiceClient(conn)
	impl := &auditQueryGRPC{client: client}
	closeFn := func() { _ = conn.Close() }
	return impl, closeFn, nil
}

func (c *auditQueryGRPC) ListAuditEventsByCase(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*AuditEvent, string, error) {
	if caseID == "" {
		return nil, "", fmt.Errorf("caseID is required")
	}
	ctx = withMetadata(ctx)
	if pageSize <= 0 {
		pageSize = 50
	}
	resp, err := c.client.ListAuditEventsByCase(ctx, &pb.ListAuditEventsByCaseRequest{
		CaseId:    caseID,
		PageSize:  pageSize,
		PageToken: pageToken,
	})
	if err != nil {
		return nil, "", err
	}
	list := make([]*AuditEvent, 0, len(resp.GetItems()))
	for _, e := range resp.GetItems() {
		list = append(list, protoAuditEventToAuditEvent(e))
	}
	return list, resp.GetNextPageToken(), nil
}

func protoAuditEventToAuditEvent(e *pb.AuditEvent) *AuditEvent {
	if e == nil {
		return nil
	}
	out := &AuditEvent{
		EventID:       e.GetEventId(),
		EventType:     e.GetEventType(),
		EntityType:    e.GetEntityType(),
		EntityID:      e.GetEntityId(),
		Action:        e.GetAction(),
		ActorUserID:   e.GetActorUserId(),
		ActorName:     e.GetActorName(),
		ChangeSummary: e.GetChangeSummary(),
	}
	if e.GetOccurredAt() != nil {
		out.OccurredAt = e.GetOccurredAt().AsTime()
	}
	return out
}

package handler

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"github.com/servicenow/audit-service/proto/audit_service"
	"github.com/servicenow/audit-service/service"
)

// TestListAuditEventsByCase_InvalidArgument tests that empty case_id returns InvalidArgument.
// Uses in-process server with nil repo (service will fail on actual list; we only care about validation).
func TestListAuditEventsByCase_InvalidArgument(t *testing.T) {
	// Service with nil repo - ListByCase will panic on nil. So we need a real repo or mock.
	// Use a real DB for integration: skip if no DB.
	// For unit test we only need handler to return InvalidArgument for empty case_id.
	// So: start server with minimal deps - repo can be nil and we only call with empty case_id.
	svc := service.NewService(nil)
	h := NewHandler(svc)

	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()
	audit_service.RegisterAuditServiceServer(srv, h)
	go func() {
		_ = srv.Serve(lis)
	}()
	t.Cleanup(srv.Stop)

	conn, err := grpc.NewClient("passthrough://bufnet", grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.Dial() }), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	client := audit_service.NewAuditServiceClient(conn)
	ctx := context.Background()

	// Empty case_id -> InvalidArgument
	_, err = client.ListAuditEventsByCase(ctx, &audit_service.ListAuditEventsByCaseRequest{CaseId: ""})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestListAuditEventsByEntity_InvalidArgument tests that missing entity_type or entity_id returns InvalidArgument.
func TestListAuditEventsByEntity_InvalidArgument(t *testing.T) {
	svc := service.NewService(nil)
	h := NewHandler(svc)
	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()
	audit_service.RegisterAuditServiceServer(srv, h)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	conn, err := grpc.NewClient("passthrough://bufnet", grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.Dial() }), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	client := audit_service.NewAuditServiceClient(conn)
	ctx := context.Background()

	_, err = client.ListAuditEventsByEntity(ctx, &audit_service.ListAuditEventsByEntityRequest{EntityType: "", EntityId: "id"})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = client.ListAuditEventsByEntity(ctx, &audit_service.ListAuditEventsByEntityRequest{EntityType: "case", EntityId: ""})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

package handler

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/soc-platform/attachment-service/audit"
	"github.com/soc-platform/attachment-service/domain"
	"github.com/soc-platform/attachment-service/service"
	"github.com/soc-platform/attachment-service/storage"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/soc-platform/attachment-service/proto/attachment_service"
)

func TestGRPC_CreateAttachment_List_Delete(t *testing.T) {
	repo := &mockRepoForHandler{get: make(map[string]*domain.Attachment)}
	store := storage.NewMemoryStore()
	aud := &mockAuditForHandler{}
	svc := &service.Service{
		Repo:    repo,
		Storage: store,
		Audit:   aud,
		MaxBytes: 1024,
	}
	h := &Handler{Service: svc}
	srv := grpc.NewServer()
	pb.RegisterAttachmentServiceServer(srv, h)
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	go srv.Serve(lis)
	defer srv.GracefulStop()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()
	client := pb.NewAttachmentServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create
	createResp, err := client.CreateAttachment(ctx, &pb.CreateAttachmentRequest{
		CaseId:           "case-1",
		FileName:         "test.txt",
		ContentType:      "text/plain",
		UploadedByUserId: "user-1",
		Content:          []byte("hello"),
	})
	require.NoError(t, err)
	require.NotEmpty(t, createResp.GetId())
	require.NotNil(t, createResp.GetAttachment())
	require.Equal(t, "test.txt", createResp.GetAttachment().GetFileName())

	// List
	listResp, err := client.ListAttachmentsByCase(ctx, &pb.ListAttachmentsByCaseRequest{CaseId: "case-1"})
	require.NoError(t, err)
	require.Len(t, listResp.GetItems(), 1)
	require.Equal(t, createResp.GetId(), listResp.GetItems()[0].GetId())

	// Delete
	_, err = client.DeleteAttachment(ctx, &pb.DeleteAttachmentRequest{AttachmentId: createResp.GetId()})
	require.NoError(t, err)

	// List again (exclude deleted by default) -> 0 items
	listResp2, err := client.ListAttachmentsByCase(ctx, &pb.ListAttachmentsByCaseRequest{CaseId: "case-1"})
	require.NoError(t, err)
	require.Len(t, listResp2.GetItems(), 0)
}

func TestGRPC_DeleteAttachment_NotFound(t *testing.T) {
	repo := &mockRepoForHandler{get: make(map[string]*domain.Attachment)}
	svc := &service.Service{Repo: repo, Storage: storage.NewMemoryStore(), Audit: &mockAuditForHandler{}}
	h := &Handler{Service: svc}
	srv := grpc.NewServer()
	pb.RegisterAttachmentServiceServer(srv, h)
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	go srv.Serve(lis)
	defer srv.GracefulStop()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()
	client := pb.NewAttachmentServiceClient(conn)
	ctx := context.Background()

	_, err = client.DeleteAttachment(ctx, &pb.DeleteAttachmentRequest{AttachmentId: "nonexistent"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "NotFound")
}

type mockRepoForHandler struct {
	get map[string]*domain.Attachment
}

func (m *mockRepoForHandler) CreateAttachment(ctx context.Context, a *domain.Attachment) error {
	if m.get == nil {
		m.get = make(map[string]*domain.Attachment)
	}
	dup := *a
	m.get[a.ID] = &dup
	return nil
}

func (m *mockRepoForHandler) GetByID(ctx context.Context, id string) (*domain.Attachment, error) {
	return m.get[id], nil
}

func (m *mockRepoForHandler) ListByCaseID(ctx context.Context, caseID string, pageSize int32, pageToken string, includeDeleted bool) ([]*domain.Attachment, string, error) {
	var out []*domain.Attachment
	for _, a := range m.get {
		if a.CaseID == caseID && (includeDeleted || !a.IsDeleted) {
			out = append(out, a)
		}
	}
	return out, "", nil
}

func (m *mockRepoForHandler) SoftDelete(ctx context.Context, id string, deletedAt time.Time) error {
	if a, ok := m.get[id]; ok {
		a.IsDeleted = true
		a.DeletedAt = &deletedAt
	}
	return nil
}

type mockAuditForHandler struct{}

func (mockAuditForHandler) Publish(ctx context.Context, evt audit.Event) error { return nil }

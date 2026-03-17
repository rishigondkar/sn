package clients

import (
	"context"
	"fmt"
	"time"

	pb "github.com/soc-platform/attachment-service/proto/attachment_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	_ AttachmentQueryService  = (*attachmentQueryGRPC)(nil)
	_ AttachmentCommandService = (*attachmentCommandGRPC)(nil)
)

type attachmentQueryGRPC struct {
	client pb.AttachmentServiceClient
}

type attachmentCommandGRPC struct {
	client pb.AttachmentServiceClient
}

// NewAttachmentClients dials the Attachment Service and returns query and command clients.
func NewAttachmentClients(addr string, timeout time.Duration) (AttachmentQueryService, AttachmentCommandService, func(), error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("attachment service dial: %w", err)
	}
	client := pb.NewAttachmentServiceClient(conn)
	q := &attachmentQueryGRPC{client: client}
	cmd := &attachmentCommandGRPC{client: client}
	closeFn := func() { _ = conn.Close() }
	return q, cmd, closeFn, nil
}

func (c *attachmentQueryGRPC) ListAttachmentsByCase(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*Attachment, string, error) {
	if caseID == "" {
		return nil, "", fmt.Errorf("caseID is required")
	}
	ctx = withMetadata(ctx)
	if pageSize <= 0 {
		pageSize = 50
	}
	resp, err := c.client.ListAttachmentsByCase(ctx, &pb.ListAttachmentsByCaseRequest{
		CaseId:         caseID,
		PageSize:       pageSize,
		PageToken:      pageToken,
		IncludeDeleted: false,
	})
	if err != nil {
		return nil, "", err
	}
	list := make([]*Attachment, 0, len(resp.GetItems()))
	for _, a := range resp.GetItems() {
		list = append(list, protoAttachmentToAttachment(a))
	}
	return list, resp.GetNextPageToken(), nil
}

func (c *attachmentCommandGRPC) CreateAttachment(ctx context.Context, req *CreateAttachmentRequest) (*Attachment, error) {
	if req == nil || req.CaseID == "" || req.FileName == "" {
		return nil, fmt.Errorf("CreateAttachmentRequest with CaseID and FileName is required")
	}
	ctx = withMetadata(ctx)
	rc := FromContext(ctx)
	resp, err := c.client.CreateAttachment(ctx, &pb.CreateAttachmentRequest{
		CaseId:             req.CaseID,
		FileName:           req.FileName,
		ContentType:        req.ContentType,
		UploadedByUserId:   rc.ActorUserID,
		Content:            nil, // metadata-only; upload may be separate
	})
	if err != nil {
		return nil, err
	}
	if resp.GetAttachment() != nil {
		return protoAttachmentToAttachment(resp.GetAttachment()), nil
	}
	return &Attachment{
		ID:        resp.GetId(),
		CaseID:    req.CaseID,
		FileName:  req.FileName,
		SizeBytes: req.SizeBytes,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func protoAttachmentToAttachment(a *pb.Attachment) *Attachment {
	if a == nil {
		return nil
	}
	out := &Attachment{
		ID:        a.GetId(),
		CaseID:    a.GetCaseId(),
		FileName:  a.GetFileName(),
		SizeBytes: a.GetFileSizeBytes(),
	}
	if a.GetCreatedAt() != nil {
		out.CreatedAt = a.GetCreatedAt().AsTime()
	}
	return out
}

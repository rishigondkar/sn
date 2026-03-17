package clients

import "context"

// AttachmentQueryService for attachment read operations.
// Upload is typically a separate flow (e.g. signed URL or direct POST to Attachment Service); gateway may proxy.
type AttachmentQueryService interface {
	ListAttachmentsByCase(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*Attachment, string, error)
}

// AttachmentCommandService for upload (if gateway proxies).
type AttachmentCommandService interface {
	CreateAttachment(ctx context.Context, req *CreateAttachmentRequest) (*Attachment, error)
}

// CreateAttachmentRequest for creating an attachment (metadata; file upload may be separate).
type CreateAttachmentRequest struct {
	CaseID     string
	FileName   string
	SizeBytes  int64
	ContentType string
}

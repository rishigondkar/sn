package clients

import "context"

// AuditQueryService for audit read operations.
type AuditQueryService interface {
	ListAuditEventsByCase(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*AuditEvent, string, error)
}

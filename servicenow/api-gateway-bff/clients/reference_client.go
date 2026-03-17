package clients

import "context"

// ReferenceQueryService for users and groups.
type ReferenceQueryService interface {
	GetUser(ctx context.Context, userID string) (*User, error)
	GetGroup(ctx context.Context, groupID string) (*Group, error)
	ListUsers(ctx context.Context, pageSize int32, pageToken string) ([]*User, string, error)
	ListGroups(ctx context.Context, pageSize int32, pageToken string) ([]*Group, string, error)
}

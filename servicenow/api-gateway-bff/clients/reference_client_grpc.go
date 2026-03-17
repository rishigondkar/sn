package clients

import (
	"context"
	"fmt"
	"time"

	pb "github.com/servicenow/assignment-reference-service/proto/assignment_reference_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ ReferenceQueryService = (*referenceQueryGRPC)(nil)

type referenceQueryGRPC struct {
	client pb.AssignmentReferenceServiceClient
}

// NewReferenceClients dials the Assignment Reference Service and returns a ReferenceQueryService.
func NewReferenceClients(addr string, timeout time.Duration) (ReferenceQueryService, func(), error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("reference service dial: %w", err)
	}
	client := pb.NewAssignmentReferenceServiceClient(conn)
	impl := &referenceQueryGRPC{client: client}
	closeFn := func() { _ = conn.Close() }
	return impl, closeFn, nil
}

func (c *referenceQueryGRPC) GetUser(ctx context.Context, userID string) (*User, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}
	ctx = withMetadata(ctx)
	resp, err := c.client.GetUser(ctx, &pb.GetUserRequest{Id: userID})
	if err != nil {
		return nil, err
	}
	return protoRefUserToUser(resp.GetUser()), nil
}

func (c *referenceQueryGRPC) GetGroup(ctx context.Context, groupID string) (*Group, error) {
	if groupID == "" {
		return nil, fmt.Errorf("groupID is required")
	}
	ctx = withMetadata(ctx)
	resp, err := c.client.GetGroup(ctx, &pb.GetGroupRequest{Id: groupID})
	if err != nil {
		return nil, err
	}
	return protoRefGroupToGroup(resp.GetGroup()), nil
}

func (c *referenceQueryGRPC) ListUsers(ctx context.Context, pageSize int32, pageToken string) ([]*User, string, error) {
	ctx = withMetadata(ctx)
	if pageSize <= 0 {
		pageSize = 100
	}
	resp, err := c.client.ListUsers(ctx, &pb.ListUsersRequest{
		PageSize:  pageSize,
		PageToken: pageToken,
	})
	if err != nil {
		return nil, "", err
	}
	list := make([]*User, 0, len(resp.GetItems()))
	for _, u := range resp.GetItems() {
		list = append(list, protoRefUserToUser(u))
	}
	return list, resp.GetNextPageToken(), nil
}

func (c *referenceQueryGRPC) ListGroups(ctx context.Context, pageSize int32, pageToken string) ([]*Group, string, error) {
	ctx = withMetadata(ctx)
	if pageSize <= 0 {
		pageSize = 100
	}
	resp, err := c.client.ListGroups(ctx, &pb.ListGroupsRequest{
		PageSize:  pageSize,
		PageToken: pageToken,
	})
	if err != nil {
		return nil, "", err
	}
	list := make([]*Group, 0, len(resp.GetItems()))
	for _, g := range resp.GetItems() {
		list = append(list, protoRefGroupToGroup(g))
	}
	return list, resp.GetNextPageToken(), nil
}

func protoRefUserToUser(u *pb.User) *User {
	if u == nil {
		return nil
	}
	return &User{
		ID:          u.GetId(),
		DisplayName: u.GetDisplayName(),
		Email:       u.GetEmail(),
	}
}

func protoRefGroupToGroup(g *pb.Group) *Group {
	if g == nil {
		return nil
	}
	name := g.GetGroupName()
	return &Group{ID: g.GetId(), Name: name}
}

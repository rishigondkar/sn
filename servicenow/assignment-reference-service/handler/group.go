package handler

import (
	"context"
	"log/slog"
	"time"

	pb "github.com/servicenow/assignment-reference-service/proto/assignment_reference_service"
)

// GetGroup returns a group by ID.
func (h *Handler) GetGroup(ctx context.Context, req *pb.GetGroupRequest) (*pb.GetGroupResponse, error) {
	start := time.Now()
	var err error
	defer func() {
		if err != nil {
			slog.Error("GetGroup failed", "error", err, "duration_ms", time.Since(start).Milliseconds())
		}
	}()

	g, err := h.Service.GetGroup(ctx, req.GetId())
	if err != nil {
		return nil, toStatus(err)
	}
	return &pb.GetGroupResponse{Group: groupToProto(g)}, nil
}

// ListGroups returns a page of groups with optional filters.
func (h *Handler) ListGroups(ctx context.Context, req *pb.ListGroupsRequest) (*pb.ListGroupsResponse, error) {
	start := time.Now()
	var err error
	defer func() {
		if err != nil {
			slog.Error("ListGroups failed", "error", err, "duration_ms", time.Since(start).Milliseconds())
		}
	}()

	items, nextToken, err := h.Service.ListGroups(ctx,
		req.GetPageSize(),
		req.GetPageToken(),
		req.ActiveOnly,
		req.FilterGroupName,
	)
	if err != nil {
		return nil, toStatus(err)
	}
	out := make([]*pb.Group, len(items))
	for i, g := range items {
		out[i] = groupToProto(g)
	}
	return &pb.ListGroupsResponse{Items: out, NextPageToken: nextToken}, nil
}

// ListGroupMembers returns a page of members for a group.
func (h *Handler) ListGroupMembers(ctx context.Context, req *pb.ListGroupMembersRequest) (*pb.ListGroupMembersResponse, error) {
	start := time.Now()
	var err error
	defer func() {
		if err != nil {
			slog.Error("ListGroupMembers failed", "error", err, "duration_ms", time.Since(start).Milliseconds())
		}
	}()

	items, nextToken, err := h.Service.ListGroupMembers(ctx,
		req.GetGroupId(),
		req.GetPageSize(),
		req.GetPageToken(),
	)
	if err != nil {
		return nil, toStatus(err)
	}
	out := make([]*pb.GroupMember, len(items))
	for i, m := range items {
		out[i] = groupMemberToProto(m)
	}
	return &pb.ListGroupMembersResponse{Items: out, NextPageToken: nextToken}, nil
}

// ValidateGroupExists checks if the group ID exists.
func (h *Handler) ValidateGroupExists(ctx context.Context, req *pb.ValidateGroupExistsRequest) (*pb.ValidateGroupExistsResponse, error) {
	exists, err := h.Service.ValidateGroupExists(ctx, req.GetGroupId())
	if err != nil {
		return nil, toStatus(err)
	}
	return &pb.ValidateGroupExistsResponse{Exists: exists}, nil
}

// CreateGroup creates a group.
func (h *Handler) CreateGroup(ctx context.Context, req *pb.CreateGroupRequest) (*pb.CreateGroupResponse, error) {
	start := time.Now()
	var err error
	defer func() {
		if err != nil {
			slog.Error("CreateGroup failed", "error", err, "duration_ms", time.Since(start).Milliseconds())
		}
	}()

	actor := actorFromContext(ctx)
	g, err := h.Service.CreateGroup(ctx,
		req.GetGroupName(),
		req.GetDescription(),
		req.GetIsActive(),
		actor,
	)
	if err != nil {
		return nil, toStatus(err)
	}
	return &pb.CreateGroupResponse{Group: groupToProto(g)}, nil
}

// UpdateGroup updates a group by ID.
func (h *Handler) UpdateGroup(ctx context.Context, req *pb.UpdateGroupRequest) (*pb.UpdateGroupResponse, error) {
	start := time.Now()
	var err error
	defer func() {
		if err != nil {
			slog.Error("UpdateGroup failed", "error", err, "duration_ms", time.Since(start).Milliseconds())
		}
	}()

	actor := actorFromContext(ctx)
	g, err := h.Service.UpdateGroup(ctx,
		req.GetId(),
		req.GroupName,
		req.Description,
		req.IsActive,
		actor,
	)
	if err != nil {
		return nil, toStatus(err)
	}
	return &pb.UpdateGroupResponse{Group: groupToProto(g)}, nil
}

// AddGroupMember adds a user to a group.
func (h *Handler) AddGroupMember(ctx context.Context, req *pb.AddGroupMemberRequest) (*pb.AddGroupMemberResponse, error) {
	start := time.Now()
	var err error
	defer func() {
		if err != nil {
			slog.Error("AddGroupMember failed", "error", err, "duration_ms", time.Since(start).Milliseconds())
		}
	}()

	actor := actorFromContext(ctx)
	m, err := h.Service.AddGroupMember(ctx,
		req.GetGroupId(),
		req.GetUserId(),
		req.GetMemberRole(),
		actor,
	)
	if err != nil {
		return nil, toStatus(err)
	}
	return &pb.AddGroupMemberResponse{Member: groupMemberToProto(m)}, nil
}

// RemoveGroupMember removes a user from a group.
func (h *Handler) RemoveGroupMember(ctx context.Context, req *pb.RemoveGroupMemberRequest) (*pb.RemoveGroupMemberResponse, error) {
	start := time.Now()
	var err error
	defer func() {
		if err != nil {
			slog.Error("RemoveGroupMember failed", "error", err, "duration_ms", time.Since(start).Milliseconds())
		}
	}()

	err = h.Service.RemoveGroupMember(ctx, req.GetGroupId(), req.GetUserId())
	if err != nil {
		return nil, toStatus(err)
	}
	return &pb.RemoveGroupMemberResponse{}, nil
}

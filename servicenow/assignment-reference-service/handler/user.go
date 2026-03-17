package handler

import (
	"context"
	"log/slog"
	"time"

	pb "github.com/servicenow/assignment-reference-service/proto/assignment_reference_service"
)

// GetUser returns a user by ID.
func (h *Handler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	start := time.Now()
	var err error
	defer func() {
		if err != nil {
			slog.Error("GetUser failed", "error", err, "duration_ms", time.Since(start).Milliseconds())
		}
	}()

	u, err := h.Service.GetUser(ctx, req.GetId())
	if err != nil {
		return nil, toStatus(err)
	}
	return &pb.GetUserResponse{User: userToProto(u)}, nil
}

// ListUsers returns a page of users with optional filters.
func (h *Handler) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	start := time.Now()
	var err error
	defer func() {
		if err != nil {
			slog.Error("ListUsers failed", "error", err, "duration_ms", time.Since(start).Milliseconds())
		}
	}()

	items, nextToken, err := h.Service.ListUsers(ctx,
		req.GetPageSize(),
		req.GetPageToken(),
		req.ActiveOnly,
		req.FilterDisplayName,
		req.FilterUsername,
		req.FilterEmail,
	)
	if err != nil {
		return nil, toStatus(err)
	}
	out := make([]*pb.User, len(items))
	for i, u := range items {
		out[i] = userToProto(u)
	}
	return &pb.ListUsersResponse{Items: out, NextPageToken: nextToken}, nil
}

// ValidateUserExists checks if the user ID exists.
func (h *Handler) ValidateUserExists(ctx context.Context, req *pb.ValidateUserExistsRequest) (*pb.ValidateUserExistsResponse, error) {
	exists, err := h.Service.ValidateUserExists(ctx, req.GetUserId())
	if err != nil {
		return nil, toStatus(err)
	}
	return &pb.ValidateUserExistsResponse{Exists: exists}, nil
}

// CreateUser creates a user.
func (h *Handler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	start := time.Now()
	var err error
	defer func() {
		if err != nil {
			slog.Error("CreateUser failed", "error", err, "duration_ms", time.Since(start).Milliseconds())
		}
	}()

	actor := actorFromContext(ctx)
	u, err := h.Service.CreateUser(ctx,
		req.GetUsername(),
		req.GetEmail(),
		req.GetDisplayName(),
		req.GetIsActive(),
		actor,
	)
	if err != nil {
		return nil, toStatus(err)
	}
	return &pb.CreateUserResponse{User: userToProto(u)}, nil
}

// UpdateUser updates a user by ID.
func (h *Handler) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	start := time.Now()
	var err error
	defer func() {
		if err != nil {
			slog.Error("UpdateUser failed", "error", err, "duration_ms", time.Since(start).Milliseconds())
		}
	}()

	actor := actorFromContext(ctx)
	u, err := h.Service.UpdateUser(ctx,
		req.GetId(),
		req.Username,
		req.Email,
		req.DisplayName,
		req.IsActive,
		actor,
	)
	if err != nil {
		return nil, toStatus(err)
	}
	return &pb.UpdateUserResponse{User: userToProto(u)}, nil
}

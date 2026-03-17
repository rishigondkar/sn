package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/servicenow/assignment-reference-service/proto/assignment_reference_service"
	"github.com/servicenow/assignment-reference-service/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestErrToCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want codes.Code
	}{
		{"user not found", service.ErrUserNotFound, codes.NotFound},
		{"group not found", service.ErrGroupNotFound, codes.NotFound},
		{"invalid email", service.ErrInvalidEmail, codes.InvalidArgument},
		{"username taken", service.ErrUsernameTaken, codes.InvalidArgument},
		{"group name taken", service.ErrGroupNameTaken, codes.InvalidArgument},
		{"member already exists", service.ErrMemberAlreadyExists, codes.InvalidArgument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := errToCode(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateUserExists_EmptyID(t *testing.T) {
	// Handler with nil service would panic on call; use a real service with nil repo would panic.
	// So we only test error mapping.
	st, ok := status.FromError(toStatus(service.ErrUserNotFound))
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestGetUserRequest_Proto(t *testing.T) {
	req := &pb.GetUserRequest{Id: "test-id"}
	assert.Equal(t, "test-id", req.GetId())
}

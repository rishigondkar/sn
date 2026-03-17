package handler

import (
	"errors"
	"testing"

	"github.com/servicenow/case-service/domain"
	"github.com/servicenow/case-service/repository"
	"github.com/servicenow/case-service/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMapError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code codes.Code
	}{
		{"not found", service.ErrNotFound, codes.NotFound},
		{"version conflict", repository.ErrVersionConflict, codes.FailedPrecondition},
		{"invalid state", domain.ErrInvalidState, codes.InvalidArgument},
		{"invalid priority", domain.ErrInvalidPriority, codes.InvalidArgument},
		{"closure required", domain.ErrClosureRequired, codes.InvalidArgument},
		{"already closed", domain.ErrAlreadyClosed, codes.InvalidArgument},
		{"use CloseCase", service.ErrUseCloseCase, codes.InvalidArgument},
		{"other", errors.New("other"), codes.Internal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapped := mapError(tt.err)
			st, ok := status.FromError(mapped)
			if !ok {
				t.Fatalf("mapError did not return status error")
			}
			if st.Code() != tt.code {
				t.Errorf("mapError() code = %v, want %v", st.Code(), tt.code)
			}
		})
	}
}

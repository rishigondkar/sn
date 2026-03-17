package handler

import (
	"context"
	"testing"

	"github.com/org/alert-observable-service/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMapErr(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code codes.Code
	}{
		{"validation", service.ErrValidation{Field: "x", Issue: "y"}, codes.InvalidArgument},
		{"not found", &service.ErrNotFound{Resource: "alert", ID: "1"}, codes.NotFound},
		{"already exists", &service.ErrAlreadyExists{Resource: "co", Detail: "dup"}, codes.AlreadyExists},
		{"other", context.DeadlineExceeded, codes.Internal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapped := mapErr(tt.err)
			st, ok := status.FromError(mapped)
			if !ok {
				t.Fatal("expected gRPC status")
			}
			if st.Code() != tt.code {
				t.Errorf("code = %v, want %v", st.Code(), tt.code)
			}
		})
	}
}

func TestInvalidArg(t *testing.T) {
	err := invalidArg("field", "required")
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status")
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", st.Code())
	}
}

// Server is constructed with a non-nil service in production; error mapping is covered above.

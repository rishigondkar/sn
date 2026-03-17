package handler

import (
	"errors"
	"log/slog"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/servicenow/assignment-reference-service/service"
)

// errToCode maps service errors to gRPC status codes (platform contract).
func errToCode(err error) codes.Code {
	if err == nil {
		return codes.OK
	}
	switch {
	case errors.Is(err, service.ErrUserNotFound), errors.Is(err, service.ErrGroupNotFound), errors.Is(err, service.ErrMemberNotFound):
		return codes.NotFound
	case errors.Is(err, service.ErrInvalidEmail), errors.Is(err, service.ErrUsernameTaken), errors.Is(err, service.ErrEmailTaken),
		errors.Is(err, service.ErrGroupNameTaken), errors.Is(err, service.ErrMemberAlreadyExists):
		return codes.InvalidArgument
	default:
		// Duplicate key / unique violation can be AlreadyExists
		if isAlreadyExists(err) {
			return codes.AlreadyExists
		}
		slog.Error("unmapped error", "error", err)
		return codes.Internal
	}
}

// toStatus returns a gRPC status from a service error. Internal errors get a generic message.
func toStatus(err error) error {
	if err == nil {
		return nil
	}
	code := errToCode(err)
	msg := err.Error()
	if code == codes.Internal {
		msg = "internal error"
	}
	return status.Error(code, msg)
}

// isAlreadyExists heuristically detects duplicate/unique constraint errors.
func isAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "unique") ||
		strings.Contains(s, "duplicate") ||
		strings.Contains(s, "already exists")
}

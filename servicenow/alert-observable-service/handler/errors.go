package handler

import (
	"errors"
	"fmt"

	"github.com/org/alert-observable-service/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	var ve service.ErrValidation
	var nf *service.ErrNotFound
	var ae *service.ErrAlreadyExists
	switch {
	case errors.As(err, &ve):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.As(err, &nf):
		return status.Error(codes.NotFound, err.Error())
	case errors.As(err, &ae):
		return status.Error(codes.AlreadyExists, err.Error())
	default:
		// Include message for debugging (e.g. Scan errors, DB errors). Sanitize in production if needed.
		msg := err.Error()
		if msg == "" {
			msg = "internal error"
		}
		return status.Error(codes.Internal, msg)
	}
}

func invalidArg(field, issue string) error {
	return status.Error(codes.InvalidArgument, fmt.Sprintf("validation: %s %s", field, issue))
}

package service

import "fmt"

// ErrValidation is a validation error (maps to gRPC InvalidArgument).
type ErrValidation struct {
	Field string
	Issue string
}

func (e ErrValidation) Error() string {
	return fmt.Sprintf("validation: %s %s", e.Field, e.Issue)
}

// ErrNotFound is a missing resource (maps to gRPC NotFound).
type ErrNotFound struct {
	Resource string
	ID       string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}

// ErrAlreadyExists is duplicate/idempotency (maps to gRPC AlreadyExists).
type ErrAlreadyExists struct {
	Resource string
	Detail   string
}

func (e ErrAlreadyExists) Error() string {
	return fmt.Sprintf("%s already exists: %s", e.Resource, e.Detail)
}

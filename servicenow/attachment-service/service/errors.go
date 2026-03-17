package service

import "fmt"

// ErrValidation is returned when request validation fails.
type ErrValidation struct {
	Field string
	Msg   string
}

func (e ErrValidation) Error() string {
	return fmt.Sprintf("validation: %s: %s", e.Field, e.Msg)
}

// ErrNotFound is returned when the requested attachment does not exist or is already deleted.
var ErrNotFound = fmt.Errorf("attachment not found")

package service

import (
	"errors"
	"testing"

	"github.com/servicenow/case-service/domain"
)

func TestCloseCase_Validation(t *testing.T) {
	if err := domain.ValidateClose("", "reason", "user-1"); err == nil {
		t.Error("expected error when closure_code empty")
	}
	if err := domain.ValidateClose("code", "reason", ""); err == nil {
		t.Error("expected error when actor empty")
	}
}

func TestErrNotFound(t *testing.T) {
	if !errors.Is(ErrNotFound, ErrNotFound) {
		t.Error("ErrNotFound should match itself")
	}
}

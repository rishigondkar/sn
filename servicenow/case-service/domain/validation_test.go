package domain

import (
	"testing"
)

func TestValidateCreateCase(t *testing.T) {
	tests := []struct {
		name      string
		shortDesc string
		state     string
		priority  string
		severity  string
		openedBy  string
		openedSet bool
		wantErr   bool
	}{
		{"ok", "short", "Draft", "P1", "high", "user-1", true, false},
		{"missing short", "", "Draft", "P1", "high", "user-1", true, true},
		{"short too long", string(make([]byte, 501)), "Draft", "P1", "high", "user-1", true, true},
		{"invalid state", "short", "invalid", "P1", "high", "user-1", true, true},
		{"invalid priority", "short", "Draft", "P5", "high", "user-1", true, true},
		{"invalid severity", "short", "Draft", "P1", "x", "user-1", true, true},
		{"missing opened_by", "short", "Draft", "P1", "high", "", true, true},
		{"missing opened_time", "short", "Draft", "P1", "high", "user-1", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCreateCase(tt.shortDesc, tt.state, tt.priority, tt.severity, tt.openedBy, tt.openedSet)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreateCase() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateShortDescription(t *testing.T) {
	tests := []struct {
		s      string
		wantErr bool
	}{
		{"a", false},
		{"valid", false},
		{"", true},
		{" ", true},
		{string(make([]byte, 501)), true},
	}
	for _, tt := range tests {
		err := ValidateShortDescription(tt.s)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateShortDescription(%q) error = %v, wantErr %v", tt.s, err, tt.wantErr)
		}
	}
}

func TestCanTransitionState(t *testing.T) {
	tests := []struct {
		old, new string
		wantErr  bool
	}{
		{"Draft", "Analysis", false},
		{"Analysis", "Contain", false},
		{"Closed", "Draft", true},
		{"Resolved", "Analysis", true},
		{"Draft", "invalid", true},
	}
	for _, tt := range tests {
		err := CanTransitionState(tt.old, tt.new)
		if (err != nil) != tt.wantErr {
			t.Errorf("CanTransitionState(%q, %q) error = %v, wantErr %v", tt.old, tt.new, err, tt.wantErr)
		}
	}
}

func TestValidateClose(t *testing.T) {
	tests := []struct {
		code, reason, actor string
		wantErr              bool
	}{
		{"code1", "reason", "user-1", false},
		{"", "reason", "user-1", true},
		{"code", "", "user-1", true},
		{"code", "reason", "", true},
	}
	for _, tt := range tests {
		err := ValidateClose(tt.code, tt.reason, tt.actor)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateClose() error = %v, wantErr %v", err, tt.wantErr)
		}
	}
}

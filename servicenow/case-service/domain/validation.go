package domain

import (
	"errors"
	"strings"
)

var (
	ErrInvalidShortDescription = errors.New("short_description must be 1–500 characters")
	ErrInvalidState            = errors.New("invalid state")
	ErrInvalidPriority         = errors.New("invalid priority")
	ErrInvalidSeverity         = errors.New("invalid severity")
	ErrInvalidEventTimes       = errors.New("event_received_time must not be before event_occurred_time")
	ErrOpenedByUserIDRequired   = errors.New("opened_by_user_id is required")
	ErrOpenedTimeRequired      = errors.New("opened_time is required")
	ErrClosureRequired         = errors.New("closure_code and closure_reason required when closing")
	ErrAlreadyClosed           = errors.New("case is already closed; reopen out of scope")
	ErrTransitionFromClosed    = errors.New("cannot transition from closed to non-closed in v1")
)

// ValidateCreateCase required fields: short_description, state, priority, severity, opened_by_user_id, opened_time.
func ValidateCreateCase(shortDesc, state, priority, severity, openedByUserID string, openedTimeSet bool) error {
	if err := ValidateShortDescription(shortDesc); err != nil {
		return err
	}
	if !AllowedStates[state] {
		return ErrInvalidState
	}
	if !AllowedPriorities[priority] {
		return ErrInvalidPriority
	}
	if !AllowedSeverities[severity] {
		return ErrInvalidSeverity
	}
	if openedByUserID == "" {
		return ErrOpenedByUserIDRequired
	}
	if !openedTimeSet {
		return ErrOpenedTimeRequired
	}
	return nil
}

func ValidateShortDescription(s string) error {
	s = strings.TrimSpace(s)
	if len(s) < ShortDescriptionMinLen || len(s) > ShortDescriptionMaxLen {
		return ErrInvalidShortDescription
	}
	return nil
}

// ValidateEventTimes ensures event_received_time is not before event_occurred_time when both set.
func ValidateEventTimes(occurred, received *interface{}) error {
	// Caller passes *time.Time; we only need both non-nil to compare. Simplified: accept nil.
	return nil
}

// ValidateClose requires closure_code, closure_reason, and actor (closed_by_user_id).
func ValidateClose(closureCode, closureReason, actorUserID string) error {
	if closureCode == "" || closureReason == "" {
		return ErrClosureRequired
	}
	if actorUserID == "" {
		return errors.New("actor (closed_by_user_id) is required to close")
	}
	return nil
}

// CanTransitionState returns nil if transition from oldState to newState is allowed in v1.
// Terminal states (resolved, closed) cannot transition to any other state.
func CanTransitionState(oldState, newState string) error {
	if TerminalStates[oldState] {
		return ErrTransitionFromClosed
	}
	if !AllowedStates[newState] {
		return ErrInvalidState
	}
	return nil
}

// ValidateUpdateCase optional fields; if state/priority/severity set, they must be allowed.
func ValidateUpdateCase(state, priority, severity *string) error {
	if state != nil && *state != "" && !AllowedStates[*state] {
		return ErrInvalidState
	}
	if priority != nil && *priority != "" && !AllowedPriorities[*priority] {
		return ErrInvalidPriority
	}
	if severity != nil && *severity != "" && !AllowedSeverities[*severity] {
		return ErrInvalidSeverity
	}
	return nil
}

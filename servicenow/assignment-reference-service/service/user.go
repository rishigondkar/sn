package service

import (
	"context"
	"errors"
	"log/slog"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/servicenow/assignment-reference-service/repository"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidEmail      = errors.New("invalid email format")
	ErrUsernameTaken     = errors.New("username already in use")
	ErrEmailTaken        = errors.New("email already in use")
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// Actor holds request metadata from gRPC metadata or REST headers.
type Actor struct {
	UserID        string
	DisplayName  string
	RequestID    string
	CorrelationID string
}

// GetUser returns a user by ID. Returns ErrUserNotFound if not found.
func (s *Service) GetUser(ctx context.Context, id string) (*repository.User, error) {
	if id == "" {
		return nil, ErrUserNotFound
	}
	u, err := s.Repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

// ListUsers returns a page of users with optional filters.
func (s *Service) ListUsers(ctx context.Context, pageSize int32, pageToken string, activeOnly *bool, filterDisplayName, filterUsername, filterEmail *string) ([]*repository.User, string, error) {
	f := repository.ListUsersFilter{}
	if activeOnly != nil {
		f.ActiveOnly = *activeOnly
	} else {
		f.ActiveOnly = true // default per PRD
	}
	if filterDisplayName != nil {
		f.FilterDisplayName = *filterDisplayName
	}
	if filterUsername != nil {
		f.FilterUsername = *filterUsername
	}
	if filterEmail != nil {
		f.FilterEmail = *filterEmail
	}
	return s.Repo.ListUsers(ctx, int(pageSize), pageToken, f)
}

// ValidateUserExists returns whether the user ID exists.
func (s *Service) ValidateUserExists(ctx context.Context, userID string) (bool, error) {
	return s.Repo.UserExistsByID(ctx, userID)
}

// CreateUser creates a user. Validates email format and username/email uniqueness.
func (s *Service) CreateUser(ctx context.Context, username, email, displayName string, isActive bool, actor Actor) (*repository.User, error) {
	if !emailRegex.MatchString(email) {
		return nil, ErrInvalidEmail
	}
	exists, err := s.Repo.UserExistsByUsername(ctx, username, "")
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUsernameTaken
	}
	exists, err = s.Repo.UserExistsByEmail(ctx, email, "")
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailTaken
	}
	u, err := s.Repo.CreateUserWithID(ctx, username, email, displayName, isActive)
	if err != nil {
		return nil, err
	}
	_ = s.publishUserUpdated(ctx, "create", nil, u, actor)
	return u, nil
}

// UpdateUser updates a user by ID. Validates email if provided and uniqueness.
func (s *Service) UpdateUser(ctx context.Context, id string, username, email, displayName *string, isActive *bool, actor Actor) (*repository.User, error) {
	if id == "" {
		return nil, ErrUserNotFound
	}
	cur, err := s.Repo.GetUserByID(ctx, id)
	if err != nil || cur == nil {
		return nil, ErrUserNotFound
	}
	prev := *cur
	if username != nil {
		cur.Username = *username
	}
	if email != nil {
		if !emailRegex.MatchString(*email) {
			return nil, ErrInvalidEmail
		}
		cur.Email = *email
	}
	if displayName != nil {
		cur.DisplayName = *displayName
	}
	if isActive != nil {
		cur.IsActive = *isActive
	}
	cur.UpdatedAt = time.Now().UTC()

	if cur.Username != prev.Username {
		exists, err := s.Repo.UserExistsByUsername(ctx, cur.Username, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrUsernameTaken
		}
	}
	if cur.Email != prev.Email {
		exists, err := s.Repo.UserExistsByEmail(ctx, cur.Email, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrEmailTaken
		}
	}

	if err := s.Repo.UpdateUser(ctx, cur); err != nil {
		return nil, err
	}
	_ = s.publishUserUpdated(ctx, "update", &prev, cur, actor)
	return cur, nil
}

func (s *Service) publishUserUpdated(ctx context.Context, action string, before, after *repository.User, actor Actor) error {
	afterData := map[string]interface{}{
		"id": after.ID, "username": after.Username, "email": after.Email,
		"display_name": after.DisplayName, "is_active": after.IsActive,
		"updated_at": after.UpdatedAt.Format(time.RFC3339),
	}
	var beforeData map[string]interface{}
	if before != nil {
		beforeData = map[string]interface{}{
			"id": before.ID, "username": before.Username, "email": before.Email,
			"display_name": before.DisplayName, "is_active": before.IsActive,
			"updated_at": before.UpdatedAt.Format(time.RFC3339),
		}
	}
	evt := AuditEvent{
		EventID:       uuid.New().String(),
		EventType:     "user.updated",
		SourceService: "assignment-reference-service",
		EntityType:    "user",
		EntityID:      after.ID,
		Action:        action,
		ActorUserID:   actor.UserID,
		ActorName:     actor.DisplayName,
		RequestID:     actor.RequestID,
		CorrelationID: actor.CorrelationID,
		AfterData:     afterData,
		BeforeData:    beforeData,
		OccurredAt:    time.Now().UTC(),
	}
	if err := s.Audit.Publish(evt); err != nil {
		slog.WarnContext(ctx, "audit publish failed", "error", err, "event_type", evt.EventType)
	}
	return nil
}

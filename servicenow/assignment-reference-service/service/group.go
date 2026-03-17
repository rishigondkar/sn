package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/servicenow/assignment-reference-service/repository"
)

var (
	ErrGroupNotFound       = errors.New("group not found")
	ErrGroupNameTaken      = errors.New("group name already in use")
	ErrMemberAlreadyExists = errors.New("user is already a member of the group")
	ErrMemberNotFound      = errors.New("membership not found")
)

// GetGroup returns a group by ID. Returns ErrGroupNotFound if not found.
func (s *Service) GetGroup(ctx context.Context, id string) (*repository.Group, error) {
	if id == "" {
		return nil, ErrGroupNotFound
	}
	g, err := s.Repo.GetGroupByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, ErrGroupNotFound
	}
	return g, nil
}

// ListGroups returns a page of groups with optional filters.
func (s *Service) ListGroups(ctx context.Context, pageSize int32, pageToken string, activeOnly *bool, filterGroupName *string) ([]*repository.Group, string, error) {
	f := repository.ListGroupsFilter{}
	if activeOnly != nil {
		f.ActiveOnly = *activeOnly
	} else {
		f.ActiveOnly = true
	}
	if filterGroupName != nil {
		f.FilterGroupName = *filterGroupName
	}
	return s.Repo.ListGroups(ctx, int(pageSize), pageToken, f)
}

// ListGroupMembers returns a page of members for a group.
func (s *Service) ListGroupMembers(ctx context.Context, groupID string, pageSize int32, pageToken string) ([]*repository.GroupMember, string, error) {
	if groupID == "" {
		return nil, "", ErrGroupNotFound
	}
	exists, err := s.Repo.GroupExistsByID(ctx, groupID)
	if err != nil {
		return nil, "", err
	}
	if !exists {
		return nil, "", ErrGroupNotFound
	}
	return s.Repo.ListGroupMembers(ctx, groupID, int(pageSize), pageToken)
}

// ValidateGroupExists returns whether the group ID exists.
func (s *Service) ValidateGroupExists(ctx context.Context, groupID string) (bool, error) {
	return s.Repo.GroupExistsByID(ctx, groupID)
}

// CreateGroup creates a group. Validates group name uniqueness.
func (s *Service) CreateGroup(ctx context.Context, groupName, description string, isActive bool, actor Actor) (*repository.Group, error) {
	exists, err := s.Repo.GroupExistsByName(ctx, groupName, "")
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrGroupNameTaken
	}
	g, err := s.Repo.CreateGroupWithID(ctx, groupName, description, isActive)
	if err != nil {
		return nil, err
	}
	_ = s.publishGroupUpdated(ctx, "create", nil, g, actor)
	return g, nil
}

// UpdateGroup updates a group by ID. Validates group name uniqueness.
func (s *Service) UpdateGroup(ctx context.Context, id string, groupName, description *string, isActive *bool, actor Actor) (*repository.Group, error) {
	if id == "" {
		return nil, ErrGroupNotFound
	}
	cur, err := s.Repo.GetGroupByID(ctx, id)
	if err != nil || cur == nil {
		return nil, ErrGroupNotFound
	}
	prev := *cur
	if groupName != nil {
		cur.GroupName = *groupName
	}
	if description != nil {
		cur.Description = *description
	}
	if isActive != nil {
		cur.IsActive = *isActive
	}
	cur.UpdatedAt = time.Now().UTC()

	if cur.GroupName != prev.GroupName {
		exists, err := s.Repo.GroupExistsByName(ctx, cur.GroupName, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrGroupNameTaken
		}
	}

	if err := s.Repo.UpdateGroup(ctx, cur); err != nil {
		return nil, err
	}
	_ = s.publishGroupUpdated(ctx, "update", &prev, cur, actor)
	return cur, nil
}

// AddGroupMember adds a user to a group. Returns ErrMemberAlreadyExists if (group_id, user_id) exists.
func (s *Service) AddGroupMember(ctx context.Context, groupID, userID, memberRole string, actor Actor) (*repository.GroupMember, error) {
	exists, err := s.Repo.GroupExistsByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrGroupNotFound
	}
	exists, err = s.Repo.UserExistsByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrUserNotFound
	}
	exists, err = s.Repo.GroupMemberExists(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrMemberAlreadyExists
	}
	return s.Repo.AddGroupMember(ctx, groupID, userID, memberRole)
}

// RemoveGroupMember removes a user from a group. Returns ErrMemberNotFound if not a member.
func (s *Service) RemoveGroupMember(ctx context.Context, groupID, userID string) error {
	removed, err := s.Repo.RemoveGroupMember(ctx, groupID, userID)
	if err != nil {
		return err
	}
	if !removed {
		return ErrMemberNotFound
	}
	return nil
}

func (s *Service) publishGroupUpdated(ctx context.Context, action string, before, after *repository.Group, actor Actor) error {
	afterData := map[string]interface{}{
		"id": after.ID, "group_name": after.GroupName, "description": after.Description,
		"is_active": after.IsActive, "updated_at": after.UpdatedAt.Format(time.RFC3339),
	}
	var beforeData map[string]interface{}
	if before != nil {
		beforeData = map[string]interface{}{
			"id": before.ID, "group_name": before.GroupName, "description": before.Description,
			"is_active": before.IsActive, "updated_at": before.UpdatedAt.Format(time.RFC3339),
		}
	}
	evt := AuditEvent{
		EventID:       uuid.New().String(),
		EventType:     "group.updated",
		SourceService: "assignment-reference-service",
		EntityType:    "group",
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

package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/servicenow/case-service/audit"
	"github.com/servicenow/case-service/domain"
	"github.com/servicenow/case-service/repository"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ActorMeta holds actor and request metadata from gRPC metadata or REST headers.
type ActorMeta struct {
	UserID        string
	Name          string
	RequestID     string
	CorrelationID string
}

// seedUserUUID is used when the actor UserID is not a valid UUID (e.g. "frontend-user"); DB UUID columns require a valid UUID.
const seedUserUUID = "a0000000-0000-4000-8000-000000000001"

// actorUserIDForDB returns a UUID string suitable for DB columns (closed_by_user_id, etc.). Non-UUID values are mapped to seedUserUUID.
func actorUserIDForDB(userID string) string {
	if userID == "" {
		return seedUserUUID
	}
	if _, err := uuid.Parse(userID); err != nil {
		return seedUserUUID
	}
	return userID
}

// CreateCase creates a case and publishes case.created.
func (s *Service) CreateCase(ctx context.Context, input *domain.Case, actor *ActorMeta) (id, caseNumber string, err error) {
	openedTimeSet := !input.OpenedTime.IsZero()
	if err := domain.ValidateCreateCase(input.ShortDescription, input.State, input.Priority, input.Severity, input.OpenedByUserID, openedTimeSet); err != nil {
		return "", "", err
	}
	now := time.Now().UTC()
	id = uuid.Must(uuid.NewV7()).String()
	var cn string
	err = s.Repo.RunInTx(ctx, func(tx pgx.Tx) error {
		var innerErr error
		cn, innerErr = s.Repo.NextCaseNumber(ctx, tx)
		if innerErr != nil {
			return innerErr
		}
		input.ID = id
		input.CaseNumber = cn
		input.VersionNo = 1
		input.CreatedAt = now
		input.UpdatedAt = now
		input.IsActive = !domain.TerminalStates[input.State]
		input.ActiveDurationSeconds = 0
		if domain.ActiveStates[input.State] {
			input.ActiveDurationSeconds = int64(time.Since(input.OpenedTime).Seconds())
		}
		return s.Repo.CreateCase(ctx, tx, input)
	})
	if err != nil {
		return "", "", err
	}
	caseNumber = cn
	evt := &audit.Event{
		EventID:       uuid.Must(uuid.NewV7()).String(),
		EventType:     audit.EventTypeCaseCreated,
		SourceService: audit.SourceService,
		EntityType:    "case",
		EntityID:      id,
		CaseID:        id,
		Action:        "create",
		ActorUserID:   actor.UserID,
		ActorName:     actor.Name,
		RequestID:     actor.RequestID,
		CorrelationID: actor.CorrelationID,
		AfterData:     caseToMap(input),
		OccurredAt:    time.Now().UTC(),
	}
	_ = s.AuditPub.Publish(context.Background(), evt)
	return id, caseNumber, nil
}

func caseToMap(c *domain.Case) map[string]interface{} {
	if c == nil {
		return nil
	}
	m := map[string]interface{}{
		"id": c.ID, "case_number": c.CaseNumber, "short_description": c.ShortDescription,
		"state": c.State, "priority": c.Priority, "severity": c.Severity,
		"opened_by_user_id": c.OpenedByUserID, "opened_time": c.OpenedTime,
		"version_no": c.VersionNo, "is_active": c.IsActive,
	}
	if c.Description != "" {
		m["description"] = c.Description
	}
	return m
}

// GetCase returns a case by ID. Updates active_duration_seconds when state is active (new/triage/in_progress).
func (s *Service) GetCase(ctx context.Context, id string) (*domain.Case, error) {
	c, err := s.Repo.GetCaseByID(ctx, id)
	if err != nil || c == nil {
		return nil, err
	}
	refreshActiveDuration(c)
	return c, nil
}

// refreshActiveDuration sets active_duration_seconds to elapsed time when state is active (v1: new, triage, in_progress).
func refreshActiveDuration(c *domain.Case) {
	if domain.ActiveStates[c.State] {
		c.ActiveDurationSeconds = int64(time.Since(c.OpenedTime).Seconds())
		if c.ActiveDurationSeconds < 0 {
			c.ActiveDurationSeconds = 0
		}
	}
}

// UpdateCaseInput carries optional fields for patch (only set fields are applied).
type UpdateCaseInput struct {
	ShortDescription      *string
	Description           *string
	State                 *string
	Priority              *string
	Severity              *string
	AffectedUserID        *string
	AssignedUserID        *string
	AssignmentGroupID     *string
	Accuracy              *string
	Determination         *string
	Impact                *string
	VersionNo             int32
	FollowupTime          *timestamppb.Timestamp
	Category              *string
	Subcategory           *string
	Source                *string
	SourceTool            *string
	SourceToolFeature     *string
	ConfigurationItem     *string
	SOCNotes              *string
	NextSteps             *string
	CSIRTClassification   *string
	SOCLeadUserID         *string
	NotificationTime      *timestamppb.Timestamp
	IsAffectedUserVIP     *bool
	RequestedByUserID     *string
	EnvironmentLevel      *string
	EnvironmentType       *string
	PDN                   *string
	ImpactedObject        *string
	MTTR                  *string
	ReassignmentCount     *int32
	AssignedToCount       *int32
	EngineeringDocument   *string
	ResponseDocument      *string
	ClosureReason         *string
	EventOccurredTime     *timestamppb.Timestamp
}

// UpdateCase updates case fields with optimistic concurrency. Publishes case.updated / case.state.changed. Closing must use CloseCase.
func (s *Service) UpdateCase(ctx context.Context, id string, updates *UpdateCaseInput, actor *ActorMeta) (*domain.Case, error) {
	c, err := s.Repo.GetCaseByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrNotFound
	}
	if err := domain.ValidateUpdateCase(updates.State, updates.Priority, updates.Severity); err != nil {
		return nil, err
	}
	if updates.ShortDescription != nil && *updates.ShortDescription != "" {
		if err := domain.ValidateShortDescription(*updates.ShortDescription); err != nil {
			return nil, err
		}
		c.ShortDescription = *updates.ShortDescription
	}
	if updates.Description != nil {
		c.Description = *updates.Description
	}
	if updates.State != nil && *updates.State != "" {
		if *updates.State == domain.StateClosed {
			return nil, ErrUseCloseCase
		}
		if err := domain.CanTransitionState(c.State, *updates.State); err != nil {
			return nil, err
		}
		c.State = *updates.State
		c.IsActive = !domain.TerminalStates[c.State]
	}
	if updates.Priority != nil && *updates.Priority != "" {
		if !domain.AllowedPriorities[*updates.Priority] {
			return nil, domain.ErrInvalidPriority
		}
		c.Priority = *updates.Priority
	}
	if updates.Severity != nil && *updates.Severity != "" {
		if !domain.AllowedSeverities[*updates.Severity] {
			return nil, domain.ErrInvalidSeverity
		}
		c.Severity = *updates.Severity
	}
	if updates.AffectedUserID != nil {
		c.AffectedUserID = updates.AffectedUserID
	}
	if updates.AssignedUserID != nil {
		c.AssignedUserID = updates.AssignedUserID
	}
	if updates.AssignmentGroupID != nil {
		c.AssignmentGroupID = updates.AssignmentGroupID
	}
	if updates.Accuracy != nil {
		c.Accuracy = updates.Accuracy
	}
	if updates.Determination != nil {
		c.Determination = updates.Determination
	}
	if updates.Impact != nil {
		c.Impact = updates.Impact
	}
	if updates.FollowupTime != nil && updates.FollowupTime.IsValid() {
		t := updates.FollowupTime.AsTime()
		c.FollowupTime = &t
	}
	if updates.Category != nil {
		c.Category = updates.Category
	}
	if updates.Subcategory != nil {
		c.Subcategory = updates.Subcategory
	}
	if updates.Source != nil {
		c.Source = updates.Source
	}
	if updates.SourceTool != nil {
		c.SourceTool = updates.SourceTool
	}
	if updates.SourceToolFeature != nil {
		c.SourceToolFeature = updates.SourceToolFeature
	}
	if updates.ConfigurationItem != nil {
		c.ConfigurationItem = updates.ConfigurationItem
	}
	if updates.SOCNotes != nil {
		c.SOCNotes = updates.SOCNotes
	}
	if updates.NextSteps != nil {
		c.NextSteps = updates.NextSteps
	}
	if updates.CSIRTClassification != nil {
		c.CSIRTClassification = updates.CSIRTClassification
	}
	if updates.SOCLeadUserID != nil {
		c.SOCLeadUserID = updates.SOCLeadUserID
	}
	if updates.NotificationTime != nil && updates.NotificationTime.IsValid() {
		t := updates.NotificationTime.AsTime()
		c.NotificationTime = &t
	}
	if updates.IsAffectedUserVIP != nil {
		c.IsAffectedUserVIP = *updates.IsAffectedUserVIP
	}
	if updates.RequestedByUserID != nil {
		c.RequestedByUserID = updates.RequestedByUserID
	}
	if updates.EnvironmentLevel != nil {
		c.EnvironmentLevel = updates.EnvironmentLevel
	}
	if updates.EnvironmentType != nil {
		c.EnvironmentType = updates.EnvironmentType
	}
	if updates.PDN != nil {
		c.PDN = updates.PDN
	}
	if updates.ImpactedObject != nil {
		c.ImpactedObject = updates.ImpactedObject
	}
	if updates.MTTR != nil {
		c.MTTR = updates.MTTR
	}
	if updates.ReassignmentCount != nil {
		c.ReassignmentCount = *updates.ReassignmentCount
	}
	if updates.AssignedToCount != nil {
		c.AssignedToCount = *updates.AssignedToCount
	}
	if updates.EngineeringDocument != nil {
		c.EngineeringDocument = updates.EngineeringDocument
	}
	if updates.ResponseDocument != nil {
		c.ResponseDocument = updates.ResponseDocument
	}
	if updates.EventOccurredTime != nil && updates.EventOccurredTime.IsValid() {
		t := updates.EventOccurredTime.AsTime()
		c.EventOccurredTime = &t
	}
	if updates.ClosureReason != nil {
		c.ClosureReason = updates.ClosureReason
		now := time.Now().UTC()
		c.ClosedTime = &now
		closedBy := actorUserIDForDB(actor.UserID)
		c.ClosedByUserID = &closedBy
		// Saving closure information from Review state moves the case to Closed.
		if c.State == domain.StateReview {
			c.State = domain.StateClosed
			c.IsActive = false
		}
	}
	c.VersionNo = updates.VersionNo
	c.UpdatedAt = time.Now().UTC()
	if err := s.Repo.UpdateCase(ctx, c); err != nil {
		return nil, err
	}
	c.VersionNo++ // DB did version_no = version_no + 1
	evt := &audit.Event{
		EventID:       uuid.Must(uuid.NewV7()).String(),
		EventType:     audit.EventTypeCaseUpdated,
		SourceService: audit.SourceService,
		EntityType:    "case",
		EntityID:      c.ID,
		CaseID:        c.ID,
		Action:        "update",
		ActorUserID:   actor.UserID,
		ActorName:     actor.Name,
		RequestID:     actor.RequestID,
		CorrelationID: actor.CorrelationID,
		AfterData:     caseToMap(c),
		OccurredAt:    time.Now().UTC(),
	}
	_ = s.AuditPub.Publish(context.Background(), evt)
	refreshActiveDuration(c)
	return c, nil
}

// ErrNotFound is returned when a case or worknote is not found.
var ErrNotFound = errors.New("not found")

// ErrUseCloseCase is returned when the client tries to set state to Closed via UpdateCase; they must use CloseCase instead.
var ErrUseCloseCase = errors.New("use CloseCase to close a case")

// AssignCase updates assignment and publishes case.assigned.
func (s *Service) AssignCase(ctx context.Context, caseID, assignedUserID, assignmentGroupID string, versionNo int32, actor *ActorMeta) (*domain.Case, error) {
	c, err := s.Repo.GetCaseByID(ctx, caseID)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrNotFound
	}
	c.AssignedUserID = ptr(assignedUserID)
	c.AssignmentGroupID = ptr(assignmentGroupID)
	c.VersionNo = versionNo
	c.UpdatedAt = time.Now().UTC()
	if err := s.Repo.UpdateCase(ctx, c); err != nil {
		return nil, err
	}
	evt := &audit.Event{
		EventID:       uuid.Must(uuid.NewV7()).String(),
		EventType:     audit.EventTypeCaseAssigned,
		SourceService: audit.SourceService,
		EntityType:    "case",
		EntityID:      c.ID,
		CaseID:        c.ID,
		Action:        "assign",
		ActorUserID:   actor.UserID,
		ActorName:     actor.Name,
		RequestID:     actor.RequestID,
		CorrelationID: actor.CorrelationID,
		AfterData:     caseToMap(c),
		OccurredAt:    time.Now().UTC(),
	}
	_ = s.AuditPub.Publish(context.Background(), evt)
	return s.GetCase(ctx, c.ID)
}

func ptr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// CloseCase sets state=closed, closure fields, closed_by/closed_time, is_active=false; publishes case.closed.
func (s *Service) CloseCase(ctx context.Context, caseID, closureCode, closureReason string, versionNo int32, actor *ActorMeta) (*domain.Case, error) {
	if err := domain.ValidateClose(closureCode, closureReason, actor.UserID); err != nil {
		return nil, err
	}
	c, err := s.Repo.GetCaseByID(ctx, caseID)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrNotFound
	}
	if c.State == domain.StateClosed {
		return nil, domain.ErrAlreadyClosed
	}
	c.State = domain.StateClosed
	c.ClosureCode = &closureCode
	c.ClosureReason = &closureReason
	closedBy := actorUserIDForDB(actor.UserID)
	c.ClosedByUserID = &closedBy
	now := time.Now().UTC()
	c.ClosedTime = &now
	c.IsActive = false
	c.VersionNo = versionNo
	c.UpdatedAt = now
	if err := s.Repo.UpdateCase(ctx, c); err != nil {
		return nil, err
	}
	evt := &audit.Event{
		EventID:       uuid.Must(uuid.NewV7()).String(),
		EventType:     audit.EventTypeCaseClosed,
		SourceService: audit.SourceService,
		EntityType:    "case",
		EntityID:      c.ID,
		CaseID:        c.ID,
		Action:        "close",
		ActorUserID:   actor.UserID,
		ActorName:     actor.Name,
		RequestID:     actor.RequestID,
		CorrelationID: actor.CorrelationID,
		AfterData:     caseToMap(c),
		OccurredAt:    now,
	}
	_ = s.AuditPub.Publish(context.Background(), evt)
	return s.GetCase(ctx, c.ID)
}

// ListCases returns cases with filters and pagination (opened_time desc).
func (s *Service) ListCases(ctx context.Context, f repository.ListCasesFilter) ([]*domain.Case, string, error) {
	list, next, err := s.Repo.ListCases(ctx, f)
	if err != nil {
		return nil, "", err
	}
	for _, c := range list {
		refreshActiveDuration(c)
	}
	return list, next, nil
}


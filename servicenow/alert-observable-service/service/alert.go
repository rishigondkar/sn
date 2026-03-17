package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/org/alert-observable-service/repository"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateAlertRule creates an alert rule and returns its ID.
func (s *Service) CreateAlertRule(ctx context.Context, name, ruleType, sourceSystem, externalRuleKey, description string, isActive bool) (uuid.UUID, error) {
	now := time.Now().UTC()
	id := uuid.Must(uuid.NewV7())
	r := &repository.AlertRule{
		ID: id, RuleName: name, RuleType: ptr(ruleType), SourceSystem: ptr(sourceSystem),
		ExternalRuleKey: ptr(externalRuleKey), Description: ptr(description), IsActive: isActive,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := s.AlertRuleRepo.Create(ctx, r); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

// CreateAlert creates an alert (source_system required); publishes alert.created after persist.
func (s *Service) CreateAlert(ctx context.Context, req *CreateAlertRequest, actor Actor) (uuid.UUID, error) {
	if req.SourceSystem == "" {
		return uuid.Nil, ErrValidation{Field: "source_system", Issue: "required"}
	}
	caseID, err := uuid.Parse(req.CaseID)
	if err != nil {
		return uuid.Nil, ErrValidation{Field: "case_id", Issue: "invalid uuid"}
	}
	var alertRuleID *uuid.UUID
	if req.AlertRuleID != "" {
		id, err := uuid.Parse(req.AlertRuleID)
		if err != nil {
			return uuid.Nil, ErrValidation{Field: "alert_rule_id", Issue: "invalid uuid"}
		}
		alertRuleID = &id
	}
	now := time.Now().UTC()
	alertID := uuid.Must(uuid.NewV7())
	var rawPayload []byte
	if req.RawPayloadJSON != "" {
		rawPayload = []byte(req.RawPayloadJSON)
		if !json.Valid(rawPayload) {
			return uuid.Nil, ErrValidation{Field: "raw_payload_json", Issue: "invalid json"}
		}
	}
	a := &repository.Alert{
		ID:                alertID,
		CaseID:            caseID,
		AlertRuleID:       alertRuleID,
		SourceSystem:      req.SourceSystem,
		SourceAlertID:     ptrOrNil(req.SourceAlertID),
		Title:             ptrOrNil(req.Title),
		Description:       ptrOrNil(req.Description),
		EventOccurredTime: protoTimeToTime(req.EventOccurredTime),
		EventReceivedTime: protoTimeToTime(req.EventReceivedTime),
		Severity:          ptrOrNil(req.Severity),
		RawPayload:        rawPayload,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := s.AlertRepo.Create(ctx, a); err != nil {
		return uuid.Nil, err
	}
	// Publish audit after successful persist
	if s.Audit != nil {
		_ = s.Audit.Publish(context.Background(), AuditEvent{
			EventID:       uuid.Must(uuid.NewV7()).String(),
			EventType:     "alert.created",
			SourceService: "alert-observable-service",
			EntityType:    "alert",
			EntityID:      alertID.String(),
			CaseID:        caseID.String(),
			Action:        "create",
			ActorUserID:   actor.UserID,
			ActorName:     actor.UserName,
			RequestID:     actor.RequestID,
			CorrelationID: actor.CorrelationID,
			AfterData:     map[string]interface{}{"case_id": caseID.String(), "source_system": req.SourceSystem},
			OccurredAt:    now.Format(time.RFC3339),
		})
	}
	return alertID, nil
}

// GetAlert returns an alert by ID.
func (s *Service) GetAlert(ctx context.Context, alertID string) (*repository.Alert, error) {
	id, err := uuid.Parse(alertID)
	if err != nil {
		return nil, ErrValidation{Field: "alert_id", Issue: "invalid uuid"}
	}
	a, err := s.AlertRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, &ErrNotFound{Resource: "alert", ID: alertID}
	}
	return a, nil
}

// ListCaseAlerts returns alerts for a case with pagination.
func (s *Service) ListCaseAlerts(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*repository.Alert, string, error) {
	cid, err := uuid.Parse(caseID)
	if err != nil {
		return nil, "", ErrValidation{Field: "case_id", Issue: "invalid uuid"}
	}
	return s.AlertRepo.ListByCaseID(ctx, cid, repository.ListOpts{PageSize: pageSize, PageToken: pageToken})
}

func ptr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func ptrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func protoTimeToTime(t *timestamppb.Timestamp) *time.Time {
	if t == nil || !t.IsValid() {
		return nil
	}
	tt := t.AsTime()
	return &tt
}

// CreateAlertRequest is the in-memory request for CreateAlert.
type CreateAlertRequest struct {
	CaseID            string
	AlertRuleID       string
	SourceSystem      string
	SourceAlertID     string
	Title             string
	Description       string
	EventOccurredTime *timestamppb.Timestamp
	EventReceivedTime *timestamppb.Timestamp
	Severity          string
	RawPayloadJSON    string
}

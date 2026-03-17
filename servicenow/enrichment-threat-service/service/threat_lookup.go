package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/servicenow/enrichment-threat-service/repository"
	"log/slog"
)

// UpsertThreatLookupInput is the input for upserting a threat lookup result.
type UpsertThreatLookupInput struct {
	ID                   string
	CaseID               *string
	ObservableID         string
	LookupType           string
	SourceName           string
	SourceRecordID       *string
	Verdict              *string
	RiskScore            *float64
	ConfidenceScore      *float64
	TagsJSON             *string
	MatchedIndicatorsJSON *string
	Summary              *string
	ResultDataJSON       string
	QueriedAt            *time.Time
	ReceivedAt           time.Time
	ExpiresAt            *time.Time
}

// UpsertThreatLookupResult validates, upserts, and publishes an audit event.
func (s *Service) UpsertThreatLookupResult(ctx context.Context, in UpsertThreatLookupInput, actor ActorInfo) (*repository.ThreatLookupResult, error) {
	if in.ObservableID == "" {
		return nil, ErrValidation
	}
	if in.ResultDataJSON == "" {
		return nil, ErrValidation
	}
	if !json.Valid([]byte(in.ResultDataJSON)) {
		return nil, ErrValidation
	}
	if s.Config != nil && s.Config.MaxPayloadBytes > 0 && len(in.ResultDataJSON) > s.Config.MaxPayloadBytes {
		return nil, ErrValidation
	}
	if in.LookupType == "" || in.SourceName == "" {
		return nil, ErrValidation
	}

	id := in.ID
	if id == "" {
		id = uuid.Must(uuid.NewV7()).String()
	}

	var tags, matchedIndicators json.RawMessage
	if in.TagsJSON != nil {
		tags = json.RawMessage(*in.TagsJSON)
	}
	if in.MatchedIndicatorsJSON != nil {
		matchedIndicators = json.RawMessage(*in.MatchedIndicatorsJSON)
	}

	t := &repository.ThreatLookupResult{
		ID:                id,
		CaseID:            in.CaseID,
		ObservableID:      in.ObservableID,
		LookupType:        in.LookupType,
		SourceName:        in.SourceName,
		SourceRecordID:    in.SourceRecordID,
		Verdict:           in.Verdict,
		RiskScore:         in.RiskScore,
		ConfidenceScore:   in.ConfidenceScore,
		Tags:              tags,
		MatchedIndicators: matchedIndicators,
		Summary:           in.Summary,
		ResultData:        json.RawMessage(in.ResultDataJSON),
		QueriedAt:         in.QueriedAt,
		ReceivedAt:        in.ReceivedAt,
		ExpiresAt:         in.ExpiresAt,
	}

	if err := s.Repo.UpsertThreatLookupResult(ctx, t); err != nil {
		slog.Error("repository upsert threat lookup failed", "err", err, "id", id)
		return nil, err
	}

	out, err := s.Repo.GetThreatLookupResultByID(ctx, id)
	if err != nil || out == nil {
		return t, nil
	}
	t = out

	if s.Audit != nil {
		event := AuditEventEnvelope{
			EventID:       uuid.Must(uuid.NewV7()).String(),
			EventType:     "threat_lookup_result.upserted",
			SourceService: "enrichment-threat-service",
			EntityType:    "threat_lookup_result",
			EntityID:      t.ID,
			ObservableID:  t.ObservableID,
			Action:        "upsert",
			ActorUserID:   actor.UserID,
			ActorName:     actor.Name,
			RequestID:     actor.RequestID,
			CorrelationID: actor.CorrelationID,
			ChangeSummary: "threat lookup result upserted",
			OccurredAt:    time.Now().UTC().Format(time.RFC3339),
		}
		if t.CaseID != nil {
			event.CaseID = *t.CaseID
		}
		_ = s.Audit.Publish(context.Background(), event)
	}
	return t, nil
}

// ListThreatLookupResultsByCase delegates to repository.
func (s *Service) ListThreatLookupResultsByCase(ctx context.Context, caseID string, f repository.ListFilter) ([]*repository.ThreatLookupResult, string, error) {
	return s.Repo.ListThreatLookupResultsByCase(ctx, caseID, f)
}

// ListThreatLookupResultsByObservable delegates to repository.
func (s *Service) ListThreatLookupResultsByObservable(ctx context.Context, observableID string, f repository.ListFilter) ([]*repository.ThreatLookupResult, string, error) {
	return s.Repo.ListThreatLookupResultsByObservable(ctx, observableID, f)
}

// GetThreatLookupResultByID returns one result by id.
func (s *Service) GetThreatLookupResultByID(ctx context.Context, id string) (*repository.ThreatLookupResult, error) {
	return s.Repo.GetThreatLookupResultByID(ctx, id)
}

// GetThreatLookupSummaryByObservable delegates to repository.
func (s *Service) GetThreatLookupSummaryByObservable(ctx context.Context, observableID string) (*repository.ThreatSummary, error) {
	return s.Repo.GetThreatLookupSummaryByObservable(ctx, observableID)
}

package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/servicenow/enrichment-threat-service/repository"
	"log/slog"
)

// UpsertEnrichmentInput is the input for upserting an enrichment result (from gRPC or REST).
type UpsertEnrichmentInput struct {
	ID              string
	CaseID          *string
	ObservableID    *string
	EnrichmentType  string
	SourceName      string
	SourceRecordID  *string
	Status          string
	Summary         *string
	ResultDataJSON  string
	Score           *float64
	Confidence      *float64
	RequestedAt     *time.Time
	ReceivedAt      time.Time
	ExpiresAt       *time.Time
	LastUpdatedBy   *string
}

// UpsertEnrichmentResult validates, upserts, and publishes an audit event.
func (s *Service) UpsertEnrichmentResult(ctx context.Context, in UpsertEnrichmentInput, actor ActorInfo) (*repository.EnrichmentResult, error) {
	// Validation: at least one of case_id or observable_id
	if (in.CaseID == nil || *in.CaseID == "") && (in.ObservableID == nil || *in.ObservableID == "") {
		return nil, ErrValidation
	}
	// result_data required and valid JSON
	if in.ResultDataJSON == "" {
		return nil, ErrValidation
	}
	if !json.Valid([]byte(in.ResultDataJSON)) {
		return nil, ErrValidation
	}
	// Payload size
	if s.Config != nil && s.Config.MaxPayloadBytes > 0 && len(in.ResultDataJSON) > s.Config.MaxPayloadBytes {
		return nil, ErrValidation
	}
	if in.EnrichmentType == "" || in.SourceName == "" || in.Status == "" {
		return nil, ErrValidation
	}

	id := in.ID
	if id == "" {
		id = uuid.Must(uuid.NewV7()).String()
	}

	e := &repository.EnrichmentResult{
		ID:             id,
		CaseID:         in.CaseID,
		ObservableID:   in.ObservableID,
		EnrichmentType: in.EnrichmentType,
		SourceName:     in.SourceName,
		SourceRecordID: in.SourceRecordID,
		Status:         in.Status,
		Summary:        in.Summary,
		ResultData:     json.RawMessage(in.ResultDataJSON),
		Score:          in.Score,
		Confidence:     in.Confidence,
		RequestedAt:    in.RequestedAt,
		ReceivedAt:     in.ReceivedAt,
		ExpiresAt:      in.ExpiresAt,
		LastUpdatedBy:  in.LastUpdatedBy,
	}

	if err := s.Repo.UpsertEnrichmentResult(ctx, e); err != nil {
		slog.Error("repository upsert enrichment failed", "err", err, "id", id)
		return nil, err
	}

	// Reload to get created_at/updated_at
	out, err := s.Repo.GetEnrichmentResultByID(ctx, id)
	if err != nil || out == nil {
		return e, nil // return what we have
	}
	e = out

	if s.Audit != nil {
		event := AuditEventEnvelope{
			EventID:       uuid.Must(uuid.NewV7()).String(),
			EventType:     "enrichment_result.upserted",
			SourceService: "enrichment-threat-service",
			EntityType:    "enrichment_result",
			EntityID:      e.ID,
			Action:        "upsert",
			ActorUserID:   actor.UserID,
			ActorName:     actor.Name,
			RequestID:     actor.RequestID,
			CorrelationID: actor.CorrelationID,
			ChangeSummary: "enrichment result upserted",
			OccurredAt:    time.Now().UTC().Format(time.RFC3339),
		}
		if e.CaseID != nil {
			event.CaseID = *e.CaseID
		}
		if e.ObservableID != nil {
			event.ObservableID = *e.ObservableID
		}
		_ = s.Audit.Publish(context.Background(), event)
	}
	return e, nil
}

// ActorInfo carries actor and tracing from context/metadata.
type ActorInfo struct {
	UserID       string
	Name         string
	RequestID    string
	CorrelationID string
}

// ListEnrichmentResultsByCase delegates to repository.
func (s *Service) ListEnrichmentResultsByCase(ctx context.Context, caseID string, f repository.ListFilter) ([]*repository.EnrichmentResult, string, error) {
	return s.Repo.ListEnrichmentResultsByCase(ctx, caseID, f)
}

// ListEnrichmentResultsByObservable delegates to repository.
func (s *Service) ListEnrichmentResultsByObservable(ctx context.Context, observableID string, f repository.ListFilter) ([]*repository.EnrichmentResult, string, error) {
	return s.Repo.ListEnrichmentResultsByObservable(ctx, observableID, f)
}

// GetEnrichmentResultByID returns one result by id.
func (s *Service) GetEnrichmentResultByID(ctx context.Context, id string) (*repository.EnrichmentResult, error) {
	return s.Repo.GetEnrichmentResultByID(ctx, id)
}

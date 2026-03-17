package handler

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/servicenow/audit-service/proto/audit_service"
	"github.com/servicenow/audit-service/repository"
)

func (h *Handler) IngestEvent(ctx context.Context, req *audit_service.IngestEventRequest) (*audit_service.IngestEventResponse, error) {
	startTime := time.Now()
	method := "IngestEvent"
	var err error
	defer func() {
		if err != nil {
			slog.Default().ErrorContext(ctx, "handler error", slog.String("layer", currentLayer), slog.String("method", method), slog.Any("error", err), slog.Duration("duration", time.Since(startTime)))
		}
	}()

	if req == nil || req.GetEvent() == nil {
		err = status.Error(codes.InvalidArgument, "event is required")
		return nil, err
	}
	e := req.GetEvent()
	if e.GetEventId() == "" {
		err = status.Error(codes.InvalidArgument, "event_id is required")
		return nil, err
	}
	if e.GetEventType() == "" || e.GetSourceService() == "" || e.GetEntityType() == "" || e.GetEntityId() == "" || e.GetAction() == "" {
		err = status.Error(codes.InvalidArgument, "event_type, source_service, entity_type, entity_id, action are required")
		return nil, err
	}

	row := protoEventToRow(e)
	inserted, err := h.Service.IngestEvent(ctx, row)
	if err != nil {
		err = status.Error(codes.Internal, "failed to ingest event")
		return nil, err
	}
	return &audit_service.IngestEventResponse{Ingested: inserted}, nil
}

func protoEventToRow(e *audit_service.AuditEvent) *repository.AuditEventRow {
	row := &repository.AuditEventRow{
		EventID:       e.GetEventId(),
		EventType:     e.GetEventType(),
		SourceService: e.GetSourceService(),
		EntityType:    e.GetEntityType(),
		EntityID:      e.GetEntityId(),
		Action:        e.GetAction(),
	}
	if e.GetId() != "" {
		row.ID = e.GetId()
	} else {
		row.ID = uuid.Must(uuid.NewV7()).String()
	}
	if e.GetParentEntityType() != "" {
		row.ParentEntityType = sql.NullString{String: e.GetParentEntityType(), Valid: true}
	}
	if e.GetParentEntityId() != "" {
		row.ParentEntityID = sql.NullString{String: e.GetParentEntityId(), Valid: true}
	}
	if e.GetCaseId() != "" {
		row.CaseID = sql.NullString{String: e.GetCaseId(), Valid: true}
	}
	if e.GetObservableId() != "" {
		row.ObservableID = sql.NullString{String: e.GetObservableId(), Valid: true}
	}
	if e.GetActorUserId() != "" {
		row.ActorUserID = sql.NullString{String: e.GetActorUserId(), Valid: true}
	}
	if e.GetActorName() != "" {
		row.ActorName = sql.NullString{String: e.GetActorName(), Valid: true}
	}
	if e.GetRequestId() != "" {
		row.RequestID = sql.NullString{String: e.GetRequestId(), Valid: true}
	}
	if e.GetCorrelationId() != "" {
		row.CorrelationID = sql.NullString{String: e.GetCorrelationId(), Valid: true}
	}
	if e.GetChangeSummary() != "" {
		row.ChangeSummary = sql.NullString{String: e.GetChangeSummary(), Valid: true}
	}
	if e.GetBeforeDataJson() != "" {
		row.BeforeData = []byte(e.GetBeforeDataJson())
	}
	if e.GetAfterDataJson() != "" {
		row.AfterData = []byte(e.GetAfterDataJson())
	}
	if e.GetMetadataJson() != "" {
		row.Metadata = []byte(e.GetMetadataJson())
	}
	if e.GetOccurredAt() != nil && e.GetOccurredAt().IsValid() {
		row.OccurredAt = e.GetOccurredAt().AsTime()
	} else {
		row.OccurredAt = time.Now().UTC()
	}
	if e.GetIngestedAt() != nil && e.GetIngestedAt().IsValid() {
		row.IngestedAt = e.GetIngestedAt().AsTime()
	} else {
		row.IngestedAt = time.Now().UTC()
	}
	return row
}

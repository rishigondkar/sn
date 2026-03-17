package handler

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/servicenow/audit-service/proto/audit_service"
	"github.com/servicenow/audit-service/repository"
)

func (h *Handler) ListAuditEventsByCase(ctx context.Context, req *audit_service.ListAuditEventsByCaseRequest) (*audit_service.ListAuditEventsResponse, error) {
	startTime := time.Now()
	method := "ListAuditEventsByCase"
	var err error
	defer func() {
		if err != nil {
			slog.Default().ErrorContext(ctx, "handler error", slog.String("layer", currentLayer), slog.String("method", method), slog.Any("error", err), slog.Duration("duration", time.Since(startTime)))
		}
	}()

	if req.GetCaseId() == "" {
		err = status.Error(codes.InvalidArgument, "case_id is required")
		return nil, err
	}

	occurredAfter, occurredBefore := protoTime(req.OccurredAfter), protoTime(req.OccurredBefore)
	result, err := h.Service.ListByCase(ctx, req.GetCaseId(), req.GetPageSize(), req.GetPageToken(), occurredAfter, occurredBefore, req.GetOldestFirst())
	if err != nil {
		err = status.Error(codes.Internal, "failed to list audit events")
		return nil, err
	}
	return listResultToProto(result), nil
}

func (h *Handler) ListAuditEventsByObservable(ctx context.Context, req *audit_service.ListAuditEventsByObservableRequest) (*audit_service.ListAuditEventsResponse, error) {
	startTime := time.Now()
	method := "ListAuditEventsByObservable"
	var err error
	defer func() {
		if err != nil {
			slog.Default().ErrorContext(ctx, "handler error", slog.String("layer", currentLayer), slog.String("method", method), slog.Any("error", err), slog.Duration("duration", time.Since(startTime)))
		}
	}()

	if req.GetObservableId() == "" {
		err = status.Error(codes.InvalidArgument, "observable_id is required")
		return nil, err
	}

	occurredAfter, occurredBefore := protoTime(req.OccurredAfter), protoTime(req.OccurredBefore)
	result, err := h.Service.ListByObservable(ctx, req.GetObservableId(), req.GetPageSize(), req.GetPageToken(), occurredAfter, occurredBefore, req.GetOldestFirst())
	if err != nil {
		err = status.Error(codes.Internal, "failed to list audit events")
		return nil, err
	}
	return listResultToProto(result), nil
}

func (h *Handler) ListAuditEventsByEntity(ctx context.Context, req *audit_service.ListAuditEventsByEntityRequest) (*audit_service.ListAuditEventsResponse, error) {
	startTime := time.Now()
	method := "ListAuditEventsByEntity"
	var err error
	defer func() {
		if err != nil {
			slog.Default().ErrorContext(ctx, "handler error", slog.String("layer", currentLayer), slog.String("method", method), slog.Any("error", err), slog.Duration("duration", time.Since(startTime)))
		}
	}()

	if req.GetEntityType() == "" || req.GetEntityId() == "" {
		err = status.Error(codes.InvalidArgument, "entity_type and entity_id are required")
		return nil, err
	}

	occurredAfter, occurredBefore := protoTime(req.OccurredAfter), protoTime(req.OccurredBefore)
	result, err := h.Service.ListByEntity(ctx, req.GetEntityType(), req.GetEntityId(), req.GetPageSize(), req.GetPageToken(), occurredAfter, occurredBefore, req.GetOldestFirst())
	if err != nil {
		err = status.Error(codes.Internal, "failed to list audit events")
		return nil, err
	}
	return listResultToProto(result), nil
}

func (h *Handler) ListAuditEventsByActor(ctx context.Context, req *audit_service.ListAuditEventsByActorRequest) (*audit_service.ListAuditEventsResponse, error) {
	startTime := time.Now()
	method := "ListAuditEventsByActor"
	var err error
	defer func() {
		if err != nil {
			slog.Default().ErrorContext(ctx, "handler error", slog.String("layer", currentLayer), slog.String("method", method), slog.Any("error", err), slog.Duration("duration", time.Since(startTime)))
		}
	}()

	if req.GetActorUserId() == "" {
		err = status.Error(codes.InvalidArgument, "actor_user_id is required")
		return nil, err
	}

	occurredAfter, occurredBefore := protoTime(req.OccurredAfter), protoTime(req.OccurredBefore)
	result, err := h.Service.ListByActor(ctx, req.GetActorUserId(), req.GetPageSize(), req.GetPageToken(), occurredAfter, occurredBefore, req.GetOldestFirst())
	if err != nil {
		err = status.Error(codes.Internal, "failed to list audit events")
		return nil, err
	}
	return listResultToProto(result), nil
}

func (h *Handler) ListAuditEventsByCorrelationId(ctx context.Context, req *audit_service.ListAuditEventsByCorrelationIdRequest) (*audit_service.ListAuditEventsResponse, error) {
	startTime := time.Now()
	method := "ListAuditEventsByCorrelationId"
	var err error
	defer func() {
		if err != nil {
			slog.Default().ErrorContext(ctx, "handler error", slog.String("layer", currentLayer), slog.String("method", method), slog.Any("error", err), slog.Duration("duration", time.Since(startTime)))
		}
	}()

	if req.GetCorrelationId() == "" {
		err = status.Error(codes.InvalidArgument, "correlation_id is required")
		return nil, err
	}

	occurredAfter, occurredBefore := protoTime(req.OccurredAfter), protoTime(req.OccurredBefore)
	result, err := h.Service.ListByCorrelationID(ctx, req.GetCorrelationId(), req.GetPageSize(), req.GetPageToken(), occurredAfter, occurredBefore, req.GetOldestFirst())
	if err != nil {
		err = status.Error(codes.Internal, "failed to list audit events")
		return nil, err
	}
	return listResultToProto(result), nil
}

func protoTime(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil || !ts.IsValid() {
		return nil
	}
	t := ts.AsTime()
	return &t
}

func listResultToProto(result *repository.ListResult) *audit_service.ListAuditEventsResponse {
	if result == nil {
		return &audit_service.ListAuditEventsResponse{}
	}
	items := make([]*audit_service.AuditEvent, 0, len(result.Items))
	for _, row := range result.Items {
		items = append(items, rowToProto(row))
	}
	return &audit_service.ListAuditEventsResponse{
		Items:         items,
		NextPageToken: result.NextPageToken,
	}
}

func rowToProto(row *repository.AuditEventRow) *audit_service.AuditEvent {
	p := &audit_service.AuditEvent{
		Id:            row.ID,
		EventId:       row.EventID,
		EventType:     row.EventType,
		SourceService: row.SourceService,
		EntityType:    row.EntityType,
		EntityId:      row.EntityID,
		Action:        row.Action,
		OccurredAt:    timestamppb.New(row.OccurredAt),
		IngestedAt:    timestamppb.New(row.IngestedAt),
	}
	if row.ParentEntityType.Valid {
		p.ParentEntityType = row.ParentEntityType.String
	}
	if row.ParentEntityID.Valid {
		p.ParentEntityId = row.ParentEntityID.String
	}
	if row.CaseID.Valid {
		p.CaseId = row.CaseID.String
	}
	if row.ObservableID.Valid {
		p.ObservableId = row.ObservableID.String
	}
	if row.ActorUserID.Valid {
		p.ActorUserId = row.ActorUserID.String
	}
	if row.ActorName.Valid {
		p.ActorName = row.ActorName.String
	}
	if row.RequestID.Valid {
		p.RequestId = row.RequestID.String
	}
	if row.CorrelationID.Valid {
		p.CorrelationId = row.CorrelationID.String
	}
	if row.ChangeSummary.Valid {
		p.ChangeSummary = row.ChangeSummary.String
	}
	if len(row.BeforeData) > 0 {
		p.BeforeDataJson = string(row.BeforeData)
	}
	if len(row.AfterData) > 0 {
		p.AfterDataJson = string(row.AfterData)
	}
	if len(row.Metadata) > 0 {
		p.MetadataJson = string(row.Metadata)
	}
	return p
}

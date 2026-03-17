package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/servicenow/audit-service/repository"
)

// IngestEvent stores one audit event. Idempotent on event_id. Returns true if newly inserted.
func (s *Service) IngestEvent(ctx context.Context, row *repository.AuditEventRow) (inserted bool, err error) {
	inserted, err = s.Repo.InsertEvent(ctx, row)
	if err != nil {
		slog.Default().ErrorContext(ctx, "IngestEvent", slog.String("layer", currentLayer), slog.Any("error", err))
		return false, err
	}
	return inserted, nil
}

// ListByCase returns audit events for the given case ID.
func (s *Service) ListByCase(ctx context.Context, caseID string, pageSize int32, pageToken string, occurredAfter, occurredBefore *time.Time, oldestFirst bool) (*repository.ListResult, error) {
	f := repository.ListFilter{
		PageSize:       pageSize,
		PageToken:     pageToken,
		OccurredAfter:  occurredAfter,
		OccurredBefore: occurredBefore,
		OldestFirst:    oldestFirst,
	}
	result, err := s.Repo.ListByCaseID(ctx, caseID, f)
	if err != nil {
		slog.Default().ErrorContext(ctx, "ListByCase", slog.String("layer", currentLayer), slog.Any("error", err))
		return nil, err
	}
	return result, nil
}

// ListByObservable returns audit events for the given observable ID.
func (s *Service) ListByObservable(ctx context.Context, observableID string, pageSize int32, pageToken string, occurredAfter, occurredBefore *time.Time, oldestFirst bool) (*repository.ListResult, error) {
	f := repository.ListFilter{
		PageSize:       pageSize,
		PageToken:     pageToken,
		OccurredAfter:  occurredAfter,
		OccurredBefore: occurredBefore,
		OldestFirst:    oldestFirst,
	}
	return s.Repo.ListByObservableID(ctx, observableID, f)
}

// ListByEntity returns audit events for the given entity type and ID.
func (s *Service) ListByEntity(ctx context.Context, entityType, entityID string, pageSize int32, pageToken string, occurredAfter, occurredBefore *time.Time, oldestFirst bool) (*repository.ListResult, error) {
	f := repository.ListFilter{
		PageSize:       pageSize,
		PageToken:     pageToken,
		OccurredAfter:  occurredAfter,
		OccurredBefore: occurredBefore,
		OldestFirst:    oldestFirst,
	}
	return s.Repo.ListByEntity(ctx, entityType, entityID, f)
}

// ListByActor returns audit events for the given actor user ID.
func (s *Service) ListByActor(ctx context.Context, actorUserID string, pageSize int32, pageToken string, occurredAfter, occurredBefore *time.Time, oldestFirst bool) (*repository.ListResult, error) {
	f := repository.ListFilter{
		PageSize:       pageSize,
		PageToken:     pageToken,
		OccurredAfter:  occurredAfter,
		OccurredBefore: occurredBefore,
		OldestFirst:    oldestFirst,
	}
	return s.Repo.ListByActorUserID(ctx, actorUserID, f)
}

// ListByCorrelationID returns audit events for the given correlation ID.
func (s *Service) ListByCorrelationID(ctx context.Context, correlationID string, pageSize int32, pageToken string, occurredAfter, occurredBefore *time.Time, oldestFirst bool) (*repository.ListResult, error) {
	f := repository.ListFilter{
		PageSize:       pageSize,
		PageToken:     pageToken,
		OccurredAfter:  occurredAfter,
		OccurredBefore: occurredBefore,
		OldestFirst:    oldestFirst,
	}
	return s.Repo.ListByCorrelationID(ctx, correlationID, f)
}

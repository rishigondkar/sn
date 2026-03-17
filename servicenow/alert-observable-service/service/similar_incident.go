package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/org/alert-observable-service/repository"
)

// RecomputeSimilarIncidentsForCase recomputes similar_security_incidents for the given case (and symmetric pairs).
func (s *Service) RecomputeSimilarIncidentsForCase(ctx context.Context, caseID string, actor Actor) error {
	cid, err := uuid.Parse(caseID)
	if err != nil {
		return ErrValidation{Field: "case_id", Issue: "invalid uuid"}
	}
	var linked []similarIncidentLink
	err = s.TxRunner.RunInTx(ctx, func(ctx context.Context, tx *repository.Tx) error {
		// Clear existing similar incidents for this case (both directions)
		if err := tx.DeleteSimilarIncidentsByCaseID(ctx, cid); err != nil {
			return err
		}
		return recomputeSimilarIncidentsInTx(ctx, tx, cid, &linked)
	})
	if err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	for _, l := range linked {
		if s.Audit != nil {
			_ = s.Audit.Publish(context.Background(), AuditEvent{
				EventID:       uuid.Must(uuid.NewV7()).String(),
				EventType:     "similar_incident.linked",
				SourceService: "alert-observable-service",
				EntityType:    "similar_incident",
				EntityID:      l.CaseID.String() + "," + l.SimilarCaseID.String(),
				CaseID:        l.CaseID.String(),
				Action:        "link",
				ActorUserID:   actor.UserID,
				ActorName:     actor.UserName,
				RequestID:     actor.RequestID,
				CorrelationID: actor.CorrelationID,
				ChangeSummary: "similar case " + l.SimilarCaseID.String(),
				OccurredAt:    now,
			})
		}
	}
	return nil
}

// ListSimilarIncidents returns similar incidents for a case.
func (s *Service) ListSimilarIncidents(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*repository.SimilarIncident, string, error) {
	cid, err := uuid.Parse(caseID)
	if err != nil {
		return nil, "", ErrValidation{Field: "case_id", Issue: "invalid uuid"}
	}
	return s.SimilarIncidentRepo.ListByCaseID(ctx, cid, repository.ListOpts{PageSize: pageSize, PageToken: pageToken})
}

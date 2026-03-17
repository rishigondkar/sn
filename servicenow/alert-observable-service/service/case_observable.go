package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/org/alert-observable-service/repository"
)

// LinkObservableToCaseRequest for LinkObservableToCase.
type LinkObservableToCaseRequest struct {
	CaseID          string
	ObservableType  string
	ObservableValue string
	NormalizedValue string
	RoleInCase      string
	TrackingStatus  string
	IsPrimary       bool
	AddedByUserID   string
}

// LinkObservableToCase creates or resolves canonical observable, links to case (idempotent), recomputes similar incidents, publishes events.
func (s *Service) LinkObservableToCase(ctx context.Context, req *LinkObservableToCaseRequest, actor Actor) (*repository.CaseObservable, error) {
	if req.CaseID == "" {
		return nil, ErrValidation{Field: "case_id", Issue: "required"}
	}
	caseID, err := uuid.Parse(req.CaseID)
	if err != nil {
		return nil, ErrValidation{Field: "case_id", Issue: "invalid uuid"}
	}
	if req.ObservableType == "" || req.ObservableValue == "" {
		return nil, ErrValidation{Field: "observable_type/observable_value", Issue: "required"}
	}
	if !ValidateObservableType(req.ObservableType) {
		return nil, ErrValidation{Field: "observable_type", Issue: "must be a valid observable type"}
	}
	if req.TrackingStatus != "" && !ValidateTrackingStatus(req.TrackingStatus) {
		return nil, ErrValidation{Field: "tracking_status", Issue: "invalid value"}
	}
	var addedBy *uuid.UUID
	if req.AddedByUserID != "" {
		id, err := uuid.Parse(req.AddedByUserID)
		if err != nil {
			return nil, ErrValidation{Field: "added_by_user_id", Issue: "invalid uuid"}
		}
		addedBy = &id
	}

	var result *repository.CaseObservable
	var createdObservable *repository.Observable
	err = s.TxRunner.RunInTx(ctx, func(ctx context.Context, tx *repository.Tx) error {
		normalized := req.NormalizedValue
		if normalized == "" {
			normalized = NormalizeObservableValue(req.ObservableType, req.ObservableValue)
		}
		if normalized == "" {
			normalized = req.ObservableValue
		}
		// Create or get observable
		obs, err := tx.GetObservableByTypeAndNormalized(ctx, req.ObservableType, normalized)
		if err != nil {
			return err
		}
		if obs == nil {
			now := time.Now().UTC()
			obs = &repository.Observable{
				ID:              uuid.Must(uuid.NewV7()),
				ObservableType:  req.ObservableType,
				ObservableValue: req.ObservableValue,
				NormalizedValue: &normalized,
				FirstSeenTime:   &now,
				LastSeenTime:    &now,
				CreatedAt:       now,
				UpdatedAt:       now,
			}
			if err := tx.CreateObservable(ctx, obs); err != nil {
				return err
			}
			createdObservable = obs
		}
		// Idempotent: if already linked, return existing
		co, err := tx.GetCaseObservableByCaseAndObservable(ctx, caseID, obs.ID)
		if err != nil {
			return err
		}
		if co != nil {
			result = co
			return nil
		}
		now := time.Now().UTC()
		co = &repository.CaseObservable{
			ID:             uuid.Must(uuid.NewV7()),
			CaseID:         caseID,
			ObservableID:   obs.ID,
			RoleInCase:     ptrOrNil(req.RoleInCase),
			TrackingStatus: ptrOrNil(req.TrackingStatus),
			IsPrimary:      req.IsPrimary,
			AddedByUserID:  addedBy,
			AddedAt:        now,
			UpdatedAt:      now,
		}
		if err := tx.CreateCaseObservable(ctx, co); err != nil {
			return err
		}
		result = co
		// Recompute similar incidents for this case (and symmetrically for others that share this observable)
		return recomputeSimilarIncidentsInTx(ctx, tx, caseID, nil)
	})
	if err != nil {
		return nil, err
	}
	// Publish audit after successful commit
	now := time.Now().UTC().Format(time.RFC3339)
	if s.Audit != nil && createdObservable != nil {
		_ = s.Audit.Publish(context.Background(), AuditEvent{
			EventID:       uuid.Must(uuid.NewV7()).String(),
			EventType:     "observable.created",
			SourceService: "alert-observable-service",
			EntityType:    "observable",
			EntityID:      createdObservable.ID.String(),
			Action:        "create",
			ActorUserID:   actor.UserID,
			ActorName:     actor.UserName,
			RequestID:     actor.RequestID,
			CorrelationID: actor.CorrelationID,
			OccurredAt:    now,
		})
	}
	if s.Audit != nil && result != nil {
		_ = s.Audit.Publish(context.Background(), AuditEvent{
			EventID:       uuid.Must(uuid.NewV7()).String(),
			EventType:     "observable.linked.to_case",
			SourceService: "alert-observable-service",
			EntityType:    "case_observable",
			EntityID:      result.ID.String(),
			CaseID:        caseID.String(),
			ObservableID:  result.ObservableID.String(),
			Action:        "link",
			ActorUserID:   actor.UserID,
			ActorName:     actor.UserName,
			RequestID:     actor.RequestID,
			CorrelationID: actor.CorrelationID,
			OccurredAt:    now,
		})
	}
	return result, nil
}

// UpdateCaseObservable updates tracking_status, accuracy, determination, impact.
func (s *Service) UpdateCaseObservable(ctx context.Context, caseObservableID string, trackingStatus, accuracy, determination, impact *string, actor Actor) (*repository.CaseObservable, error) {
	id, err := uuid.Parse(caseObservableID)
	if err != nil {
		return nil, ErrValidation{Field: "case_observable_id", Issue: "invalid uuid"}
	}
	if trackingStatus != nil && !ValidateTrackingStatus(*trackingStatus) {
		return nil, ErrValidation{Field: "tracking_status", Issue: "invalid value"}
	}
	co, err := s.CaseObservableRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if co == nil {
		return nil, &ErrNotFound{Resource: "case_observable", ID: caseObservableID}
	}
	if trackingStatus != nil {
		co.TrackingStatus = trackingStatus
	}
	if accuracy != nil {
		co.Accuracy = accuracy
	}
	if determination != nil {
		co.Determination = determination
	}
	if impact != nil {
		co.Impact = impact
	}
	co.UpdatedAt = time.Now().UTC()
	if err := s.CaseObservableRepo.Update(ctx, co); err != nil {
		return nil, err
	}
	if s.Audit != nil {
		_ = s.Audit.Publish(context.Background(), AuditEvent{
			EventID:       uuid.Must(uuid.NewV7()).String(),
			EventType:     "observable.updated",
			SourceService: "alert-observable-service",
			EntityType:    "case_observable",
			EntityID:      co.ID.String(),
			CaseID:        co.CaseID.String(),
			ObservableID:  co.ObservableID.String(),
			Action:        "update",
			ActorUserID:   actor.UserID,
			ActorName:     actor.UserName,
			RequestID:     actor.RequestID,
			CorrelationID: actor.CorrelationID,
			OccurredAt:    co.UpdatedAt.Format(time.RFC3339),
		})
	}
	return co, nil
}

// ListCaseObservables returns case_observables for a case with optional filters.
func (s *Service) ListCaseObservables(ctx context.Context, caseID string, pageSize int32, pageToken string, filterType, filterStatus *string) ([]*repository.CaseObservable, string, error) {
	cid, err := uuid.Parse(caseID)
	if err != nil {
		return nil, "", ErrValidation{Field: "case_id", Issue: "invalid uuid"}
	}
	return s.CaseObservableRepo.ListByCaseID(ctx, cid, repository.ListOpts{PageSize: pageSize, PageToken: pageToken}, filterType, filterStatus)
}

// ListCaseObservablesWithDetails returns case observables with observable type/value from JOIN (no N+1).
func (s *Service) ListCaseObservablesWithDetails(ctx context.Context, caseID string, pageSize int32, pageToken string, filterType, filterStatus *string) ([]*repository.CaseObservableWithDetails, string, error) {
	cid, err := uuid.Parse(caseID)
	if err != nil {
		return nil, "", ErrValidation{Field: "case_id", Issue: "invalid uuid"}
	}
	return s.CaseObservableRepo.ListCaseObservablesWithDetailsByCaseID(ctx, cid, repository.ListOpts{PageSize: pageSize, PageToken: pageToken}, filterType, filterStatus)
}

// FindCasesByObservable returns case IDs that have this observable linked.
func (s *Service) FindCasesByObservable(ctx context.Context, observableID string, pageSize int32, pageToken string) ([]string, string, error) {
	id, err := uuid.Parse(observableID)
	if err != nil {
		return nil, "", ErrValidation{Field: "observable_id", Issue: "invalid uuid"}
	}
	ids, next, err := s.CaseObservableRepo.CaseIDsByObservable(ctx, id, repository.ListOpts{PageSize: pageSize, PageToken: pageToken})
	if err != nil {
		return nil, "", err
	}
	strs := make([]string, len(ids))
	for i, u := range ids {
		strs[i] = u.String()
	}
	return strs, next, nil
}

// UnlinkObservableFromCase removes the case_observable link and recomputes similar incidents.
func (s *Service) UnlinkObservableFromCase(ctx context.Context, caseID, observableID string, actor Actor) error {
	cid, err := uuid.Parse(caseID)
	if err != nil {
		return ErrValidation{Field: "case_id", Issue: "invalid uuid"}
	}
	oid, err := uuid.Parse(observableID)
	if err != nil {
		return ErrValidation{Field: "observable_id", Issue: "invalid uuid"}
	}
	err = s.TxRunner.RunInTx(ctx, func(ctx context.Context, tx *repository.Tx) error {
		if err := tx.DeleteCaseObservableByCaseAndObservable(ctx, cid, oid); err != nil {
			return err
		}
		return recomputeSimilarIncidentsInTx(ctx, tx, cid, nil)
	})
	if err != nil {
		return err
	}
	if s.Audit != nil {
		now := time.Now().UTC().Format(time.RFC3339)
		_ = s.Audit.Publish(context.Background(), AuditEvent{
			EventID:       uuid.Must(uuid.NewV7()).String(),
			EventType:     "observable.unlinked.from_case",
			SourceService: "alert-observable-service",
			EntityType:    "case_observable",
			CaseID:        caseID,
			ObservableID:  observableID,
			Action:        "unlink",
			ActorUserID:   actor.UserID,
			ActorName:     actor.UserName,
			RequestID:     actor.RequestID,
			CorrelationID: actor.CorrelationID,
			OccurredAt:    now,
		})
	}
	return nil
}

// similarIncidentLink is used to report upserted pairs for audit.
type similarIncidentLink struct {
	CaseID        uuid.UUID
	SimilarCaseID uuid.UUID
}

// recomputeSimilarIncidentsInTx aggregates shared observables by other case and upserts similar_security_incidents (both directions).
// If outLinked is non-nil, appends each (caseID, similarCaseID) pair upserted for audit publishing.
func recomputeSimilarIncidentsInTx(ctx context.Context, tx *repository.Tx, caseID uuid.UUID, outLinked *[]similarIncidentLink) error {
	observableIDs, err := tx.GetObservableIDsByCaseID(ctx, caseID)
	if err != nil {
		return err
	}
	// otherCaseID -> list of shared observable IDs
	sharedByCase := make(map[uuid.UUID][]uuid.UUID)
	for _, obsID := range observableIDs {
		otherCaseIDs, err := tx.AllCaseIDsByObservable(ctx, obsID)
		if err != nil {
			return err
		}
		for _, cid := range otherCaseIDs {
			if cid == caseID {
				continue
			}
			sharedByCase[cid] = append(sharedByCase[cid], obsID)
		}
	}
	now := time.Now().UTC()
	// Upsert A->B and B->A for each pair
	for similarCaseID, sharedIDs := range sharedByCase {
		count := len(sharedIDs)
		idsJSON, _ := json.Marshal(uuidSliceToStrings(sharedIDs))
		// Resolve observable values for display (observable_value for each shared ID)
		valuesJSON := observableValuesJSONInTx(ctx, tx, sharedIDs)
		score := float64(count)
		// A -> B
		s1 := &repository.SimilarIncident{
			ID:                     uuid.Must(uuid.NewV7()),
			CaseID:                 caseID,
			SimilarCaseID:          similarCaseID,
			MatchReason:            "shared_observable",
			SharedObservableCount:  count,
			SharedObservableIDs:    idsJSON,
			SharedObservableValues: valuesJSON,
			SimilarityScore:        &score,
			LastComputedAt:         now,
			CreatedAt:              now,
			UpdatedAt:              now,
		}
		if err := tx.UpsertSimilarIncident(ctx, s1); err != nil {
			return err
		}
		// B -> A
		s2 := &repository.SimilarIncident{
			ID:                     uuid.Must(uuid.NewV7()),
			CaseID:                 similarCaseID,
			SimilarCaseID:          caseID,
			MatchReason:            "shared_observable",
			SharedObservableCount:  count,
			SharedObservableIDs:    idsJSON,
			SharedObservableValues: valuesJSON,
			SimilarityScore:        &score,
			LastComputedAt:         now,
			CreatedAt:              now,
			UpdatedAt:              now,
		}
		if err := tx.UpsertSimilarIncident(ctx, s2); err != nil {
			return err
		}
		if outLinked != nil {
			*outLinked = append(*outLinked, similarIncidentLink{CaseID: caseID, SimilarCaseID: similarCaseID})
		}
	}
	return nil
}

func uuidSliceToStrings(ids []uuid.UUID) []string {
	s := make([]string, len(ids))
	for i, id := range ids {
		s[i] = id.String()
	}
	return s
}

// observableValuesJSONInTx returns a JSON array of observable_value strings for the given observable IDs (for similar-incident display).
func observableValuesJSONInTx(ctx context.Context, tx *repository.Tx, observableIDs []uuid.UUID) []byte {
	if len(observableIDs) == 0 {
		return nil
	}
	values := make([]string, 0, len(observableIDs))
	for _, id := range observableIDs {
		o, err := tx.GetObservableByID(ctx, id)
		if err != nil || o == nil {
			continue
		}
		if o.ObservableValue != "" {
			values = append(values, o.ObservableValue)
		} else if o.NormalizedValue != nil && *o.NormalizedValue != "" {
			values = append(values, *o.NormalizedValue)
		}
	}
	if len(values) == 0 {
		return nil
	}
	b, _ := json.Marshal(values)
	return b
}

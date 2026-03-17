package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/org/alert-observable-service/repository"
)

// CreateChildObservableRelationRequest for CreateChildObservableRelation.
type CreateChildObservableRelationRequest struct {
	ParentObservableID string
	ChildObservableID  string
	RelationshipType  string
	RelationshipDirection string
	Confidence        float64
	SourceName        string
	SourceRecordID    string
	MetadataJSON      string
}

// CreateChildObservableRelation creates a child-observable relation. Rejects self-links (parent == child).
func (s *Service) CreateChildObservableRelation(ctx context.Context, req *CreateChildObservableRelationRequest, actor Actor) (*repository.ChildObservable, error) {
	if req.ParentObservableID == "" || req.ChildObservableID == "" {
		return nil, ErrValidation{Field: "parent_observable_id/child_observable_id", Issue: "required"}
	}
	if req.RelationshipType == "" {
		return nil, ErrValidation{Field: "relationship_type", Issue: "required"}
	}
	parentID, err := uuid.Parse(req.ParentObservableID)
	if err != nil {
		return nil, ErrValidation{Field: "parent_observable_id", Issue: "invalid uuid"}
	}
	childID, err := uuid.Parse(req.ChildObservableID)
	if err != nil {
		return nil, ErrValidation{Field: "child_observable_id", Issue: "invalid uuid"}
	}
	if parentID == childID {
		return nil, ErrValidation{Field: "child_observable_id", Issue: "cannot link observable to itself"}
	}
	existing, err := s.ChildObservableRepo.GetByParentChildType(ctx, parentID, childID, req.RelationshipType)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, &ErrAlreadyExists{Resource: "child_observable", Detail: "same parent, child, and relationship_type"}
	}
	var metadata []byte
	if req.MetadataJSON != "" {
		metadata = []byte(req.MetadataJSON)
		if !json.Valid(metadata) {
			return nil, ErrValidation{Field: "metadata_json", Issue: "invalid json"}
		}
	}
	now := time.Now().UTC()
	var confidence *float64
	if req.Confidence != 0 {
		confidence = &req.Confidence
	}
	c := &repository.ChildObservable{
		ID:                   uuid.Must(uuid.NewV7()),
		ParentObservableID:   parentID,
		ChildObservableID:    childID,
		RelationshipType:     req.RelationshipType,
		RelationshipDirection: ptrOrNil(req.RelationshipDirection),
		Confidence:           confidence,
		SourceName:           ptrOrNil(req.SourceName),
		SourceRecordID:       ptrOrNil(req.SourceRecordID),
		Metadata:             metadata,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if err := s.ChildObservableRepo.Create(ctx, c); err != nil {
		return nil, err
	}
	if s.Audit != nil {
		_ = s.Audit.Publish(context.Background(), AuditEvent{
			EventID:       uuid.Must(uuid.NewV7()).String(),
			EventType:     "child_observable.created",
			SourceService: "alert-observable-service",
			EntityType:    "child_observable",
			EntityID:      c.ID.String(),
			Action:        "create",
			ActorUserID:   actor.UserID,
			ActorName:     actor.UserName,
			RequestID:     actor.RequestID,
			CorrelationID: actor.CorrelationID,
			OccurredAt:    now.Format(time.RFC3339),
		})
	}
	return c, nil
}

// ListChildObservables returns child observables for a parent.
func (s *Service) ListChildObservables(ctx context.Context, parentObservableID string, pageSize int32, pageToken string) ([]*repository.ChildObservable, string, error) {
	parentID, err := uuid.Parse(parentObservableID)
	if err != nil {
		return nil, "", ErrValidation{Field: "parent_observable_id", Issue: "invalid uuid"}
	}
	return s.ChildObservableRepo.ListByParentID(ctx, parentID, repository.ListOpts{PageSize: pageSize, PageToken: pageToken})
}

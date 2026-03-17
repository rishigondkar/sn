package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/org/alert-observable-service/repository"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateOrGetObservableRequest for CreateOrGetObservable.
type CreateOrGetObservableRequest struct {
	ObservableType  string
	ObservableValue string
	NormalizedValue string // optional; if empty we compute from type+value
	FirstSeenTime   *timestamppb.Timestamp
	LastSeenTime    *timestamppb.Timestamp
}

// CreateOrGetObservable creates a canonical observable or returns existing (by type + normalized value). Idempotent.
func (s *Service) CreateOrGetObservable(ctx context.Context, req *CreateOrGetObservableRequest) (*repository.Observable, bool, error) {
	if req.ObservableType == "" || req.ObservableValue == "" {
		return nil, false, ErrValidation{Field: "observable_type/observable_value", Issue: "required"}
	}
	if !ValidateObservableType(req.ObservableType) {
		return nil, false, ErrValidation{Field: "observable_type", Issue: "must be a valid observable type (e.g. ipv4, ipv6, domain, email, url, md5, sha256, cve, asn, file_name, registry_key, mutex, command_line)"}
	}
	normalized := req.NormalizedValue
	if normalized == "" {
		normalized = NormalizeObservableValue(req.ObservableType, req.ObservableValue)
	}
	if normalized == "" {
		normalized = req.ObservableValue // fallback for uniqueness when type has no normalization
	}
	existing, err := s.ObservableRepo.GetByTypeAndNormalized(ctx, req.ObservableType, normalized)
	if err != nil {
		return nil, false, err
	}
	if existing != nil {
		return existing, false, nil
	}
	now := time.Now().UTC()
	firstSeen := protoTimeToTime(req.FirstSeenTime)
	lastSeen := protoTimeToTime(req.LastSeenTime)
	if firstSeen == nil {
		firstSeen = &now
	}
	if lastSeen == nil {
		lastSeen = &now
	}
	o := &repository.Observable{
		ID:              uuid.Must(uuid.NewV7()),
		ObservableType:  req.ObservableType,
		ObservableValue: req.ObservableValue,
		NormalizedValue: &normalized,
		FirstSeenTime:   firstSeen,
		LastSeenTime:    lastSeen,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.ObservableRepo.Create(ctx, o); err != nil {
		return nil, false, err
	}
	return o, true, nil
}

// GetObservable returns an observable by ID.
func (s *Service) GetObservable(ctx context.Context, observableID string) (*repository.Observable, error) {
	id, err := uuid.Parse(observableID)
	if err != nil {
		return nil, ErrValidation{Field: "observable_id", Issue: "invalid uuid"}
	}
	o, err := s.ObservableRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, &ErrNotFound{Resource: "observable", ID: observableID}
	}
	return o, nil
}

// ListObservables returns observables with optional search on value/normalized_value (paginated).
func (s *Service) ListObservables(ctx context.Context, search string, pageSize int32, pageToken string) ([]*repository.Observable, string, error) {
	if pageSize <= 0 {
		pageSize = 50
	}
	var searchPtr *string
	if search != "" {
		searchPtr = &search
	}
	return s.ObservableRepo.List(ctx, repository.ListOpts{PageSize: pageSize, PageToken: pageToken}, searchPtr)
}

// UpdateObservableRequest carries optional fields to update (nil = leave unchanged).
type UpdateObservableRequest struct {
	ObservableID    string
	ObservableValue *string
	ObservableType  *string
	Finding         *string
	Notes           *string
}

// UpdateObservable updates an observable by ID. Only non-nil fields are updated.
func (s *Service) UpdateObservable(ctx context.Context, req *UpdateObservableRequest) (*repository.Observable, error) {
	id, err := uuid.Parse(req.ObservableID)
	if err != nil {
		return nil, ErrValidation{Field: "observable_id", Issue: "invalid uuid"}
	}
	o, err := s.ObservableRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, &ErrNotFound{Resource: "observable", ID: req.ObservableID}
	}
	if req.ObservableValue != nil {
		o.ObservableValue = *req.ObservableValue
	}
	if req.ObservableType != nil {
		if !ValidateObservableType(*req.ObservableType) {
			return nil, ErrValidation{Field: "observable_type", Issue: "must be a valid observable type"}
		}
		o.ObservableType = *req.ObservableType
	}
	if req.Finding != nil {
		o.Finding = req.Finding
	}
	if req.Notes != nil {
		o.Notes = req.Notes
	}
	// Recompute normalized if value or type changed
	normalized := NormalizeObservableValue(o.ObservableType, o.ObservableValue)
	if normalized == "" {
		normalized = o.ObservableValue
	}
	o.NormalizedValue = &normalized
	// Check unique (type, normalized) - another row with same type+normalized but different id would conflict
	existing, err := s.ObservableRepo.GetByTypeAndNormalized(ctx, o.ObservableType, normalized)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.ID != o.ID {
		return nil, ErrAlreadyExists{Resource: "observable", Detail: "another observable with same type and normalized value exists"}
	}
	o.UpdatedAt = time.Now().UTC()
	if err := s.ObservableRepo.Update(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}

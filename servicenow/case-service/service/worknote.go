package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/servicenow/case-service/audit"
	"github.com/servicenow/case-service/domain"
)

// AddWorknote creates a worknote and publishes worknote.created.
func (s *Service) AddWorknote(ctx context.Context, caseID, noteText, noteType, createdByUserID string, actor *ActorMeta) (*domain.Worknote, error) {
	if noteText == "" {
		return nil, errors.New("note_text is required")
	}
	c, err := s.Repo.GetCaseByID(ctx, caseID)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrNotFound
	}
	if noteType == "" {
		noteType = domain.DefaultNoteType
	}
	now := time.Now().UTC()
	w := &domain.Worknote{
		ID:              uuid.Must(uuid.NewV7()).String(),
		CaseID:          caseID,
		NoteText:        noteText,
		NoteType:        noteType,
		CreatedByUserID: createdByUserID,
		CreatedAt:       now,
		IsDeleted:       false,
	}
	if err := s.Repo.CreateWorknote(ctx, w); err != nil {
		return nil, err
	}
	evt := &audit.Event{
		EventID:       uuid.Must(uuid.NewV7()).String(),
		EventType:     audit.EventTypeWorknoteCreated,
		SourceService: audit.SourceService,
		EntityType:    "worknote",
		EntityID:      w.ID,
		ParentEntityType: "case",
		ParentEntityID:   caseID,
		CaseID:        caseID,
		Action:        "create",
		ActorUserID:   actor.UserID,
		ActorName:     actor.Name,
		RequestID:     actor.RequestID,
		CorrelationID: actor.CorrelationID,
		AfterData:     map[string]interface{}{"id": w.ID, "case_id": caseID, "note_type": noteType},
		OccurredAt:    now,
	}
	_ = s.AuditPub.Publish(context.Background(), evt)
	return w, nil
}

// ListWorknotes returns worknotes for a case with pagination.
func (s *Service) ListWorknotes(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*domain.Worknote, string, error) {
	return s.Repo.ListWorknotes(ctx, caseID, pageSize, pageToken)
}

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/soc-platform/attachment-service/audit"
	"github.com/soc-platform/attachment-service/domain"
	"github.com/soc-platform/attachment-service/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRepo struct {
	createErr  error
	get        map[string]*domain.Attachment
	list       []*domain.Attachment
	listToken  string
	softDelErr error
}

func (m *mockRepo) CreateAttachment(ctx context.Context, a *domain.Attachment) error {
	if m.createErr != nil {
		return m.createErr
	}
	if m.get == nil {
		m.get = make(map[string]*domain.Attachment)
	}
	// copy so we can mutate in tests
	dup := *a
	m.get[a.ID] = &dup
	return nil
}

func (m *mockRepo) GetByID(ctx context.Context, id string) (*domain.Attachment, error) {
	if a, ok := m.get[id]; ok {
		return a, nil
	}
	return nil, nil
}

func (m *mockRepo) ListByCaseID(ctx context.Context, caseID string, pageSize int32, pageToken string, includeDeleted bool) ([]*domain.Attachment, string, error) {
	return m.list, m.listToken, nil
}

func (m *mockRepo) SoftDelete(ctx context.Context, id string, deletedAt time.Time) error {
	return m.softDelErr
}

type mockStorage struct {
	putErr    error
	deleteErr error
}

func (m *mockStorage) Put(ctx context.Context, key, bucket string, content []byte) error {
	return m.putErr
}

func (m *mockStorage) Get(ctx context.Context, key, bucket string) ([]byte, error) {
	return nil, storage.ErrNotFound
}

func (m *mockStorage) Delete(ctx context.Context, key, bucket string) error {
	return m.deleteErr
}

type mockAudit struct {
	events []audit.Event
}

func (m *mockAudit) Publish(ctx context.Context, evt audit.Event) error {
	if m.events == nil {
		m.events = []audit.Event{}
	}
	m.events = append(m.events, evt)
	return nil
}

func TestCreateAttachment_Success(t *testing.T) {
	repo := &mockRepo{get: make(map[string]*domain.Attachment)}
	store := &mockStorage{}
	aud := &mockAudit{}
	svc := &Service{
		Repo:     repo,
		Storage:  store,
		Audit:    aud,
		MaxBytes: 1024,
	}
	ctx := context.Background()
	a, err := svc.CreateAttachment(ctx, "case-1", "file.txt", "text/plain", "user-1", []byte("hello"), "", "", "")
	require.NoError(t, err)
	require.NotNil(t, a)
	assert.NotEmpty(t, a.ID)
	assert.Equal(t, "case-1", a.CaseID)
	assert.Equal(t, "file.txt", a.FileName)
	assert.Equal(t, int64(5), a.FileSizeBytes)
	assert.Len(t, aud.events, 1)
	assert.Equal(t, "attachment.uploaded", aud.events[0].EventType)
}

func TestCreateAttachment_Validation(t *testing.T) {
	svc := &Service{Repo: &mockRepo{}, Storage: &mockStorage{}, Audit: &mockAudit{}, MaxBytes: 10}
	ctx := context.Background()

	_, err := svc.CreateAttachment(ctx, "", "f", "text/plain", "u", []byte("x"), "", "", "")
	require.Error(t, err)
	assert.True(t, errors.As(err, new(ErrValidation)))

	_, err = svc.CreateAttachment(ctx, "c", "f", "text/plain", "", []byte("x"), "", "", "")
	require.Error(t, err)
	assert.True(t, errors.As(err, new(ErrValidation)))

	_, err = svc.CreateAttachment(ctx, "c", "f", "text/plain", "u", make([]byte, 20), "", "", "")
	require.Error(t, err)
	assert.True(t, errors.As(err, new(ErrValidation)))
}

func TestCreateAttachment_StorageThenDBFailure_Compensation(t *testing.T) {
	repo := &mockRepo{createErr: errors.New("db fail"), get: make(map[string]*domain.Attachment)}
	store := storage.NewMemoryStore()
	aud := &mockAudit{}
	svc := &Service{
		Repo:     repo,
		Storage:  store,
		Audit:    aud,
		MaxBytes: 1024,
	}
	ctx := context.Background()
	_, err := svc.CreateAttachment(ctx, "case-1", "file.txt", "text/plain", "user-1", []byte("hello"), "", "", "")
	require.Error(t, err)
	assert.Len(t, aud.events, 0)
}

func TestDeleteAttachment_NotFound(t *testing.T) {
	repo := &mockRepo{get: make(map[string]*domain.Attachment)}
	svc := &Service{Repo: repo, Storage: &mockStorage{}, Audit: &mockAudit{}}
	ctx := context.Background()
	err := svc.DeleteAttachment(ctx, "missing", "", "", "", "")
	require.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
}

func TestDeleteAttachment_Success(t *testing.T) {
	a := &domain.Attachment{ID: "att-1", CaseID: "case-1", StorageKey: "k", StorageBucket: "b", IsDeleted: false}
	repo := &mockRepo{get: map[string]*domain.Attachment{"att-1": a}}
	store := &mockStorage{}
	aud := &mockAudit{}
	svc := &Service{Repo: repo, Storage: store, Audit: aud}
	ctx := context.Background()
	err := svc.DeleteAttachment(ctx, "att-1", "req-1", "corr-1", "user-1", "User One")
	require.NoError(t, err)
	assert.Len(t, aud.events, 1)
	assert.Equal(t, "attachment.deleted", aud.events[0].EventType)
}

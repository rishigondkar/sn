package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/lib/pq"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = os.Getenv("DATABASE_URL")
	}
	if url == "" {
		t.Skip("TEST_DATABASE_URL or DATABASE_URL not set")
	}
	db, err := sql.Open("postgres", url)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	require.NoError(t, db.Ping())
	return db
}

func TestInsertEvent_Idempotent(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	// Assumes audit_events table exists (run migrations/001_audit_events.up.sql first)

	repo := NewRepository(db)
	eventID := "test-event-" + uuid.New().String()
	row := &AuditEventRow{
		ID:            uuid.New().String(),
		EventID:       eventID,
		EventType:     "case.created",
		SourceService: "case-service",
		EntityType:    "case",
		EntityID:      uuid.New().String(),
		Action:        "create",
		OccurredAt:    time.Now().UTC(),
		IngestedAt:    time.Now().UTC(),
	}

	inserted, err := repo.InsertEvent(ctx, row)
	require.NoError(t, err)
	assert.True(t, inserted, "first insert should report inserted")

	// Same event_id again, different id (simulate duplicate message)
	row2 := *row
	row2.ID = uuid.New().String()
	inserted2, err := repo.InsertEvent(ctx, &row2)
	require.NoError(t, err)
	assert.False(t, inserted2, "second insert with same event_id should be idempotent (not inserted)")

	// Verify single row
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM audit_events WHERE event_id = $1", eventID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestListByCaseID_EmptyAndPagination(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	repo := NewRepository(db)

	// List by non-existent case -> empty
	result, err := repo.ListByCaseID(ctx, uuid.New().String(), ListFilter{PageSize: 10})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Items)
	assert.Empty(t, result.NextPageToken)
}

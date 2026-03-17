package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func TestPostgres_AlertRule_CRUD(t *testing.T) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("TEST_DB_DSN not set, skipping repository test")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Skipf("db ping failed (is Postgres running?): %v", err)
	}

	ctx := context.Background()
	pg := NewPostgres(db)
	now := time.Now().UTC()
	r := &AlertRule{
		ID:        uuid.Must(uuid.NewV7()),
		RuleName:  "test-rule",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := pg.CreateAlertRule(ctx, r); err != nil {
		t.Fatalf("CreateAlertRule: %v", err)
	}
	got, err := pg.GetAlertRuleByID(ctx, r.ID)
	if err != nil {
		t.Fatalf("GetAlertRuleByID: %v", err)
	}
	if got == nil || got.RuleName != r.RuleName {
		t.Errorf("got rule %+v, want RuleName %q", got, r.RuleName)
	}
}

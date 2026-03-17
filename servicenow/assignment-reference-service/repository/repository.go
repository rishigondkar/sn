package repository

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

// User is the domain model for a user row.
type User struct {
	ID          string
	Username    string
	Email       string
	DisplayName string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Group is the domain model for an assignment group row.
type Group struct {
	ID          string
	GroupName   string
	Description string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// GroupMember is the domain model for a group_members row.
type GroupMember struct {
	ID        string
	GroupID   string
	UserID    string
	MemberRole string
	CreatedAt time.Time
}

// ListUsersFilter holds optional filters for listing users.
type ListUsersFilter struct {
	ActiveOnly       bool
	FilterDisplayName string
	FilterUsername   string
	FilterEmail      string
}

// ListGroupsFilter holds optional filters for listing groups.
type ListGroupsFilter struct {
	ActiveOnly        bool
	FilterGroupName   string
}

// PageToken is used for cursor-based pagination (offset stored as opaque token).
type PageToken struct {
	Offset int
}

// Repository provides data access for users, groups, and group members.
type Repository struct {
	db        *sql.DB
	dbTimeout time.Duration
}

// New creates a Repository with the given DB and timeout.
func New(db *sql.DB, dbTimeout time.Duration) *Repository {
	return &Repository{db: db, dbTimeout: dbTimeout}
}

// WithTx runs fn inside a transaction. Context timeout is applied to the transaction.
func (r *Repository) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	txCtx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()
	tx, err := r.db.BeginTx(txCtx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

// Ping checks DB connectivity (for health).
func (r *Repository) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()
	return r.db.PingContext(ctx)
}

// SeedPOCIfEmpty inserts POC seed data (users, assignment_groups, group_members) if assignment_groups is empty.
// Idempotent: uses ON CONFLICT DO NOTHING so safe to run multiple times.
func (r *Repository) SeedPOCIfEmpty(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()
	var n int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM assignment_groups").Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	// Same as migrations/002_seed_poc.sql - keep in sync (one statement per Exec)
	if _, err := r.db.ExecContext(ctx, `INSERT INTO users (id, username, email, display_name, is_active, created_at, updated_at)
VALUES ('a0000000-0000-4000-8000-000000000001', 'soc-admin', 'soc-admin@example.com', 'SOC Admin', true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING`); err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, `INSERT INTO assignment_groups (id, group_name, description, is_active, created_at, updated_at)
VALUES
  ('b0000000-0000-4000-8000-000000000001', 'Triage', 'Default triage assignment group', true, NOW(), NOW()),
  ('b0000000-0000-4000-8000-000000000002', 'SOC L1', 'Security Operations Center Level 1', true, NOW(), NOW()),
  ('b0000000-0000-4000-8000-000000000003', 'SOC L2', 'Security Operations Center Level 2', true, NOW(), NOW()),
  ('b0000000-0000-4000-8000-000000000004', 'CSIRT', 'Computer Security Incident Response Team', true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING`); err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, `INSERT INTO group_members (id, group_id, user_id, member_role, created_at)
VALUES
  ('c0000000-0000-4000-8000-000000000001', 'b0000000-0000-4000-8000-000000000001', 'a0000000-0000-4000-8000-000000000001', 'member', NOW()),
  ('c0000000-0000-4000-8000-000000000002', 'b0000000-0000-4000-8000-000000000002', 'a0000000-0000-4000-8000-000000000001', 'member', NOW()),
  ('c0000000-0000-4000-8000-000000000003', 'b0000000-0000-4000-8000-000000000003', 'a0000000-0000-4000-8000-000000000001', 'member', NOW()),
  ('c0000000-0000-4000-8000-000000000004', 'b0000000-0000-4000-8000-000000000004', 'a0000000-0000-4000-8000-000000000001', 'member', NOW())
ON CONFLICT (group_id, user_id) DO NOTHING`); err != nil {
		return err
	}
	return nil
}

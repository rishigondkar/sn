package repository

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/google/uuid"
)

const defaultPageSize = 50
const maxPageSize = 100

// GetUserByID returns a user by ID, or nil if not found.
func (r *Repository) GetUserByID(ctx context.Context, id string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()
	var u User
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, email, display_name, is_active, created_at, updated_at
		 FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Username, &u.Email, &u.DisplayName, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// ListUsers returns a page of users and the next page token (empty if no more).
func (r *Repository) ListUsers(ctx context.Context, pageSize int, pageToken string, f ListUsersFilter) ([]*User, string, error) {
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	offset := 0
	if pageToken != "" {
		if pt := parseOffsetToken(pageToken); pt != nil {
			offset = pt.Offset
		}
	}

	ctx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()

	query := `SELECT id, username, email, display_name, is_active, created_at, updated_at FROM users WHERE 1=1`
	args := []interface{}{}
	argNum := 1
	if f.ActiveOnly {
		query += ` AND is_active = true`
	}
	if f.FilterDisplayName != "" {
		query += ` AND display_name ILIKE $` + strconv.Itoa(argNum)
		args = append(args, "%"+f.FilterDisplayName+"%")
		argNum++
	}
	if f.FilterUsername != "" {
		query += ` AND username ILIKE $` + strconv.Itoa(argNum)
		args = append(args, "%"+f.FilterUsername+"%")
		argNum++
	}
	if f.FilterEmail != "" {
		query += ` AND email ILIKE $` + strconv.Itoa(argNum)
		args = append(args, "%"+f.FilterEmail+"%")
		argNum++
	}
	query += ` ORDER BY display_name, id LIMIT $` + strconv.Itoa(argNum) + ` OFFSET $` + strconv.Itoa(argNum+1)
	args = append(args, pageSize+1, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var items []*User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.DisplayName, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, "", err
		}
		items = append(items, &u)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	nextToken := ""
	if len(items) > pageSize {
		items = items[:pageSize]
		nextToken = encodeOffsetToken(offset + pageSize)
	}
	return items, nextToken, nil
}

// CreateUser inserts a new user. ID must be set (UUID).
func (r *Repository) CreateUser(ctx context.Context, u *User) error {
	ctx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, username, email, display_name, is_active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		u.ID, u.Username, u.Email, u.DisplayName, u.IsActive, u.CreatedAt, u.UpdatedAt,
	)
	return err
}

// UpdateUser updates an existing user by ID.
func (r *Repository) UpdateUser(ctx context.Context, u *User) error {
	ctx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()
	result, err := r.db.ExecContext(ctx,
		`UPDATE users SET username=$2, email=$3, display_name=$4, is_active=$5, updated_at=$6 WHERE id=$1`,
		u.ID, u.Username, u.Email, u.DisplayName, u.IsActive, u.UpdatedAt,
	)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// UserExistsByUsername returns true if a user with the given username exists (optionally excluding id).
func (r *Repository) UserExistsByUsername(ctx context.Context, username, excludeID string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()
	var count int
	q := `SELECT COUNT(*) FROM users WHERE username = $1`
	args := []interface{}{username}
	if excludeID != "" {
		q += ` AND id != $2`
		args = append(args, excludeID)
	}
	err := r.db.QueryRowContext(ctx, q, args...).Scan(&count)
	return count > 0, err
}

// UserExistsByEmail returns true if a user with the given email exists (optionally excluding id).
func (r *Repository) UserExistsByEmail(ctx context.Context, email, excludeID string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()
	var count int
	q := `SELECT COUNT(*) FROM users WHERE email = $1`
	args := []interface{}{email}
	if excludeID != "" {
		q += ` AND id != $2`
		args = append(args, excludeID)
	}
	err := r.db.QueryRowContext(ctx, q, args...).Scan(&count)
	return count > 0, err
}

// UserExistsByID returns true if a user with the given ID exists.
func (r *Repository) UserExistsByID(ctx context.Context, id string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE id = $1`, id).Scan(&count)
	return count > 0, err
}

// CreateUserWithID generates a new UUID and inserts the user.
func (r *Repository) CreateUserWithID(ctx context.Context, username, email, displayName string, isActive bool) (*User, error) {
	id := uuid.New().String()
	now := time.Now().UTC()
	u := &User{
		ID: id, Username: username, Email: email, DisplayName: displayName,
		IsActive: isActive, CreatedAt: now, UpdatedAt: now,
	}
	if err := r.CreateUser(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

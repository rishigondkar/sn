package repository

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// GetGroupByID returns a group by ID, or nil if not found.
func (r *Repository) GetGroupByID(ctx context.Context, id string) (*Group, error) {
	ctx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()
	var g Group
	err := r.db.QueryRowContext(ctx,
		`SELECT id, group_name, description, is_active, created_at, updated_at
		 FROM assignment_groups WHERE id = $1`,
		id,
	).Scan(&g.ID, &g.GroupName, &g.Description, &g.IsActive, &g.CreatedAt, &g.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}

// ListGroups returns a page of groups and the next page token (empty if no more).
func (r *Repository) ListGroups(ctx context.Context, pageSize int, pageToken string, f ListGroupsFilter) ([]*Group, string, error) {
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

	query := `SELECT id, group_name, description, is_active, created_at, updated_at FROM assignment_groups WHERE 1=1`
	args := []interface{}{}
	argNum := 1
	if f.ActiveOnly {
		query += ` AND is_active = true`
	}
	if f.FilterGroupName != "" {
		query += ` AND group_name ILIKE $` + strconv.Itoa(argNum)
		args = append(args, "%"+f.FilterGroupName+"%")
		argNum++
	}
	query += ` ORDER BY group_name, id LIMIT $` + strconv.Itoa(argNum) + ` OFFSET $` + strconv.Itoa(argNum+1)
	args = append(args, pageSize+1, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var items []*Group
	for rows.Next() {
		var g Group
		if err := rows.Scan(&g.ID, &g.GroupName, &g.Description, &g.IsActive, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, "", err
		}
		items = append(items, &g)
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

// CreateGroup inserts a new group. ID must be set (UUID).
func (r *Repository) CreateGroup(ctx context.Context, g *Group) error {
	ctx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO assignment_groups (id, group_name, description, is_active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		g.ID, g.GroupName, g.Description, g.IsActive, g.CreatedAt, g.UpdatedAt,
	)
	return err
}

// UpdateGroup updates an existing group by ID.
func (r *Repository) UpdateGroup(ctx context.Context, g *Group) error {
	ctx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()
	result, err := r.db.ExecContext(ctx,
		`UPDATE assignment_groups SET group_name=$2, description=$3, is_active=$4, updated_at=$5 WHERE id=$1`,
		g.ID, g.GroupName, g.Description, g.IsActive, g.UpdatedAt,
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

// GroupExistsByName returns true if a group with the given name exists (optionally excluding id).
func (r *Repository) GroupExistsByName(ctx context.Context, groupName, excludeID string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()
	var count int
	q := `SELECT COUNT(*) FROM assignment_groups WHERE group_name = $1`
	args := []interface{}{groupName}
	if excludeID != "" {
		q += ` AND id != $2`
		args = append(args, excludeID)
	}
	err := r.db.QueryRowContext(ctx, q, args...).Scan(&count)
	return count > 0, err
}

// GroupExistsByID returns true if a group with the given ID exists.
func (r *Repository) GroupExistsByID(ctx context.Context, id string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM assignment_groups WHERE id = $1`, id).Scan(&count)
	return count > 0, err
}

// CreateGroupWithID generates a new UUID and inserts the group.
func (r *Repository) CreateGroupWithID(ctx context.Context, groupName, description string, isActive bool) (*Group, error) {
	id := uuid.New().String()
	now := time.Now().UTC()
	g := &Group{
		ID: id, GroupName: groupName, Description: description,
		IsActive: isActive, CreatedAt: now, UpdatedAt: now,
	}
	if err := r.CreateGroup(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

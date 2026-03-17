package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ListGroupMembers returns a page of members for a group and the next page token (empty if no more).
func (r *Repository) ListGroupMembers(ctx context.Context, groupID string, pageSize int, pageToken string) ([]*GroupMember, string, error) {
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

	query := `SELECT id, group_id, user_id, member_role, created_at FROM group_members WHERE group_id = $1 ORDER BY created_at, id LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, groupID, pageSize+1, offset)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var items []*GroupMember
	for rows.Next() {
		var m GroupMember
		if err := rows.Scan(&m.ID, &m.GroupID, &m.UserID, &m.MemberRole, &m.CreatedAt); err != nil {
			return nil, "", err
		}
		items = append(items, &m)
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

// AddGroupMember inserts a group member. Fails if (group_id, user_id) already exists (unique constraint).
func (r *Repository) AddGroupMember(ctx context.Context, groupID, userID, memberRole string) (*GroupMember, error) {
	id := uuid.New().String()
	now := time.Now().UTC()
	ctx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO group_members (id, group_id, user_id, member_role, created_at) VALUES ($1, $2, $3, $4, $5)`,
		id, groupID, userID, memberRole, now,
	)
	if err != nil {
		return nil, err
	}
	return &GroupMember{
		ID: id, GroupID: groupID, UserID: userID, MemberRole: memberRole, CreatedAt: now,
	}, nil
}

// RemoveGroupMember deletes the membership for (group_id, user_id). Returns true if a row was deleted.
func (r *Repository) RemoveGroupMember(ctx context.Context, groupID, userID string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM group_members WHERE group_id = $1 AND user_id = $2`,
		groupID, userID,
	)
	if err != nil {
		return false, err
	}
	n, _ := result.RowsAffected()
	return n > 0, nil
}

// GroupMemberExists returns true if (group_id, user_id) already exists.
func (r *Repository) GroupMemberExists(ctx context.Context, groupID, userID string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.dbTimeout)
	defer cancel()
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM group_members WHERE group_id = $1 AND user_id = $2`,
		groupID, userID,
	).Scan(&count)
	return count > 0, err
}

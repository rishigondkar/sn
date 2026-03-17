package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	r := New(db, 5*time.Second)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .+ FROM users WHERE id`).
		WithArgs("nonexistent").
		WillReturnRows(sqlmock.NewRows([]string{}))

	u, err := r.GetUserByID(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, u)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByID_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	r := New(db, 5*time.Second)
	ctx := context.Background()
	id := uuid.New().String()
	now := time.Now().UTC()

	mock.ExpectQuery(`SELECT .+ FROM users WHERE id`).
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "display_name", "is_active", "created_at", "updated_at"}).
			AddRow(id, "u1", "u1@test.com", "User One", true, now, now))

	u, err := r.GetUserByID(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, u)
	assert.Equal(t, id, u.ID)
	assert.Equal(t, "u1", u.Username)
	assert.Equal(t, "u1@test.com", u.Email)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserExistsByID_Exists(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	r := New(db, 5*time.Second)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT`).WithArgs("some-id").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	exists, err := r.UserExistsByID(ctx, "some-id")
	require.NoError(t, err)
	assert.True(t, exists)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserExistsByID_NotExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	r := New(db, 5*time.Second)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT`).WithArgs("missing").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	exists, err := r.UserExistsByID(ctx, "missing")
	require.NoError(t, err)
	assert.False(t, exists)
	require.NoError(t, mock.ExpectationsWereMet())
}

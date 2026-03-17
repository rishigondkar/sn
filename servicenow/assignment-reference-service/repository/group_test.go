package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetGroupByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	r := New(db, 5*time.Second)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .+ FROM assignment_groups WHERE id`).
		WithArgs("nonexistent").
		WillReturnRows(sqlmock.NewRows([]string{}))

	g, err := r.GetGroupByID(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, g)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGroupExistsByID_Exists(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	r := New(db, 5*time.Second)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT`).WithArgs("group-id").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	exists, err := r.GroupExistsByID(ctx, "group-id")
	require.NoError(t, err)
	assert.True(t, exists)
	require.NoError(t, mock.ExpectationsWereMet())
}

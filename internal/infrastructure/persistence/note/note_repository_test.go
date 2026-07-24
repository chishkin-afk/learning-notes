package notepersistence

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/chishkin-afk/learning_notes/internal/domain/note"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRepo(t *testing.T) (*noteRepository, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(nil, nil))
	repo := New(logger, db)

	cleanup := func() {
		require.NoError(t, mock.ExpectationsWereMet())
		db.Close()
	}

	return repo, mock, cleanup
}

func createTestNote(t *testing.T) *note.Note {
	t.Helper()
	n, err := note.New("Valid Title", "Valid Description", note.NewPriority(1))
	require.NoError(t, err)
	return n
}

func TestNoteRepository_Save(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr error
	}{
		{
			name: "success",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO notes")).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedErr: nil,
		},
		{
			name: "error_note_already_exists",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO notes")).
					WillReturnError(&pgconn.PgError{Code: "23505", Message: "duplicate key"})
			},
			expectedErr: note.ErrNoteAlreadyExists,
		},
		{
			name: "error_db_failure",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta("INSERT INTO notes")).
					WillReturnError(errors.New("connection refused"))
			},
			expectedErr: errors.New("connection refused"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo, mock, cleanup := setupRepo(t)
			defer cleanup()

			tt.setupMock(mock)

			n := createTestNote(t)
			err := repo.Save(context.Background(), n)

			if tt.expectedErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNoteRepository_GetByID(t *testing.T) {
	t.Parallel()

	testID := uuid.New()
	now := time.Now().UTC()

	tests := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr error
	}{
		{
			name: "success",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "title", "description", "status_int", "priority_int", "created_at", "updated_at",
				}).AddRow(testID, "title", "desc", 1, 1, now, now)

				mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
					WithArgs(testID).
					WillReturnRows(rows)
			},
			expectedErr: nil,
		},
		{
			name: "error_not_found",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
					WithArgs(testID).
					WillReturnError(sql.ErrNoRows)
			},
			expectedErr: note.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo, mock, cleanup := setupRepo(t)
			defer cleanup()

			tt.setupMock(mock)

			res, err := repo.GetByID(context.Background(), testID)

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, res)
			} else {
				require.NoError(t, err)
				require.NotNil(t, res)
				assert.Equal(t, testID, res.ID())
			}
		})
	}
}

func TestNoteRepository_List(t *testing.T) {
	t.Parallel()

	testID := uuid.New()
	now := time.Now().UTC()

	tests := []struct {
		name        string
		page        uint32
		limit       uint32
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr error
	}{
		{
			name:  "success",
			page:  1,
			limit: 10,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "title", "description", "status_int", "priority_int", "created_at", "updated_at",
				}).AddRow(testID, "title1", "desc1", 1, 1, now, now)

				mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
					WillReturnRows(rows)
			},
			expectedErr: nil,
		},
		{
			name:  "error_db_failure",
			page:  1,
			limit: 10,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
					WillReturnError(errors.New("db timeout"))
			},
			expectedErr: errors.New("db timeout"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo, mock, cleanup := setupRepo(t)
			defer cleanup()

			tt.setupMock(mock)

			res, err := repo.List(context.Background(), tt.page, tt.limit)

			if tt.expectedErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, res, 1)
			}
		})
	}
}

func TestNoteRepository_Update(t *testing.T) {
	t.Parallel()

	testID := uuid.New()
	now := time.Now().UTC()

	tests := []struct {
		name        string
		updateFunc  note.UpdateFunc
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr error
	}{
		{
			name: "success",
			updateFunc: func(ctx context.Context, n *note.Note) error {
				return n.UpdateTitle("New Title")
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				rows := sqlmock.NewRows([]string{
					"id", "title", "description", "status_int", "priority_int", "created_at", "updated_at",
				}).AddRow(testID, "old", "old", 1, 1, now, now)

				mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
					WithArgs(testID).
					WillReturnRows(rows)

				mock.ExpectExec(regexp.QuoteMeta("UPDATE notes")).
					WillReturnResult(sqlmock.NewResult(0, 1))

				mock.ExpectCommit()
			},
			expectedErr: nil,
		},
		{
			name: "error_domain_validation_in_update_func",
			updateFunc: func(ctx context.Context, n *note.Note) error {
				return errors.New("domain validation failed")
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				rows := sqlmock.NewRows([]string{
					"id", "title", "description", "status_int", "priority_int", "created_at", "updated_at",
				}).AddRow(testID, "old", "old", 1, 1, now, now)

				mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
					WithArgs(testID).
					WillReturnRows(rows)

				mock.ExpectRollback()
			},
			expectedErr: errors.New("domain validation failed"),
		},
		{
			name: "error_not_found_during_lock",
			updateFunc: func(ctx context.Context, n *note.Note) error {
				return nil
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
					WithArgs(testID).
					WillReturnError(sql.ErrNoRows)

				mock.ExpectRollback()
			},
			expectedErr: note.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo, mock, cleanup := setupRepo(t)
			defer cleanup()

			tt.setupMock(mock)

			res, err := repo.Update(context.Background(), testID, tt.updateFunc)

			if tt.expectedErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, res)
			}
		})
	}
}

func TestNoteRepository_Delete(t *testing.T) {
	t.Parallel()

	testID := uuid.New()

	tests := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock)
		expectedErr error
	}{
		{
			name: "success",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta("DELETE FROM notes")).
					WithArgs(testID.String()).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedErr: nil,
		},
		{
			name: "error_not_found_zero_rows_affected",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta("DELETE FROM notes")).
					WithArgs(testID.String()).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedErr: note.ErrNotFound,
		},
		{
			name: "error_db_failure",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta("DELETE FROM notes")).
					WillReturnError(errors.New("foreign key constraint"))
			},
			expectedErr: errors.New("foreign key constraint"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo, mock, cleanup := setupRepo(t)
			defer cleanup()

			tt.setupMock(mock)

			err := repo.Delete(context.Background(), testID)

			if tt.expectedErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

package notecache

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"log/slog"

	"github.com/alicebob/miniredis/v2"
	"github.com/chishkin-afk/learning_notes/internal/domain/note"
	"github.com/chishkin-afk/learning_notes/internal/infrastructure/config"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRepo(t *testing.T) (*noteRepository, *miniredis.Miniredis, func()) {
	t.Helper()

	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	cfg := &config.Config{
		Cache: config.Cache{
			NoteTTL: 5 * time.Second,
		},
	}

	logger := slog.New(slog.NewTextHandler(nil, nil))
	repo := New(cfg, logger, client)

	cleanup := func() {
		client.Close()
		mr.Close()
	}

	return repo, mr, cleanup
}

func createTestNote(t *testing.T) *note.Note {
	t.Helper()
	n, err := note.New("Valid Title", "Valid Description", note.NewPriority(2))
	require.NoError(t, err)
	return n
}

func TestNoteRepository_Save(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		expectedErr error
	}{
		{
			name:        "success",
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo, mr, cleanup := setupRepo(t)
			defer cleanup()

			n := createTestNote(t)
			err := repo.Save(context.Background(), n)

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)

				key := getNoteKey(n.ID())
				cachedValue, err := mr.Get(key)
				require.NoError(t, err)

				var record noteRecord
				err = json.Unmarshal([]byte(cachedValue), &record)
				require.NoError(t, err)

				assert.Equal(t, n.ID(), record.ID)
				assert.Equal(t, n.Title(), record.Title)
			}
		})
	}
}

func TestNoteRepository_GetByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupCache  func(mr *miniredis.Miniredis, id uuid.UUID)
		expectedErr error
	}{
		{
			name: "success",
			setupCache: func(mr *miniredis.Miniredis, id uuid.UUID) {
				n, _ := note.New("Cached Title", "Cached Desc", note.NewPriority(2))
				bytes, _ := noteToBytes(n)
				mr.Set(getNoteKey(id), string(bytes))
			},
			expectedErr: nil,
		},
		{
			name: "error_not_found",
			setupCache: func(mr *miniredis.Miniredis, id uuid.UUID) {
			},
			expectedErr: note.ErrNotFound,
		},
		{
			name: "error_invalid_json",
			setupCache: func(mr *miniredis.Miniredis, id uuid.UUID) {
				mr.Set(getNoteKey(id), "invalid json data")
			},
			expectedErr: json.Unmarshal([]byte("invalid"), &struct{}{}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo, mr, cleanup := setupRepo(t)
			defer cleanup()

			testID := uuid.New()
			if tt.setupCache != nil {
				tt.setupCache(mr, testID)
			}

			res, err := repo.GetByID(context.Background(), testID)

			if tt.expectedErr != nil {
				require.Error(t, err)
				if tt.name != "error_invalid_json" {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				assert.Nil(t, res)
			} else {
				require.NoError(t, err)
				require.NotNil(t, res)
			}
		})
	}
}

func TestNoteRepository_Delete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupCache  func(mr *miniredis.Miniredis, id uuid.UUID)
		expectedErr error
	}{
		{
			name: "success",
			setupCache: func(mr *miniredis.Miniredis, id uuid.UUID) {
				n, _ := note.New("To Delete", "Desc", note.NewPriority(2))
				bytes, _ := noteToBytes(n)
				mr.Set(getNoteKey(id), string(bytes))
			},
			expectedErr: nil,
		},
		{
			name: "error_not_found_zero_deleted",
			setupCache: func(mr *miniredis.Miniredis, id uuid.UUID) {
			},
			expectedErr: note.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo, mr, cleanup := setupRepo(t)
			defer cleanup()

			testID := uuid.New()
			if tt.setupCache != nil {
				tt.setupCache(mr, testID)
			}

			err := repo.Delete(context.Background(), testID)

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				exists := mr.Exists(getNoteKey(testID))
				assert.False(t, exists, "key must be removed from cache")
			}
		})
	}
}

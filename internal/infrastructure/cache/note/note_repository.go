package notecache

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/chishkin-afk/learning_notes/internal/domain/note"
	"github.com/chishkin-afk/learning_notes/internal/infrastructure/config"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Set(ctx context.Context, key string, value any, duration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, key ...string) *redis.IntCmd
}

type noteRepository struct {
	cfg   *config.Config
	log   *slog.Logger
	cache Cache
}

// New is a constructor for note cache repository.
//
// It simply creates a new Redis-based repository
// for caching notes using the provided configuration,
// logger, and cache client.
func New(cfg *config.Config, log *slog.Logger, cache Cache) *noteRepository {
	return &noteRepository{
		cfg:   cfg,
		log:   log,
		cache: cache,
	}
}

// Save stores a Note in the cache.
//
// This method serializes the given note and stores it
// in Redis using the configured cache TTL.
// Cache-specific errors are translated into domain errors
// before being returned to the caller.
func (nr *noteRepository) Save(ctx context.Context, note *note.Note) error {
	nr.log.Debug("saving note into cache",
		slog.String("note_id", note.ID().String()),
	)

	if err := nr.save(ctx, nr.cache, note); err != nil {
		return fmt.Errorf("failed to save note into cache: %w",
			handleError(err),
		)
	}

	return nil
}

func (nr *noteRepository) save(ctx context.Context, cache Cache, note *note.Note) error {
	bytes, err := noteToBytes(note)
	if err != nil {
		return err
	}

	key := getNoteKey(note.ID())
	return cache.Set(ctx, key, bytes, nr.cfg.Cache.NoteTTL).Err()
}

// GetByID retrieves a Note from the cache by its identifier.
//
// This method looks up the cached note with the given ID,
// deserializes it into the domain model, and returns it.
// Cache-specific errors are translated into domain errors
// before being returned to the caller.
func (nr *noteRepository) GetByID(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	nr.log.Debug("getting note from cache",
		slog.String("note_id", id.String()),
	)

	note, err := nr.getByID(ctx, nr.cache, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get note from cache: %w",
			handleError(err),
		)
	}

	return note, nil
}

func (nr *noteRepository) getByID(ctx context.Context, cache Cache, id uuid.UUID) (*note.Note, error) {
	bytes, err := cache.Get(ctx, getNoteKey(id)).Bytes()
	if err != nil {
		return nil, err
	}

	return bytesToNote(bytes)
}

// Delete removes a Note from the cache by its identifier.
//
// This method removes the cached note with the given ID.
// Cache-specific errors are translated into domain errors
// before being returned to the caller.
func (nr *noteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	nr.log.Debug("deleting note from cache",
		slog.String("note_id", id.String()),
	)

	if err := nr.delete(ctx, nr.cache, id); err != nil {
		return fmt.Errorf("failed to delete note from cache: %w",
			handleError(err),
		)
	}

	return nil
}

func (nr *noteRepository) delete(ctx context.Context, cache Cache, id uuid.UUID) error {
	result := cache.Del(ctx, getNoteKey(id))
	if err := result.Err(); err != nil {
		return err
	}

	if result.Val() == 0 {
		return note.ErrNotFound
	}

	return nil
}

func getNoteKey(id uuid.UUID) string {
	return fmt.Sprintf("note:%s", id.String())
}

func handleError(err error) error {
	if errors.Is(err, redis.Nil) {
		return note.ErrNotFound
	}

	return err
}

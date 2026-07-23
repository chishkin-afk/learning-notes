package notepersistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Masterminds/squirrel"
	"github.com/chishkin-afk/learning_notes/internal/domain/note"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

var noteColumns = []string{
	"id",
	"title",
	"description",
	"status_int",
	"priority_int",
	"created_at",
	"updated_at",
}

// DB is an interface for connection
//
// An interface for the base database structure
// any sql.DB can be used with it.
type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// WithTx is an interface for tx connection
//
// This interface implements sql.DB,
// but unlike the DB interface, it includes a BeginTx method.
type WithTx interface {
	DB
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type scanner interface {
	Scan(dest ...any) error
}

type noteRepository struct {
	log        *slog.Logger
	db         WithTx
	sqlBuilder squirrel.StatementBuilderType
}

// New is a constructor for note persistence repository
//
// It simply creates a new persistence repository
// for the note using the input arguments.
func New(log *slog.Logger, db WithTx) *noteRepository {
	return &noteRepository{
		log:        log,
		db:         db,
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save stores a Note in the persistent storage.
//
// This method inserts the given note into the database.
// If the note already exists, it returns a domain-specific error.
// Any database-specific errors are translated into domain errors
// before being returned to the caller.
func (nr *noteRepository) Save(ctx context.Context, note *note.Note) error {
	nr.log.Debug("saving note into db",
		slog.String("note_id", note.ID().String()),
	)

	if err := nr.save(ctx, nr.db, note); err != nil {
		return fmt.Errorf("failed to save note into db: %w",
			handleError(err),
		)
	}

	return nil
}

func (nr *noteRepository) save(ctx context.Context, db DB, note *note.Note) error {
	query, args, err := nr.buildSaveQuery(noteToRecord(note))
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		return err
	}

	return nil
}

func (nr *noteRepository) buildSaveQuery(record *noteRecord) (string, []any, error) {
	return nr.sqlBuilder.Insert("notes").Columns(noteColumns...).Values(
		record.ID,
		record.Title,
		record.Description,
		record.StatusInt,
		record.PriorityInt,
		record.CreatedAt,
		record.UpdatedAt,
	).ToSql()
}

// GetByID retrieves a Note from persistent storage by its identifier.
//
// This method searches for a note with the given ID and restores
// the domain model from the stored record.
// Database-specific errors are translated into domain errors
// before being returned to the caller.
func (nr *noteRepository) GetByID(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	nr.log.Debug("getting note from db",
		slog.String("note_id", id.String()),
	)

	note, err := nr.getByID(ctx, nr.db, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get note from db: %w",
			handleError(err),
		)
	}

	return note, nil
}

func (nr *noteRepository) getByID(ctx context.Context, db DB, id uuid.UUID) (*note.Note, error) {
	query, args, err := nr.buildGetByIDQuery(id)
	if err != nil {
		return nil, err
	}

	record, err := mapRow(db.QueryRowContext(ctx, query, args...))
	if err != nil {
		return nil, err
	}

	note, warn := recordToNote(record)
	if warn != nil {
		nr.log.Warn("error with restore note from db",
			slog.String("warn", warn.Error()),
		)
	}

	return note, nil
}

func (nr *noteRepository) buildGetByIDQuery(id uuid.UUID) (string, []any, error) {
	return nr.sqlBuilder.Select(noteColumns...).From("notes").Limit(1).Where("id = ?", id).ToSql()
}

// List retrieves a paginated list of Notes from persistent storage.
//
// This method returns notes using the specified page number and limit.
// Notes are ordered by creation time in descending order.
// Database-specific errors are translated into domain errors
// before being returned to the caller.
func (nr *noteRepository) List(ctx context.Context, page, limit uint32) ([]*note.Note, error) {
	nr.log.Debug("listing notes from db",
		slog.Int("page", int(page)),
		slog.Int("limit", int(limit)),
	)

	list, err := nr.list(ctx, nr.db, page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list notes from db: %w",
			handleError(err),
		)
	}

	return list, nil
}

func (nr *noteRepository) list(ctx context.Context, db DB, page, limit uint32) ([]*note.Note, error) {
	query, args, err := nr.buildListQuery(page, limit)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records, err := mapRows(rows)
	if err != nil {
		return nil, err
	}

	return nr.recordsToNotes(records), nil
}

func (nr *noteRepository) buildListQuery(page, limit uint32) (string, []any, error) {
	page = max(page, 1)
	offset := uint64(limit) * uint64((page - 1))
	return nr.sqlBuilder.Select(noteColumns...).
		From("notes").
		Limit(uint64(limit)).Offset(offset).
		OrderBy("created_at DESC").ToSql()
}

// Update updates a Note in persistent storage using the provided update function.
//
// This method starts a transaction, locks the note row for update,
// applies the given update function to the domain model, and persists
// the resulting state back to the database.
//
// The update function is responsible for applying domain-level changes
// to the Note. If the update function or any database operation fails,
// the transaction is rolled back and the error is returned.
//
// Database-specific errors are translated into domain errors
// before being returned to the caller.
func (nr *noteRepository) Update(ctx context.Context, id uuid.UUID, uf note.UpdateFunc) (*note.Note, error) {
	nr.log.Debug("start tx for update",
		slog.String("note_id", id.String()),
	)

	tx, rollback, err := nr.beginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start tx: %w", err)
	}
	defer rollback()

	noteForUpdate, err := nr.getForUpdate(ctx, tx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get note for update: %w",
			handleError(err),
		)
	}

	nr.log.Debug("executing update function",
		slog.String("note_id", noteForUpdate.ID().String()),
	)

	if err := uf(ctx, noteForUpdate); err != nil {
		return nil, err
	}

	nr.log.Debug("updating note in db",
		slog.String("note_id", noteForUpdate.ID().String()),
	)

	if err := nr.update(ctx, tx, noteForUpdate); err != nil {
		return nil, fmt.Errorf("failed to update note: %w",
			handleError(err),
		)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit tx: %w", err)
	}

	return noteForUpdate, nil
}

func (nr *noteRepository) update(ctx context.Context, db DB, updNote *note.Note) error {
	query, args, err := nr.buildUpdateQuery(noteToRecord(updNote))
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		return err
	}

	return nil
}

func (nr *noteRepository) buildUpdateQuery(record *noteRecord) (string, []any, error) {
	return nr.sqlBuilder.Update("notes").SetMap(map[string]any{
		"status_int":   record.StatusInt,
		"priority_int": record.PriorityInt,
		"title":        record.Title,
		"description":  record.Description,
	}).Where("id = ?", record.ID).ToSql()
}

func (nr *noteRepository) getForUpdate(ctx context.Context, db DB, id uuid.UUID) (*note.Note, error) {
	query, args, err := nr.buildGetForUpdate(id)
	if err != nil {
		return nil, err
	}

	record, err := mapRow(db.QueryRowContext(ctx, query, args...))
	if err != nil {
		return nil, err
	}

	note, warn := recordToNote(record)
	if warn != nil {
		nr.log.Warn("error with restore note from db",
			slog.String("warn", warn.Error()),
		)
	}

	return note, nil
}

func (nr *noteRepository) buildGetForUpdate(id uuid.UUID) (string, []any, error) {
	return nr.sqlBuilder.Select(noteColumns...).
		From("notes").Where("id = ?", id).Suffix("FOR UPDATE").ToSql()
}

// Delete removes a Note from persistent storage by its identifier.
//
// This method deletes the note with the given ID from the database.
// If the database operation fails, the error is returned to the caller.
// Database-specific errors are translated into domain errors
// before being returned.
func (nr *noteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	nr.log.Debug("deleting note from db",
		slog.String("note_id", id.String()),
	)

	if err := nr.delete(ctx, nr.db, id); err != nil {
		return fmt.Errorf("failed to delete note from db: %w",
			handleError(err),
		)
	}

	return nil
}

func (nr *noteRepository) delete(ctx context.Context, db DB, id uuid.UUID) error {
	query, args, err := nr.buildDeleteQuery(id)
	if err != nil {
		return err
	}

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	if count, _ := result.RowsAffected(); count == 0 {
		return note.ErrNotFound
	}

	return nil
}

func (nr *noteRepository) buildDeleteQuery(id uuid.UUID) (string, []any, error) {
	return nr.sqlBuilder.Delete("notes").Where("id = ?", id.String()).ToSql()
}

func (nr *noteRepository) recordsToNotes(records []*noteRecord) []*note.Note {
	notes := make([]*note.Note, len(records))
	for i, record := range records {
		note, warn := recordToNote(record)
		if warn != nil {
			nr.log.Warn("error with restore note from db",
				slog.String("warn", warn.Error()),
			)
		}

		notes[i] = note
	}

	return notes
}

func (nr *noteRepository) beginTx(ctx context.Context) (*sql.Tx, func(), error) {
	tx, err := nr.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
	if err != nil {
		return nil, nil, err
	}

	return tx, func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			nr.log.Error("failed to rollback tx",
				slog.String("error", err.Error()),
			)
		}
	}, nil
}

func mapRows(rows *sql.Rows) ([]*noteRecord, error) {
	var records []*noteRecord
	for rows.Next() {
		record, err := scanRecord(rows)
		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func mapRow(row *sql.Row) (*noteRecord, error) {
	return scanRecord(row)
}

func scanRecord(scn scanner) (*noteRecord, error) {
	var record noteRecord
	if err := scn.Scan(
		&record.ID,
		&record.Title,
		&record.Description,
		&record.StatusInt,
		&record.PriorityInt,
		&record.CreatedAt,
		&record.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &record, nil
}

func handleError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return note.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return note.ErrNoteAlreadyExists
		}
	}

	return err
}

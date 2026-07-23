package note

import (
	"context"
	"errors"
)

type (
	UpdateFunc func(ctx context.Context, updNote *Note) error
)

type NotePersistenceRepository interface {
}

var (
	ErrNoteAlreadyExists = errors.New("note already exists")
	ErrNotFound          = errors.New("note not found")
)

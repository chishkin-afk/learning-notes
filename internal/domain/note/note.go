package note

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	// Max-min values of len of note title
	MaxTitle = 128
	MinTitle = 3

	// Max-min values of len of note description
	MaxDescription = 512
	MinDescription = 0
)

var (
	// Validation errors
	ErrInvalidTitle       = errors.New("invalid title of task")
	ErrInvalidDescription = errors.New("invalid description")

	// Business error
	ErrIsntPending = errors.New("note isn't pending")
)

// Note is a main domain model of notes
//
// This structure describes the basic
// behavior of the service's domain object
// for example, updating the title
// will result in an update to updated_at.
type Note struct {
	id          uuid.UUID
	title       string
	description string
	status      Status
	priority    Priority
	createdAt   time.Time
	updatedAt   time.Time
}

// New is a constructor for domain model Note
//
// This constructor accepts initial user values ​​to create a Note,
// validates them, and returns the structure, or otherwise an error.
func New(
	title string,
	description string,
	priority Priority,
) (*Note, error) {
	title = strings.TrimSpace(title)
	if err := validateTitle(title); err != nil {
		return nil, err
	}

	description = strings.TrimSpace(description)
	if err := validateDescription(description); err != nil {
		return nil, err
	}

	if err := priority.Validate(); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &Note{
		id:          uuid.New(),
		title:       title,
		description: description,
		status:      NewStatus(StatusPending),
		priority:    priority,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// Done changes status to StatusDone
func (n *Note) Done() error {
	if !n.status.EqualsStatus(StatusPending) {
		return ErrIsntPending
	}

	n.status = NewStatus(StatusDone)
	n.touch()

	return nil
}

// Cancel changes status to StatusCanceled
func (n *Note) Cancel() error {
	if !n.status.EqualsStatus(StatusPending) {
		return ErrIsntPending
	}

	n.status = NewStatus(StatusCanceled)
	n.touch()

	return nil
}

// UpdateTitle changes title to new user title
func (n *Note) UpdateTitle(title string) error {
	title = strings.TrimSpace(title)
	if err := validateTitle(title); err != nil {
		return err
	}

	n.title = title
	n.touch()

	return nil
}

// UpdateDescription changes description to new user description
func (n *Note) UpdateDescription(description string) error {
	description = strings.TrimSpace(description)
	if err := validateDescription(description); err != nil {
		return err
	}

	n.description = description
	n.touch()

	return nil
}

// UpdatePriority changes priority to new user priority
func (n *Note) UpdatePriority(priority Priority) {
	n.priority = priority
	n.touch()
}

// ID returns id of Note
func (n *Note) ID() uuid.UUID {
	return n.id
}

// Title returns title of Note
func (n *Note) Title() string {
	return n.title
}

// Description returns description of Note
func (n *Note) Description() string {
	return n.description
}

// Status returns status of Note
func (n *Note) Status() Status {
	return n.status
}

// Priority returns priority of Note
func (n *Note) Priority() Priority {
	return n.priority
}

// CreatedAt returns created time of Note
func (n *Note) CreatedAt() time.Time {
	return n.createdAt
}

// UpdatedAt returns last updated time of Note
func (n *Note) UpdatedAt() time.Time {
	return n.updatedAt
}

func (n *Note) touch() {
	n.updatedAt = time.Now().UTC()
}

func validateTitle(title string) error {
	n := len([]rune(title))
	if n > MaxTitle || n < MinTitle {
		return fmt.Errorf("%w: %d", ErrInvalidTitle, n)
	}

	return nil
}

func validateDescription(description string) error {
	n := len([]rune(description))
	if n > MaxDescription || n < MinDescription {
		return fmt.Errorf("%w: %d", ErrInvalidDescription, n)
	}

	return nil
}

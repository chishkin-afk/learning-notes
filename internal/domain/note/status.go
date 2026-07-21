package note

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidStatus = errors.New("invalid status of task")
)

type status int

func (s status) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusDone:
		return "done"
	case StatusCanceled:
		return "canceled"
	}

	return "unknown"
}

func (s status) Int() int {
	return int(s)
}

const (
	StatusUnknown status = iota
	StatusPending
	StatusDone
	StatusCanceled
)

// Status is a value-object of note
//
// This structure manages the status of the task itself
// for example, it could be "pending," "done," or something similar.
type Status struct {
	status status
}

func (s Status) String() string {
	return s.status.String()
}

func (s Status) Int() int {
	return s.status.Int()
}

func (s Status) EqualsStatus(other status) bool {
	return s.status == other
}

func (s Status) Validate() error {
	if s.status <= StatusUnknown || s.status > StatusCanceled {
		return ErrInvalidStatus
	}

	return nil
}

// NewStatus is a constructor for Status vo
//
// This constructor does not check or validate the parameter p in any way,
// since the type is private and the user cannot specify an invalid type.
func NewStatus(s status) Status {
	return Status{
		status: s,
	}
}

// FromStatus is a constructor for Status vo
//
// This constructor accepts a raw status value and validates it
// if it is invalid, it returns StatusUnknown and an error.
func FromStatus(raw int) (Status, error) {
	s := status(raw)
	if s <= StatusUnknown || s > StatusCanceled {
		return Status{
			status: StatusUnknown,
		}, fmt.Errorf("%w: %d", ErrInvalidStatus, raw)
	}

	return Status{
		status: s,
	}, nil
}

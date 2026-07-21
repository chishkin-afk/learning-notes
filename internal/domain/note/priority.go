package note

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidPriority = errors.New("invalid priority of task")
)

type priority int

func (p priority) String() string {
	switch p {
	case PriorityLow:
		return "low"
	case PriorityMedium:
		return "medium"
	case PriorityHigh:
		return "high"
	}

	return "unknown"
}

func (p priority) Int() int {
	return int(p)
}

const (
	PriorityUnknown priority = iota
	PriorityLow
	PriorityMedium
	PriorityHigh
)

// Priority is a value-object of Note
//
// This structure is responsible for the task priority
// that is, low, high, or medium.
type Priority struct {
	priority priority
}

func (p Priority) String() string {
	return p.priority.String()
}

func (p Priority) Int() int {
	return p.priority.Int()
}

func (p Priority) Equals(other Priority) bool {
	return p.priority == other.priority
}

func (p Priority) Validate() error {
	if p.priority <= PriorityUnknown || p.priority > PriorityHigh {
		return ErrInvalidPriority
	}

	return nil
}

// NewPriority is a constructor for Priority VO
//
// A constructor whose safety relies on accepting
// an argument of the private type priority.
func NewPriority(p priority) Priority {
	return Priority{
		priority: p,
	}
}

// FromPriority is a constructor for Priority VO
//
// The constructor checks the raw priority value
// if it is valid, it returns the structure
// otherwise, it returns an error indicating an unknown priority.
func FromPriority(raw int) (Priority, error) {
	p := priority(raw)
	if p <= PriorityUnknown || p > PriorityHigh {
		return Priority{
			priority: PriorityUnknown,
		}, fmt.Errorf("%w: %d", ErrInvalidPriority, raw)
	}

	return Priority{
		priority: p,
	}, nil
}

package notepersistence

import (
	"time"

	"github.com/google/uuid"
)

type noteRecord struct {
	ID          uuid.UUID
	Title       string
	Description string
	StatusInt   int
	PriorityInt int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

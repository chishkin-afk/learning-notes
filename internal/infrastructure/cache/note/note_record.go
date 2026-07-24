package notecache

import (
	"time"

	"github.com/google/uuid"
)

type noteRecord struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StatusInt   int       `json:"status_int"`
	PriorityInt int       `json:"priority_int"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

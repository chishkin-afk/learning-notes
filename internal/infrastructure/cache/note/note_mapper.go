package notecache

import (
	"encoding/json"

	"github.com/chishkin-afk/learning_notes/internal/domain/note"
)

func noteToBytes(note *note.Note) ([]byte, error) {
	return json.Marshal(noteRecord{
		ID:          note.ID(),
		Title:       note.Title(),
		Description: note.Description(),
		StatusInt:   note.Status().Int(),
		PriorityInt: note.Priority().Int(),
		CreatedAt:   note.CreatedAt(),
		UpdatedAt:   note.UpdatedAt(),
	})
}

func bytesToNote(bytes []byte) (*note.Note, error) {
	var (
		record noteRecord
		errs   []error
	)

	if err := json.Unmarshal(bytes, &record); err != nil {
		return nil, err
	}

	status, err := note.FromStatus(record.StatusInt)
	if err != nil {
		errs = append(errs, err)
	}

	priority, err := note.FromPriority(record.PriorityInt)
	if err != nil {
		errs = append(errs, err)
	}

	return note.Restore(
		record.ID,
		record.Title,
		record.Description,
		status,
		priority,
		record.CreatedAt,
		record.UpdatedAt,
	)
}

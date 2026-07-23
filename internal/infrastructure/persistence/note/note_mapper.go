package notepersistence

import "github.com/chishkin-afk/learning_notes/internal/domain/note"

func noteToRecord(note *note.Note) *noteRecord {
	return &noteRecord{
		ID:          note.ID(),
		Title:       note.Title(),
		Description: note.Description(),
		StatusInt:   note.Status().Int(),
		PriorityInt: note.Priority().Int(),
		CreatedAt:   note.CreatedAt(),
		UpdatedAt:   note.UpdatedAt(),
	}
}

func recordToNote(record *noteRecord) (*note.Note, error) {
	var errs []error

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

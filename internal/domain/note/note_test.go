package note

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		note, err := New(
			"  My task  ",
			"  My description  ",
			NewPriority(PriorityHigh),
		)

		require.NoError(t, err)
		require.NotNil(t, note)

		assert.NotEqual(t, uuid.Nil, note.ID())
		assert.Equal(t, "My task", note.Title())
		assert.Equal(t, "My description", note.Description())
		assert.True(t, note.Status().EqualsStatus(StatusPending))
		assert.True(t, note.Priority().Equals(NewPriority(PriorityHigh)))

		assert.False(t, note.CreatedAt().IsZero())
		assert.False(t, note.UpdatedAt().IsZero())
		assert.Equal(t, note.CreatedAt(), note.UpdatedAt())
	})

	t.Run("invalid title", func(t *testing.T) {
		_, err := New(
			"",
			"description",
			NewPriority(PriorityLow),
		)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidTitle)
	})

	t.Run("invalid description", func(t *testing.T) {
		description := strings.Repeat("a", MaxDescription+1)

		_, err := New(
			"title",
			description,
			NewPriority(PriorityLow),
		)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidDescription)
	})
}

func TestNote_Done(t *testing.T) {
	note, err := New(
		"title",
		"description",
		NewPriority(PriorityMedium),
	)
	require.NoError(t, err)

	oldUpdated := note.UpdatedAt()

	time.Sleep(time.Millisecond)

	err = note.Done()

	require.NoError(t, err)

	assert.True(t, note.Status().EqualsStatus(StatusDone))
	assert.True(t, note.UpdatedAt().After(oldUpdated))
}

func TestNote_Done_NotPending(t *testing.T) {
	note, err := New(
		"title",
		"description",
		NewPriority(PriorityLow),
	)
	require.NoError(t, err)

	require.NoError(t, note.Done())

	err = note.Done()

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrIsntPending)
}

func TestNote_Cancel(t *testing.T) {
	note, err := New(
		"title",
		"description",
		NewPriority(PriorityLow),
	)
	require.NoError(t, err)

	oldUpdated := note.UpdatedAt()

	time.Sleep(time.Millisecond)

	err = note.Cancel()

	require.NoError(t, err)

	assert.True(t, note.Status().EqualsStatus(StatusCanceled))
	assert.True(t, note.UpdatedAt().After(oldUpdated))
}

func TestNote_Cancel_NotPending(t *testing.T) {
	note, err := New(
		"title",
		"description",
		NewPriority(PriorityLow),
	)
	require.NoError(t, err)

	require.NoError(t, note.Cancel())

	err = note.Cancel()

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrIsntPending)
}

func TestNote_UpdateTitle(t *testing.T) {
	note, err := New(
		"title",
		"description",
		NewPriority(PriorityLow),
	)
	require.NoError(t, err)

	oldUpdated := note.UpdatedAt()

	time.Sleep(time.Millisecond)

	err = note.UpdateTitle("  new title  ")

	require.NoError(t, err)

	assert.Equal(t, "new title", note.Title())
	assert.True(t, note.UpdatedAt().After(oldUpdated))
}

func TestNote_UpdateTitle_Invalid(t *testing.T) {
	note, err := New(
		"title",
		"description",
		NewPriority(PriorityLow),
	)
	require.NoError(t, err)

	oldUpdated := note.UpdatedAt()

	err = note.UpdateTitle("")

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidTitle)

	assert.Equal(t, "title", note.Title())
	assert.Equal(t, oldUpdated, note.UpdatedAt())
}

func TestNote_UpdateDescription(t *testing.T) {
	note, err := New(
		"title",
		"description",
		NewPriority(PriorityLow),
	)
	require.NoError(t, err)

	oldUpdated := note.UpdatedAt()

	time.Sleep(time.Millisecond)

	err = note.UpdateDescription("new description")

	require.NoError(t, err)

	assert.Equal(t, "new description", note.Description())
	assert.True(t, note.UpdatedAt().After(oldUpdated))
}

func TestNote_UpdateDescription_Invalid(t *testing.T) {
	note, err := New(
		"title",
		"description",
		NewPriority(PriorityLow),
	)
	require.NoError(t, err)

	oldUpdated := note.UpdatedAt()

	description := strings.Repeat("a", MaxDescription+1)

	err = note.UpdateDescription(description)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidDescription)

	assert.Equal(t, "description", note.Description())
	assert.Equal(t, oldUpdated, note.UpdatedAt())
}

func TestNote_UpdatePriority(t *testing.T) {
	note, err := New(
		"title",
		"description",
		NewPriority(PriorityLow),
	)
	require.NoError(t, err)

	oldUpdated := note.UpdatedAt()

	time.Sleep(time.Millisecond)

	note.UpdatePriority(NewPriority(PriorityHigh))

	assert.True(t, note.Priority().Equals(NewPriority(PriorityHigh)))
	assert.True(t, note.UpdatedAt().After(oldUpdated))
}

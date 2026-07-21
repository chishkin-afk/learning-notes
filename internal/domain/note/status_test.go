package note

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatus_String(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
		expected string
	}{
		{
			name:     "pending",
			status:   NewStatus(StatusPending),
			expected: "pending",
		},
		{
			name:     "done",
			status:   NewStatus(StatusDone),
			expected: "done",
		},
		{
			name:     "canceled",
			status:   NewStatus(StatusCanceled),
			expected: "canceled",
		},
		{
			name:     "unknown",
			status:   NewStatus(StatusUnknown),
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}

func TestStatus_Int(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
		expected int
	}{
		{
			name:     "pending",
			status:   NewStatus(StatusPending),
			expected: 1,
		},
		{
			name:     "done",
			status:   NewStatus(StatusDone),
			expected: 2,
		},
		{
			name:     "canceled",
			status:   NewStatus(StatusCanceled),
			expected: 3,
		},
		{
			name:     "unknown",
			status:   NewStatus(StatusUnknown),
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.Int())
		})
	}
}

func TestFromStatus(t *testing.T) {
	tests := []struct {
		name      string
		raw       int
		want      Status
		wantError bool
	}{
		{
			name:      "pending",
			raw:       1,
			want:      NewStatus(StatusPending),
			wantError: false,
		},
		{
			name:      "done",
			raw:       2,
			want:      NewStatus(StatusDone),
			wantError: false,
		},
		{
			name:      "canceled",
			raw:       3,
			want:      NewStatus(StatusCanceled),
			wantError: false,
		},
		{
			name:      "unknown",
			raw:       0,
			wantError: true,
		},
		{
			name:      "negative",
			raw:       -1,
			wantError: true,
		},
		{
			name:      "too large",
			raw:       4,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromStatus(tt.raw)

			if tt.wantError {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidStatus)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.String(), got.String())
			assert.Equal(t, tt.want.Int(), got.Int())
		})
	}
}

func TestFromStatus_ErrorContainsValue(t *testing.T) {
	_, err := FromStatus(42)

	require.Error(t, err)

	assert.ErrorIs(t, err, ErrInvalidStatus)
	assert.True(t, errors.Is(err, ErrInvalidStatus))
	assert.Equal(t, "invalid status of task: 42", err.Error())
}

func TestStatus_ZeroValue(t *testing.T) {
	var s Status

	assert.Equal(t, "unknown", s.String())
	assert.Equal(t, 0, s.Int())
}

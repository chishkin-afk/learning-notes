package note

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPriority_String(t *testing.T) {
	tests := []struct {
		name     string
		priority Priority
		expected string
	}{
		{
			name:     "low",
			priority: NewPriority(PriorityLow),
			expected: "low",
		},
		{
			name:     "medium",
			priority: NewPriority(PriorityMedium),
			expected: "medium",
		},
		{
			name:     "high",
			priority: NewPriority(PriorityHigh),
			expected: "high",
		},
		{
			name:     "unknown",
			priority: NewPriority(PriorityUnknown),
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.priority.String())
		})
	}
}

func TestPriority_Int(t *testing.T) {
	tests := []struct {
		name     string
		priority Priority
		expected int
	}{
		{"low", NewPriority(PriorityLow), 1},
		{"medium", NewPriority(PriorityMedium), 2},
		{"high", NewPriority(PriorityHigh), 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.priority.Int())
		})
	}
}

func TestPriority_Equals(t *testing.T) {
	low1 := NewPriority(PriorityLow)
	low2 := NewPriority(PriorityLow)
	high := NewPriority(PriorityHigh)

	assert.True(t, low1.Equals(low2))
	assert.False(t, low1.Equals(high))
}

func TestFromPriority(t *testing.T) {
	tests := []struct {
		name      string
		raw       int
		want      Priority
		wantError bool
	}{
		{
			name:      "low",
			raw:       1,
			want:      NewPriority(PriorityLow),
			wantError: false,
		},
		{
			name:      "medium",
			raw:       2,
			want:      NewPriority(PriorityMedium),
			wantError: false,
		},
		{
			name:      "high",
			raw:       3,
			want:      NewPriority(PriorityHigh),
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
			got, err := FromPriority(tt.raw)

			if tt.wantError {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidPriority)
				return
			}

			require.NoError(t, err)

			assert.True(t, got.Equals(tt.want))
			assert.Equal(t, tt.want.String(), got.String())
			assert.Equal(t, tt.want.Int(), got.Int())
		})
	}
}

func TestFromPriority_ErrorContainsValue(t *testing.T) {
	_, err := FromPriority(42)

	require.Error(t, err)

	assert.ErrorIs(t, err, ErrInvalidPriority)
	assert.True(t, errors.Is(err, ErrInvalidPriority))
	assert.Equal(
		t,
		"invalid priority of task: 42",
		err.Error(),
	)
}

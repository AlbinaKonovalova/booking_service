package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewResource_ValidName(t *testing.T) {
	r, err := NewResource("Room A")
	require.NoError(t, err)
	assert.Equal(t, "Room A", r.Name)
	assert.NotEmpty(t, r.ID)
	assert.False(t, r.CreatedAt.IsZero())
	assert.Nil(t, r.RemovedAt)
}

func TestNewResource_TrimmedName(t *testing.T) {
	r, err := NewResource("  Room B  ")
	require.NoError(t, err)
	assert.Equal(t, "Room B", r.Name)
}

func TestNewResource_EmptyName(t *testing.T) {
	_, err := NewResource("")
	assert.ErrorIs(t, err, ErrResourceNameEmpty)
}

func TestNewResource_WhitespaceName(t *testing.T) {
	_, err := NewResource("   ")
	assert.ErrorIs(t, err, ErrResourceNameEmpty)
}

func TestResource_Remove(t *testing.T) {
	r, _ := NewResource("Room")
	assert.False(t, r.IsRemoved())

	err := r.Remove()
	require.NoError(t, err)
	assert.True(t, r.IsRemoved())
	assert.NotNil(t, r.RemovedAt)
}

func TestResource_RemoveTwice(t *testing.T) {
	r, _ := NewResource("Room")
	_ = r.Remove()

	err := r.Remove()
	assert.ErrorIs(t, err, ErrResourceAlreadyRemoved)
}

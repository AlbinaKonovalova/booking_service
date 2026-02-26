package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBookingStatus_IsActive(t *testing.T) {
	assert.True(t, StatusCreated.IsActive())
	assert.True(t, StatusConfirmed.IsActive())
	assert.False(t, StatusCancelled.IsActive())
	assert.False(t, StatusExpired.IsActive())
}

func TestBookingStatus_String(t *testing.T) {
	assert.Equal(t, "CREATED", StatusCreated.String())
	assert.Equal(t, "CONFIRMED", StatusConfirmed.String())
	assert.Equal(t, "CANCELLED", StatusCancelled.String())
	assert.Equal(t, "EXPIRED", StatusExpired.String())
}

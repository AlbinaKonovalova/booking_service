package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateD_AfterNoon(t *testing.T) {
	tz := time.UTC
	startTime := time.Date(2026, 6, 15, 14, 0, 0, 0, tz)
	D := CalculateD(startTime, tz)
	assert.Equal(t, time.Date(2026, 6, 15, 0, 0, 0, 0, tz), D)
}

func TestCalculateD_ExactlyNoon(t *testing.T) {
	tz := time.UTC
	startTime := time.Date(2026, 6, 15, 12, 0, 0, 0, tz)
	D := CalculateD(startTime, tz)
	assert.Equal(t, time.Date(2026, 6, 15, 0, 0, 0, 0, tz), D)
}

func TestCalculateD_BeforeNoon(t *testing.T) {
	tz := time.UTC
	startTime := time.Date(2026, 6, 15, 1, 30, 0, 0, tz)
	D := CalculateD(startTime, tz)
	assert.Equal(t, time.Date(2026, 6, 14, 0, 0, 0, 0, tz), D)
}

func TestCalculateD_1159(t *testing.T) {
	tz := time.UTC
	startTime := time.Date(2026, 6, 15, 11, 59, 0, 0, tz)
	D := CalculateD(startTime, tz)
	assert.Equal(t, time.Date(2026, 6, 14, 0, 0, 0, 0, tz), D)
}

func TestCalculateD_Midnight(t *testing.T) {
	tz := time.UTC
	startTime := time.Date(2026, 6, 15, 0, 0, 0, 0, tz)
	D := CalculateD(startTime, tz)
	assert.Equal(t, time.Date(2026, 6, 14, 0, 0, 0, 0, tz), D)
}

func TestValidateCheckInWindow_1200_OK(t *testing.T) {
	tz := time.UTC
	D := time.Date(2026, 6, 15, 0, 0, 0, 0, tz)
	startTime := time.Date(2026, 6, 15, 12, 0, 0, 0, tz)

	err := ValidateCheckInWindow(startTime, D, tz)
	assert.NoError(t, err)
}

func TestValidateCheckInWindow_1400_OK(t *testing.T) {
	tz := time.UTC
	D := time.Date(2026, 6, 15, 0, 0, 0, 0, tz)
	startTime := time.Date(2026, 6, 15, 14, 0, 0, 0, tz)

	err := ValidateCheckInWindow(startTime, D, tz)
	assert.NoError(t, err)
}

func TestValidateCheckInWindow_0159_OK(t *testing.T) {
	tz := time.UTC
	D := time.Date(2026, 6, 15, 0, 0, 0, 0, tz)
	startTime := time.Date(2026, 6, 16, 1, 59, 0, 0, tz)

	err := ValidateCheckInWindow(startTime, D, tz)
	assert.NoError(t, err)
}

func TestValidateCheckInWindow_0200_Fail(t *testing.T) {
	tz := time.UTC
	D := time.Date(2026, 6, 15, 0, 0, 0, 0, tz)
	startTime := time.Date(2026, 6, 16, 2, 0, 0, 0, tz)

	err := ValidateCheckInWindow(startTime, D, tz)
	assert.ErrorIs(t, err, ErrBookingNotAvailable)
}

func TestValidateCheckInWindow_1159_Fail(t *testing.T) {
	tz := time.UTC
	D := time.Date(2026, 6, 15, 0, 0, 0, 0, tz)
	startTime := time.Date(2026, 6, 15, 11, 59, 0, 0, tz)

	err := ValidateCheckInWindow(startTime, D, tz)
	assert.ErrorIs(t, err, ErrBookingNotAvailable)
}

func TestCalculateEndTime_N1(t *testing.T) {
	tz := time.UTC
	D := time.Date(2026, 6, 15, 0, 0, 0, 0, tz)
	checkOut := time.Date(2026, 6, 16, 10, 0, 0, 0, tz)

	endTime, err := CalculateEndTime(D, checkOut, tz)
	require.NoError(t, err)
	assert.Equal(t, time.Date(2026, 6, 16, 12, 0, 0, 0, tz), endTime)
}

func TestCalculateEndTime_N2(t *testing.T) {
	tz := time.UTC
	D := time.Date(2026, 6, 15, 0, 0, 0, 0, tz)
	checkOut := time.Date(2026, 6, 16, 12, 1, 0, 0, tz)

	endTime, err := CalculateEndTime(D, checkOut, tz)
	require.NoError(t, err)
	assert.Equal(t, time.Date(2026, 6, 17, 12, 0, 0, 0, tz), endTime)
}

func TestCalculateEndTime_ExactBoundary(t *testing.T) {
	tz := time.UTC
	D := time.Date(2026, 6, 15, 0, 0, 0, 0, tz)
	checkOut := time.Date(2026, 6, 16, 12, 0, 0, 0, tz)

	endTime, err := CalculateEndTime(D, checkOut, tz)
	require.NoError(t, err)
	assert.Equal(t, time.Date(2026, 6, 16, 12, 0, 0, 0, tz), endTime)
}

func TestCalculateEndTime_TooLong(t *testing.T) {
	tz := time.UTC
	D := time.Date(2026, 6, 15, 0, 0, 0, 0, tz)
	checkOut := time.Date(2028, 6, 15, 12, 1, 0, 0, tz)

	_, err := CalculateEndTime(D, checkOut, tz)
	assert.ErrorIs(t, err, ErrBookingTooLong)
}

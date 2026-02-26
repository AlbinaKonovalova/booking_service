package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var utc = time.UTC

func futureTime(h, m int) time.Time {
	return time.Date(2026, 6, 15, h, m, 0, 0, utc)
}

func TestNewBooking_HappyPath(t *testing.T) {
	now := time.Date(2026, 6, 15, 10, 0, 0, 0, utc)
	checkIn := futureTime(14, 0)
	checkOut := futureTime(20, 0)

	b, err := NewBooking(uuid.New(), checkIn, checkOut, utc, now)
	require.NoError(t, err)
	assert.Equal(t, StatusCreated, b.Status)
	assert.Equal(t, checkIn.UTC(), b.StartTime)
	expectedEnd := time.Date(2026, 6, 16, 12, 0, 0, 0, utc)
	assert.Equal(t, expectedEnd, b.EndTime)
}

func TestNewBooking_CheckInAfterCheckOut(t *testing.T) {
	now := time.Date(2026, 6, 15, 10, 0, 0, 0, utc)
	checkIn := futureTime(16, 0)
	checkOut := futureTime(14, 0)

	_, err := NewBooking(uuid.New(), checkIn, checkOut, utc, now)
	assert.ErrorIs(t, err, ErrBookingCheckInAfterCheckOut)
}

func TestNewBooking_CheckInEqualsCheckOut(t *testing.T) {
	now := time.Date(2026, 6, 15, 10, 0, 0, 0, utc)
	checkIn := futureTime(14, 0)

	_, err := NewBooking(uuid.New(), checkIn, checkIn, utc, now)
	assert.ErrorIs(t, err, ErrBookingCheckInAfterCheckOut)
}

func TestNewBooking_InPast(t *testing.T) {
	now := time.Date(2026, 6, 15, 15, 0, 0, 0, utc)
	checkIn := time.Date(2026, 6, 15, 14, 0, 0, 0, utc)
	checkOut := time.Date(2026, 6, 16, 10, 0, 0, 0, utc)

	_, err := NewBooking(uuid.New(), checkIn, checkOut, utc, now)
	assert.ErrorIs(t, err, ErrBookingInPast)
}

func TestNewBooking_CheckInAtNow(t *testing.T) {
	now := time.Date(2026, 6, 15, 14, 0, 0, 0, utc)
	checkIn := now
	checkOut := time.Date(2026, 6, 16, 10, 0, 0, 0, utc)

	_, err := NewBooking(uuid.New(), checkIn, checkOut, utc, now)
	assert.ErrorIs(t, err, ErrBookingInPast)
}

func TestNewBooking_NotAvailable_TooLate(t *testing.T) {
	now := time.Date(2026, 6, 15, 3, 0, 0, 0, utc)
	checkIn := time.Date(2026, 6, 15, 5, 0, 0, 0, utc)
	checkOut := time.Date(2026, 6, 16, 10, 0, 0, 0, utc)

	_, err := NewBooking(uuid.New(), checkIn, checkOut, utc, now)
	assert.ErrorIs(t, err, ErrBookingNotAvailable)
}

func TestNewBooking_EdgeCheckIn_0159_OK(t *testing.T) {
	now := time.Date(2026, 6, 15, 0, 0, 0, 0, utc)
	checkIn := time.Date(2026, 6, 15, 1, 59, 0, 0, utc)
	checkOut := time.Date(2026, 6, 15, 10, 0, 0, 0, utc)

	b, err := NewBooking(uuid.New(), checkIn, checkOut, utc, now)
	require.NoError(t, err)
	assert.Equal(t, time.Date(2026, 6, 15, 12, 0, 0, 0, utc), b.EndTime)
}

func TestNewBooking_EdgeCheckIn_0200_NotAvailable(t *testing.T) {
	now := time.Date(2026, 6, 15, 0, 0, 0, 0, utc)
	checkIn := time.Date(2026, 6, 15, 2, 0, 0, 0, utc)
	checkOut := time.Date(2026, 6, 16, 10, 0, 0, 0, utc)

	_, err := NewBooking(uuid.New(), checkIn, checkOut, utc, now)
	assert.ErrorIs(t, err, ErrBookingNotAvailable)
}

func TestNewBooking_MultiPeriod(t *testing.T) {
	now := time.Date(2026, 6, 15, 10, 0, 0, 0, utc)
	checkIn := time.Date(2026, 6, 15, 14, 0, 0, 0, utc)
	checkOut := time.Date(2026, 6, 16, 13, 0, 0, 0, utc)

	b, err := NewBooking(uuid.New(), checkIn, checkOut, utc, now)
	require.NoError(t, err)
	assert.Equal(t, time.Date(2026, 6, 17, 12, 0, 0, 0, utc), b.EndTime)
}

func TestNewBooking_CheckOutExactly1200(t *testing.T) {
	now := time.Date(2026, 6, 15, 10, 0, 0, 0, utc)
	checkIn := time.Date(2026, 6, 15, 14, 0, 0, 0, utc)
	checkOut := time.Date(2026, 6, 16, 12, 0, 0, 0, utc)

	b, err := NewBooking(uuid.New(), checkIn, checkOut, utc, now)
	require.NoError(t, err)
	assert.Equal(t, time.Date(2026, 6, 16, 12, 0, 0, 0, utc), b.EndTime)
}

func TestNewBooking_TooLong(t *testing.T) {
	now := time.Date(2026, 6, 15, 10, 0, 0, 0, utc)
	checkIn := time.Date(2026, 6, 15, 14, 0, 0, 0, utc)
	checkOut := time.Date(2028, 6, 15, 14, 0, 0, 0, utc)

	_, err := NewBooking(uuid.New(), checkIn, checkOut, utc, now)
	assert.ErrorIs(t, err, ErrBookingTooLong)
}

func TestConfirm_FromCreated(t *testing.T) {
	b := &Booking{
		Status:    StatusCreated,
		StartTime: time.Date(2026, 7, 1, 14, 0, 0, 0, utc),
	}
	now := time.Date(2026, 6, 15, 10, 0, 0, 0, utc)

	err := b.Confirm(now)
	require.NoError(t, err)
	assert.Equal(t, StatusConfirmed, b.Status)
}

func TestConfirm_FromConfirmed(t *testing.T) {
	b := &Booking{Status: StatusConfirmed}
	err := b.Confirm(time.Now())
	assert.ErrorIs(t, err, ErrBookingInvalidTransition)
}

func TestConfirm_FromCancelled(t *testing.T) {
	b := &Booking{Status: StatusCancelled}
	err := b.Confirm(time.Now())
	assert.ErrorIs(t, err, ErrBookingInvalidTransition)
}

func TestConfirm_FromExpired(t *testing.T) {
	b := &Booking{Status: StatusExpired}
	err := b.Confirm(time.Now())
	assert.ErrorIs(t, err, ErrBookingInvalidTransition)
}

func TestConfirm_AutoExpire(t *testing.T) {
	b := &Booking{
		Status:    StatusCreated,
		StartTime: time.Date(2026, 6, 14, 14, 0, 0, 0, utc),
	}
	now := time.Date(2026, 6, 15, 10, 0, 0, 0, utc)

	err := b.Confirm(now)
	assert.ErrorIs(t, err, ErrBookingExpired)
	assert.Equal(t, StatusExpired, b.Status)
}

func TestCancel_FromCreated(t *testing.T) {
	b := &Booking{
		Status:    StatusCreated,
		StartTime: time.Date(2026, 7, 1, 14, 0, 0, 0, utc),
	}
	now := time.Date(2026, 6, 15, 10, 0, 0, 0, utc)

	err := b.Cancel(now)
	require.NoError(t, err)
	assert.Equal(t, StatusCancelled, b.Status)
}

func TestCancel_FromConfirmed(t *testing.T) {
	b := &Booking{Status: StatusConfirmed}
	err := b.Cancel(time.Now())
	require.NoError(t, err)
	assert.Equal(t, StatusCancelled, b.Status)
}

func TestCancel_FromExpired(t *testing.T) {
	b := &Booking{Status: StatusExpired}
	err := b.Cancel(time.Now())
	assert.ErrorIs(t, err, ErrBookingInvalidTransition)
}

func TestCancel_FromCancelled(t *testing.T) {
	b := &Booking{Status: StatusCancelled}
	err := b.Cancel(time.Now())
	assert.ErrorIs(t, err, ErrBookingInvalidTransition)
}

func TestCancel_AutoExpire(t *testing.T) {
	b := &Booking{
		Status:    StatusCreated,
		StartTime: time.Date(2026, 6, 14, 14, 0, 0, 0, utc),
	}
	now := time.Date(2026, 6, 15, 10, 0, 0, 0, utc)

	err := b.Cancel(now)
	assert.ErrorIs(t, err, ErrBookingExpired)
	assert.Equal(t, StatusExpired, b.Status)
}

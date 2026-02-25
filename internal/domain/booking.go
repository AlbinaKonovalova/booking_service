package domain

import (
	"time"

	"github.com/google/uuid"
)

// Booking — агрегат "Бронирование" (Aggregate Root).
type Booking struct {
	ID         uuid.UUID
	ResourceID uuid.UUID
	StartTime  time.Time
	EndTime    time.Time
	CheckIn    time.Time
	CheckOut   time.Time
	Status     BookingStatus
	CreatedAt  time.Time
}

// NewBooking создаёт новое бронирование со всей доменной валидацией.
func NewBooking(resourceID uuid.UUID, checkIn, checkOut time.Time, hotelTZ *time.Location, now time.Time) (*Booking, error) {
	if !checkIn.Before(checkOut) {
		return nil, ErrBookingCheckInAfterCheckOut
	}

	if !checkIn.After(now) {
		return nil, ErrBookingInPast
	}

	startTime := checkIn

	D := CalculateD(startTime, hotelTZ)

	if err := ValidateCheckInWindow(startTime, D, hotelTZ); err != nil {
		return nil, err
	}

	endTime, err := CalculateEndTime(D, checkOut, hotelTZ)
	if err != nil {
		return nil, err
	}

	return &Booking{
		ID:         uuid.New(),
		ResourceID: resourceID,
		StartTime:  startTime.UTC(),
		EndTime:    endTime,
		CheckIn:    checkIn.UTC(),
		CheckOut:   checkOut.UTC(),
		Status:     StatusCreated,
		CreatedAt:  now.UTC(),
	}, nil
}

// Confirm переводит бронирование в статус CONFIRMED.
// Если бронирование просрочено (now > startTime и статус CREATED) — автоматически истекает.
func (b *Booking) Confirm(now time.Time) error {
	if b.Status != StatusCreated {
		return ErrBookingInvalidTransition
	}

	if now.After(b.StartTime) {
		b.Status = StatusExpired
		return ErrBookingExpired
	}

	b.Status = StatusConfirmed
	return nil
}

// Cancel переводит бронирование в статус CANCELLED.
func (b *Booking) Cancel(now time.Time) error {
	if b.Status == StatusCreated && now.After(b.StartTime) {
		b.Status = StatusExpired
		return ErrBookingExpired
	}

	if b.Status != StatusCreated && b.Status != StatusConfirmed {
		return ErrBookingInvalidTransition
	}

	b.Status = StatusCancelled
	return nil
}

// Expire переводит бронирование в статус EXPIRED.
func (b *Booking) Expire() error {
	if b.Status != StatusCreated {
		return ErrBookingInvalidTransition
	}
	b.Status = StatusExpired
	return nil
}

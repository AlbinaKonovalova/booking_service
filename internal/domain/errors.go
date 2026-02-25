package domain

import "errors"

// Resource errors
var (
	ErrResourceNameEmpty      = errors.New("resource name must not be empty")
	ErrResourceNotFound       = errors.New("resource not found")
	ErrResourceAlreadyRemoved = errors.New("resource already removed")
)

// Booking errors
var (
	ErrBookingNotFound             = errors.New("booking not found")
	ErrBookingInPast               = errors.New("booking cannot be created in the past")
	ErrBookingNotAvailable         = errors.New("check-in time is outside the allowed window")
	ErrBookingOverlap              = errors.New("booking overlaps with an existing active booking")
	ErrBookingCheckInAfterCheckOut = errors.New("check_in must be before check_out")
	ErrBookingTooLong              = errors.New("booking duration exceeds maximum of 365 periods")
	ErrBookingInvalidTransition    = errors.New("invalid booking status transition")
	ErrBookingExpired              = errors.New("booking has expired")
)

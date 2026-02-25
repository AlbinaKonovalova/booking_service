package input

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/AlbinaKonovalova/booking_service/internal/domain"
)

// BookingService — входящий порт для операций с бронированиями.
type BookingService interface {
	CreateBooking(ctx context.Context, resourceID uuid.UUID, checkIn, checkOut time.Time) (*domain.Booking, error)
}

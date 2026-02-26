package output

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/AlbinaKonovalova/booking_service/internal/domain"
)

// BookingRepository — исходящий порт для хранения бронирований.
type BookingRepository interface {
	Save(ctx context.Context, booking *domain.Booking) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Booking, error)
	UpdateStatus(ctx context.Context, booking *domain.Booking) error
	HasOverlap(ctx context.Context, resourceID uuid.UUID, startTime, endTime time.Time) (bool, error)
	HasActiveByResourceID(ctx context.Context, resourceID uuid.UUID) (bool, error)
	ListByResourceID(ctx context.Context, resourceID uuid.UUID, status *domain.BookingStatus) ([]*domain.Booking, error)
	ExpireOverdue(ctx context.Context, now time.Time) (int64, error)
	CompleteFinished(ctx context.Context, now time.Time) (int64, error)
}

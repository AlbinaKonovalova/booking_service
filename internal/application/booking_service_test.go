package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AlbinaKonovalova/booking_service/internal/domain"
)

func TestBookingService_CreateBooking_HappyPath(t *testing.T) {
	resourceID := uuid.New()
	checkIn := time.Date(2026, 7, 1, 14, 0, 0, 0, time.UTC)
	checkOut := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)

	svc := NewBookingService(
		&mockBookingRepo{},
		&mockResourceRepo{
			getByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.Resource, error) {
				return &domain.Resource{ID: id, Name: "Room"}, nil
			},
		},
		&mockTxManager{},
		time.UTC,
	)

	booking, err := svc.CreateBooking(context.Background(), resourceID, checkIn, checkOut)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusCreated, booking.Status)
	assert.Equal(t, resourceID, booking.ResourceID)
}

func TestBookingService_CreateBooking_ResourceNotFound(t *testing.T) {
	svc := NewBookingService(
		&mockBookingRepo{},
		&mockResourceRepo{
			getByIDFunc: func(_ context.Context, _ uuid.UUID) (*domain.Resource, error) {
				return nil, domain.ErrResourceNotFound
			},
		},
		&mockTxManager{},
		time.UTC,
	)

	_, err := svc.CreateBooking(context.Background(), uuid.New(),
		time.Date(2026, 7, 1, 14, 0, 0, 0, time.UTC),
		time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC))
	assert.ErrorIs(t, err, domain.ErrResourceNotFound)
}

func TestBookingService_CreateBooking_ResourceRemoved(t *testing.T) {
	svc := NewBookingService(
		&mockBookingRepo{},
		&mockResourceRepo{
			getByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.Resource, error) {
				r := &domain.Resource{ID: id, Name: "Room"}
				_ = r.Remove()
				return r, nil
			},
		},
		&mockTxManager{},
		time.UTC,
	)

	_, err := svc.CreateBooking(context.Background(), uuid.New(),
		time.Date(2026, 7, 1, 14, 0, 0, 0, time.UTC),
		time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC))
	assert.ErrorIs(t, err, domain.ErrResourceAlreadyRemoved)
}

func TestBookingService_CreateBooking_Overlap(t *testing.T) {
	svc := NewBookingService(
		&mockBookingRepo{
			hasOverlapFunc: func(_ context.Context, _ uuid.UUID, _, _ time.Time) (bool, error) {
				return true, nil
			},
		},
		&mockResourceRepo{},
		&mockTxManager{},
		time.UTC,
	)

	_, err := svc.CreateBooking(context.Background(), uuid.New(),
		time.Date(2026, 7, 1, 14, 0, 0, 0, time.UTC),
		time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC))
	assert.ErrorIs(t, err, domain.ErrBookingOverlap)
}

func TestBookingService_CreateBooking_InPast(t *testing.T) {
	svc := NewBookingService(
		&mockBookingRepo{},
		&mockResourceRepo{},
		&mockTxManager{},
		time.UTC,
	)

	_, err := svc.CreateBooking(context.Background(), uuid.New(),
		time.Date(2020, 1, 1, 14, 0, 0, 0, time.UTC),
		time.Date(2020, 1, 2, 10, 0, 0, 0, time.UTC))
	assert.ErrorIs(t, err, domain.ErrBookingInPast)
}

func TestBookingService_ConfirmBooking_HappyPath(t *testing.T) {
	bookingID := uuid.New()
	svc := NewBookingService(
		&mockBookingRepo{
			getByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.Booking, error) {
				return &domain.Booking{
					ID:        id,
					Status:    domain.StatusCreated,
					StartTime: time.Date(2026, 7, 1, 14, 0, 0, 0, time.UTC),
				}, nil
			},
		},
		&mockResourceRepo{},
		&mockTxManager{},
		time.UTC,
	)

	booking, err := svc.ConfirmBooking(context.Background(), bookingID)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusConfirmed, booking.Status)
}

func TestBookingService_ConfirmBooking_NotFound(t *testing.T) {
	svc := NewBookingService(
		&mockBookingRepo{},
		&mockResourceRepo{},
		&mockTxManager{},
		time.UTC,
	)

	_, err := svc.ConfirmBooking(context.Background(), uuid.New())
	assert.ErrorIs(t, err, domain.ErrBookingNotFound)
}

func TestBookingService_ConfirmBooking_AutoExpire(t *testing.T) {
	var statusSaved domain.BookingStatus
	svc := NewBookingService(
		&mockBookingRepo{
			getByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.Booking, error) {
				return &domain.Booking{
					ID:        id,
					Status:    domain.StatusCreated,
					StartTime: time.Date(2025, 1, 1, 14, 0, 0, 0, time.UTC),
				}, nil
			},
			updateStatusFunc: func(_ context.Context, b *domain.Booking) error {
				statusSaved = b.Status
				return nil
			},
		},
		&mockResourceRepo{},
		&mockTxManager{},
		time.UTC,
	)

	_, err := svc.ConfirmBooking(context.Background(), uuid.New())
	assert.ErrorIs(t, err, domain.ErrBookingExpired)
	assert.Equal(t, domain.StatusExpired, statusSaved)
}

func TestBookingService_CancelBooking_FromCreated(t *testing.T) {
	svc := NewBookingService(
		&mockBookingRepo{
			getByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.Booking, error) {
				return &domain.Booking{
					ID:        id,
					Status:    domain.StatusCreated,
					StartTime: time.Date(2026, 7, 1, 14, 0, 0, 0, time.UTC),
				}, nil
			},
		},
		&mockResourceRepo{},
		&mockTxManager{},
		time.UTC,
	)

	booking, err := svc.CancelBooking(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.Equal(t, domain.StatusCancelled, booking.Status)
}

func TestBookingService_CancelBooking_FromConfirmed(t *testing.T) {
	svc := NewBookingService(
		&mockBookingRepo{
			getByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.Booking, error) {
				return &domain.Booking{
					ID:     id,
					Status: domain.StatusConfirmed,
				}, nil
			},
		},
		&mockResourceRepo{},
		&mockTxManager{},
		time.UTC,
	)

	booking, err := svc.CancelBooking(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.Equal(t, domain.StatusCancelled, booking.Status)
}

func TestBookingService_CancelBooking_FromExpired(t *testing.T) {
	svc := NewBookingService(
		&mockBookingRepo{
			getByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.Booking, error) {
				return &domain.Booking{
					ID:     id,
					Status: domain.StatusExpired,
				}, nil
			},
		},
		&mockResourceRepo{},
		&mockTxManager{},
		time.UTC,
	)

	_, err := svc.CancelBooking(context.Background(), uuid.New())
	assert.ErrorIs(t, err, domain.ErrBookingInvalidTransition)
}

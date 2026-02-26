package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/AlbinaKonovalova/booking_service/internal/domain"
)

type mockTxManager struct{}

func (m *mockTxManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type mockResourceRepo struct {
	saveFunc    func(ctx context.Context, resource *domain.Resource) error
	getByIDFunc func(ctx context.Context, id uuid.UUID) (*domain.Resource, error)
	updateFunc  func(ctx context.Context, resource *domain.Resource) error
}

func (m *mockResourceRepo) Save(ctx context.Context, resource *domain.Resource) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, resource)
	}
	return nil
}

func (m *mockResourceRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Resource, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return &domain.Resource{ID: id, Name: "Test", CreatedAt: time.Now()}, nil
}

func (m *mockResourceRepo) Update(ctx context.Context, resource *domain.Resource) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, resource)
	}
	return nil
}

type mockBookingRepo struct {
	saveFunc             func(ctx context.Context, booking *domain.Booking) error
	getByIDFunc          func(ctx context.Context, id uuid.UUID) (*domain.Booking, error)
	updateStatusFunc     func(ctx context.Context, booking *domain.Booking) error
	hasOverlapFunc       func(ctx context.Context, resourceID uuid.UUID, startTime, endTime time.Time) (bool, error)
	hasActiveFunc        func(ctx context.Context, resourceID uuid.UUID) (bool, error)
	listByResourceIDFunc func(ctx context.Context, resourceID uuid.UUID, status *domain.BookingStatus) ([]*domain.Booking, error)
}

func (m *mockBookingRepo) Save(ctx context.Context, booking *domain.Booking) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, booking)
	}
	return nil
}

func (m *mockBookingRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Booking, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, domain.ErrBookingNotFound
}

func (m *mockBookingRepo) UpdateStatus(ctx context.Context, booking *domain.Booking) error {
	if m.updateStatusFunc != nil {
		return m.updateStatusFunc(ctx, booking)
	}
	return nil
}

func (m *mockBookingRepo) HasOverlap(ctx context.Context, resourceID uuid.UUID, startTime, endTime time.Time) (bool, error) {
	if m.hasOverlapFunc != nil {
		return m.hasOverlapFunc(ctx, resourceID, startTime, endTime)
	}
	return false, nil
}

func (m *mockBookingRepo) HasActiveByResourceID(ctx context.Context, resourceID uuid.UUID) (bool, error) {
	if m.hasActiveFunc != nil {
		return m.hasActiveFunc(ctx, resourceID)
	}
	return false, nil
}

func (m *mockBookingRepo) ListByResourceID(ctx context.Context, resourceID uuid.UUID, status *domain.BookingStatus) ([]*domain.Booking, error) {
	if m.listByResourceIDFunc != nil {
		return m.listByResourceIDFunc(ctx, resourceID, status)
	}
	return nil, nil
}

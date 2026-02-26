package application

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AlbinaKonovalova/booking_service/internal/domain"
)

func TestResourceService_CreateResource_HappyPath(t *testing.T) {
	svc := NewResourceService(&mockResourceRepo{}, &mockBookingRepo{}, &mockTxManager{})

	res, err := svc.CreateResource(context.Background(), "Room A")
	require.NoError(t, err)
	assert.Equal(t, "Room A", res.Name)
	assert.NotEmpty(t, res.ID)
}

func TestResourceService_CreateResource_EmptyName(t *testing.T) {
	svc := NewResourceService(&mockResourceRepo{}, &mockBookingRepo{}, &mockTxManager{})

	_, err := svc.CreateResource(context.Background(), "")
	assert.ErrorIs(t, err, domain.ErrResourceNameEmpty)
}

func TestResourceService_DeleteResource_HappyPath(t *testing.T) {
	resourceID := uuid.New()
	svc := NewResourceService(
		&mockResourceRepo{
			getByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.Resource, error) {
				return &domain.Resource{ID: id, Name: "Room"}, nil
			},
		},
		&mockBookingRepo{
			hasActiveFunc: func(_ context.Context, _ uuid.UUID) (bool, error) {
				return false, nil
			},
		},
		&mockTxManager{},
	)

	err := svc.DeleteResource(context.Background(), resourceID)
	require.NoError(t, err)
}

func TestResourceService_DeleteResource_NotFound(t *testing.T) {
	svc := NewResourceService(
		&mockResourceRepo{
			getByIDFunc: func(_ context.Context, _ uuid.UUID) (*domain.Resource, error) {
				return nil, domain.ErrResourceNotFound
			},
		},
		&mockBookingRepo{},
		&mockTxManager{},
	)

	err := svc.DeleteResource(context.Background(), uuid.New())
	assert.ErrorIs(t, err, domain.ErrResourceNotFound)
}

func TestResourceService_DeleteResource_HasActiveBookings(t *testing.T) {
	svc := NewResourceService(
		&mockResourceRepo{},
		&mockBookingRepo{
			hasActiveFunc: func(_ context.Context, _ uuid.UUID) (bool, error) {
				return true, nil
			},
		},
		&mockTxManager{},
	)

	err := svc.DeleteResource(context.Background(), uuid.New())
	assert.ErrorIs(t, err, domain.ErrResourceHasActiveBookings)
}

func TestResourceService_DeleteResource_AlreadyRemoved(t *testing.T) {
	svc := NewResourceService(
		&mockResourceRepo{
			getByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.Resource, error) {
				removed := domain.Resource{ID: id, Name: "Room"}
				_ = removed.Remove()
				return &removed, nil
			},
		},
		&mockBookingRepo{},
		&mockTxManager{},
	)

	err := svc.DeleteResource(context.Background(), uuid.New())
	assert.ErrorIs(t, err, domain.ErrResourceAlreadyRemoved)
}

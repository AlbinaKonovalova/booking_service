package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/AlbinaKonovalova/booking_service/internal/domain"
	"github.com/AlbinaKonovalova/booking_service/internal/ports/output"
)

// ResourceService реализует бизнес-логику работы с ресурсами.
type ResourceService struct {
	resourceRepo output.ResourceRepository
	bookingRepo  output.BookingRepository
	txManager    output.TxManager
}

// NewResourceService создаёт новый ResourceService.
func NewResourceService(
	resourceRepo output.ResourceRepository,
	bookingRepo output.BookingRepository,
	txManager output.TxManager,
) *ResourceService {
	return &ResourceService{
		resourceRepo: resourceRepo,
		bookingRepo:  bookingRepo,
		txManager:    txManager,
	}
}

// CreateResource создаёт новый ресурс.
func (s *ResourceService) CreateResource(ctx context.Context, name string) (*domain.Resource, error) {
	resource, err := domain.NewResource(name)
	if err != nil {
		return nil, err
	}

	if err := s.resourceRepo.Save(ctx, resource); err != nil {
		return nil, fmt.Errorf("saving resource: %w", err)
	}

	return resource, nil
}

// DeleteResource выполняет soft delete ресурса.
// Запрещено если есть активные бронирования (CREATED/CONFIRMED).
func (s *ResourceService) DeleteResource(ctx context.Context, id uuid.UUID) error {
	return s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		resource, err := s.resourceRepo.GetByID(txCtx, id)
		if err != nil {
			return err
		}

		hasActive, err := s.bookingRepo.HasActiveByResourceID(txCtx, id)
		if err != nil {
			return fmt.Errorf("checking active bookings: %w", err)
		}
		if hasActive {
			return domain.ErrResourceHasActiveBookings
		}

		if err := resource.Remove(); err != nil {
			return err
		}

		if err := s.resourceRepo.Update(txCtx, resource); err != nil {
			return fmt.Errorf("updating resource: %w", err)
		}

		return nil
	})
}

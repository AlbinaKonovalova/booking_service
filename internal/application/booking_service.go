package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/AlbinaKonovalova/booking_service/internal/domain"
	"github.com/AlbinaKonovalova/booking_service/internal/ports/output"
)

// BookingService реализует бизнес-логику работы с бронированиями.
type BookingService struct {
	bookingRepo  output.BookingRepository
	resourceRepo output.ResourceRepository
	txManager    output.TxManager
	hotelTZ      *time.Location
}

// NewBookingService создаёт новый BookingService.
func NewBookingService(
	bookingRepo output.BookingRepository,
	resourceRepo output.ResourceRepository,
	txManager output.TxManager,
	hotelTZ *time.Location,
) *BookingService {
	return &BookingService{
		bookingRepo:  bookingRepo,
		resourceRepo: resourceRepo,
		txManager:    txManager,
		hotelTZ:      hotelTZ,
	}
}

// CreateBooking создаёт новое бронирование.
func (s *BookingService) CreateBooking(ctx context.Context, resourceID uuid.UUID, checkIn, checkOut time.Time) (*domain.Booking, error) {
	var booking *domain.Booking

	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		// 1. Проверяем что ресурс существует и не удалён (с блокировкой FOR UPDATE)
		resource, err := s.resourceRepo.GetByID(txCtx, resourceID)
		if err != nil {
			return err
		}
		if resource.IsRemoved() {
			return domain.ErrResourceAlreadyRemoved
		}

		// 2. Создаём доменный объект (вся валидация внутри)
		now := time.Now()
		booking, err = domain.NewBooking(resourceID, checkIn, checkOut, s.hotelTZ, now)
		if err != nil {
			return err
		}

		// 3. Проверяем пересечения с активными бронями
		hasOverlap, err := s.bookingRepo.HasOverlap(txCtx, resourceID, booking.StartTime, booking.EndTime)
		if err != nil {
			return fmt.Errorf("checking overlap: %w", err)
		}
		if hasOverlap {
			return domain.ErrBookingOverlap
		}

		// 4. Сохраняем
		if err := s.bookingRepo.Save(txCtx, booking); err != nil {
			return fmt.Errorf("saving booking: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return booking, nil
}

// ConfirmBooking подтверждает бронирование (CREATED → CONFIRMED).
// Если бронирование просрочено — автоматически переводит в EXPIRED.
func (s *BookingService) ConfirmBooking(ctx context.Context, id uuid.UUID) (*domain.Booking, error) {
	var booking *domain.Booking

	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		var err error
		booking, err = s.bookingRepo.GetByID(txCtx, id)
		if err != nil {
			return err
		}

		now := time.Now()
		if confirmErr := booking.Confirm(now); confirmErr != nil {
			// Если бронь автоматически истекла — сохраняем EXPIRED в БД
			if booking.Status == domain.StatusExpired {
				if saveErr := s.bookingRepo.UpdateStatus(txCtx, booking); saveErr != nil {
					return fmt.Errorf("saving expired status: %w", saveErr)
				}
			}
			return confirmErr
		}

		if err := s.bookingRepo.UpdateStatus(txCtx, booking); err != nil {
			return fmt.Errorf("saving confirmed status: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return booking, nil
}

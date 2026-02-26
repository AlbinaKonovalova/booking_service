package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/AlbinaKonovalova/booking_service/internal/ports/output"
)

type ExpirationService struct {
	bookingRepo output.BookingRepository
	logger      *slog.Logger
}

func NewExpirationService(bookingRepo output.BookingRepository, logger *slog.Logger) *ExpirationService {
	return &ExpirationService{
		bookingRepo: bookingRepo,
		logger:      logger,
	}
}

func (s *ExpirationService) RunExpire(ctx context.Context) {
	now := time.Now()

	expired, err := s.bookingRepo.ExpireOverdue(ctx, now)
	if err != nil {
		s.logger.Error("failed to expire overdue bookings", slog.Any("error", err))
	} else if expired > 0 {
		s.logger.Info("expired overdue bookings", slog.Int64("count", expired))
	}
}

func (s *ExpirationService) RunComplete(ctx context.Context) {
	now := time.Now()

	completed, err := s.bookingRepo.CompleteFinished(ctx, now)
	if err != nil {
		s.logger.Error("failed to complete finished bookings", slog.Any("error", err))
	} else if completed > 0 {
		s.logger.Info("completed finished bookings", slog.Int64("count", completed))
	}
}

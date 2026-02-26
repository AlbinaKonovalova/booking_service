package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/AlbinaKonovalova/booking_service/internal/application"
)

type Scheduler struct {
	expirationService *application.ExpirationService
	expireInterval    time.Duration
	completionHour    int
	completionMinute  int
	logger            *slog.Logger
}

func NewScheduler(
	expirationService *application.ExpirationService,
	expireInterval time.Duration,
	completionTime string,
	logger *slog.Logger,
) (*Scheduler, error) {
	hour, minute, err := parseTime(completionTime)
	if err != nil {
		return nil, fmt.Errorf("invalid completion_time: %w", err)
	}
	return &Scheduler{
		expirationService: expirationService,
		expireInterval:    expireInterval,
		completionHour:    hour,
		completionMinute:  minute,
		logger:            logger,
	}, nil
}

func (s *Scheduler) Start(ctx context.Context) {
	s.logger.Info("scheduler started",
		slog.String("expire_interval", s.expireInterval.String()),
		slog.String("completion_time", fmt.Sprintf("%02d:%02d UTC daily", s.completionHour, s.completionMinute)),
	)

	go s.runExpireLoop(ctx)
	s.runCompletionLoop(ctx)
}

func (s *Scheduler) runExpireLoop(ctx context.Context) {
	ticker := time.NewTicker(s.expireInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.expirationService.RunExpire(ctx)
		}
	}
}

func (s *Scheduler) runCompletionLoop(ctx context.Context) {
	for {
		delay := timeUntilNext(s.completionHour, s.completionMinute)
		s.logger.Info("next completion run scheduled", slog.String("in", delay.String()))

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			s.logger.Info("scheduler stopped")
			return
		case <-timer.C:
			s.expirationService.RunComplete(ctx)
		}
	}
}

func timeUntilNext(hour, minute int) time.Duration {
	now := time.Now().UTC()
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, time.UTC)
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next.Sub(now)
}

func parseTime(s string) (int, int, error) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return 0, 0, err
	}
	return t.Hour(), t.Minute(), nil
}

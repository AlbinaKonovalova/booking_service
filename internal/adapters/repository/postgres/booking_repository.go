package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/AlbinaKonovalova/booking_service/internal/domain"
)

// BookingRepository реализует хранение бронирований в PostgreSQL.
type BookingRepository struct {
	db *sql.DB
}

// NewBookingRepository создаёт новый BookingRepository.
func NewBookingRepository(db *sql.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

// Save сохраняет новое бронирование в БД.
func (r *BookingRepository) Save(ctx context.Context, booking *domain.Booking) error {
	exec := getExecutor(ctx, r.db)
	query := `INSERT INTO bookings (id, resource_id, start_time, end_time, check_in, check_out, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := exec.ExecContext(ctx, query,
		booking.ID,
		booking.ResourceID,
		booking.StartTime,
		booking.EndTime,
		booking.CheckIn,
		booking.CheckOut,
		string(booking.Status),
		booking.CreatedAt,
	)
	return err
}

// GetByID загружает бронирование по ID. Внутри транзакции использует FOR UPDATE.
func (r *BookingRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Booking, error) {
	exec := getExecutor(ctx, r.db)

	query := `SELECT id, resource_id, start_time, end_time, check_in, check_out, status, created_at
		FROM bookings WHERE id = $1`
	if _, ok := ctx.Value(ctxKeyTx{}).(*sql.Tx); ok {
		query += ` FOR UPDATE`
	}

	var b domain.Booking
	var status string
	err := exec.QueryRowContext(ctx, query, id).Scan(
		&b.ID, &b.ResourceID, &b.StartTime, &b.EndTime,
		&b.CheckIn, &b.CheckOut, &status, &b.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrBookingNotFound
		}
		return nil, err
	}
	b.Status = domain.BookingStatus(status)
	return &b, nil
}

// UpdateStatus обновляет статус бронирования.
func (r *BookingRepository) UpdateStatus(ctx context.Context, booking *domain.Booking) error {
	exec := getExecutor(ctx, r.db)
	query := `UPDATE bookings SET status = $2 WHERE id = $1`
	_, err := exec.ExecContext(ctx, query, booking.ID, string(booking.Status))
	return err
}

// HasOverlap проверяет наличие пересечений с активными бронированиями ресурса.
// Интервалы пересекаются если: existing.start < new.end AND existing.end > new.start
func (r *BookingRepository) HasOverlap(ctx context.Context, resourceID uuid.UUID, startTime, endTime time.Time) (bool, error) {
	exec := getExecutor(ctx, r.db)
	query := `SELECT EXISTS(
		SELECT 1 FROM bookings
		WHERE resource_id = $1
			AND status IN ('CREATED', 'CONFIRMED')
			AND start_time < $3
			AND end_time > $2
	)`
	var exists bool
	err := exec.QueryRowContext(ctx, query, resourceID, startTime, endTime).Scan(&exists)
	return exists, err
}

// HasActiveByResourceID проверяет наличие активных бронирований у ресурса.
func (r *BookingRepository) HasActiveByResourceID(ctx context.Context, resourceID uuid.UUID) (bool, error) {
	exec := getExecutor(ctx, r.db)
	query := `SELECT EXISTS(
		SELECT 1 FROM bookings
		WHERE resource_id = $1
			AND status IN ('CREATED', 'CONFIRMED')
	)`
	var exists bool
	err := exec.QueryRowContext(ctx, query, resourceID).Scan(&exists)
	return exists, err
}

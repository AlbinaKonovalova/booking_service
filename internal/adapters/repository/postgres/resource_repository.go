package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/AlbinaKonovalova/booking_service/internal/domain"
	"github.com/google/uuid"
)

// ResourceRepository реализует хранение ресурсов в PostgreSQL.
type ResourceRepository struct {
	db *sql.DB
}

// NewResourceRepository создаёт новый ResourceRepository.
func NewResourceRepository(db *sql.DB) *ResourceRepository {
	return &ResourceRepository{db: db}
}

// Save сохраняет новый ресурс в БД.
func (r *ResourceRepository) Save(ctx context.Context, resource *domain.Resource) error {
	exec := getExecutor(ctx, r.db)
	query := `INSERT INTO resources (id, name, created_at) VALUES ($1, $2, $3)`
	_, err := exec.ExecContext(ctx, query, resource.ID, resource.Name, resource.CreatedAt)
	return err
}

// GetByID загружает ресурс по ID. Внутри транзакции использует FOR UPDATE для блокировки.
func (r *ResourceRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Resource, error) {
	exec := getExecutor(ctx, r.db)

	query := `SELECT id, name, created_at, removed_at FROM resources WHERE id = $1`
	if _, ok := ctx.Value(ctxKeyTx{}).(*sql.Tx); ok {
		query += ` FOR UPDATE`
	}

	var res domain.Resource
	var removedAt sql.NullTime
	err := exec.QueryRowContext(ctx, query, id).Scan(&res.ID, &res.Name, &res.CreatedAt, &removedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrResourceNotFound
		}
		return nil, err
	}

	if removedAt.Valid {
		t := removedAt.Time
		res.RemovedAt = &t
	}

	return &res, nil
}

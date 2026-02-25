package postgres

import (
	"context"
	"database/sql"

	"github.com/AlbinaKonovalova/booking_service/internal/domain"
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
	query := `INSERT INTO resources (id, name, created_at) VALUES ($1, $2, $3)`
	_, err := r.db.ExecContext(ctx, query, resource.ID, resource.Name, resource.CreatedAt)
	return err
}

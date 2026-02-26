package output

import (
	"context"

	"github.com/AlbinaKonovalova/booking_service/internal/domain"
	"github.com/google/uuid"
)

// ResourceRepository — исходящий порт для хранения ресурсов.
type ResourceRepository interface {
	Save(ctx context.Context, resource *domain.Resource) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Resource, error)
	Update(ctx context.Context, resource *domain.Resource) error
}

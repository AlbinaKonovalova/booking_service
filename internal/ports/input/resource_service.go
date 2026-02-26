package input

import (
	"context"

	"github.com/google/uuid"

	"github.com/AlbinaKonovalova/booking_service/internal/domain"
)

// ResourceService — входящий порт для операций с ресурсами.
type ResourceService interface {
	CreateResource(ctx context.Context, name string) (*domain.Resource, error)
	DeleteResource(ctx context.Context, id uuid.UUID) error
}

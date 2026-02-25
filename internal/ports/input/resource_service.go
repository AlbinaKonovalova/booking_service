package input

import (
	"context"

	"github.com/AlbinaKonovalova/booking_service/internal/domain"
)

// ResourceService — входящий порт для операций с ресурсами.
type ResourceService interface {
	CreateResource(ctx context.Context, name string) (*domain.Resource, error)
}

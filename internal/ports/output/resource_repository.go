package output

import (
	"context"

	"github.com/AlbinaKonovalova/booking_service/internal/domain"
)

// ResourceRepository — исходящий порт для хранения ресурсов.
type ResourceRepository interface {
	Save(ctx context.Context, resource *domain.Resource) error
}

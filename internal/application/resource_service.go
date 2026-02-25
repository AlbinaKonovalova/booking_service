package application

import (
	"context"
	"fmt"

	"github.com/AlbinaKonovalova/booking_service/internal/domain"
	"github.com/AlbinaKonovalova/booking_service/internal/ports/output"
)

// ResourceService реализует бизнес-логику работы с ресурсами.
type ResourceService struct {
	resourceRepo output.ResourceRepository
}

// NewResourceService создаёт новый ResourceService.
func NewResourceService(resourceRepo output.ResourceRepository) *ResourceService {
	return &ResourceService{
		resourceRepo: resourceRepo,
	}
}

// CreateResource создаёт новый ресурс.
func (s *ResourceService) CreateResource(ctx context.Context, name string) (*domain.Resource, error) {
	resource, err := domain.NewResource(name)
	if err != nil {
		return nil, err
	}

	if err := s.resourceRepo.Save(ctx, resource); err != nil {
		return nil, fmt.Errorf("saving resource: %w", err)
	}

	return resource, nil
}

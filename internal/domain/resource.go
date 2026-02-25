package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Resource — агрегат "Ресурс" (номер в отеле).
type Resource struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
	RemovedAt *time.Time
}

// NewResource создаёт новый ресурс с валидацией.
func NewResource(name string) (*Resource, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrResourceNameEmpty
	}

	now := time.Now().UTC()
	return &Resource{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: now,
	}, nil
}

// IsRemoved возвращает true, если ресурс удалён (soft delete).
func (r *Resource) IsRemoved() bool {
	return r.RemovedAt != nil
}

// Remove помечает ресурс как удалённый.
func (r *Resource) Remove() error {
	if r.IsRemoved() {
		return ErrResourceAlreadyRemoved
	}
	now := time.Now().UTC()
	r.RemovedAt = &now
	return nil
}

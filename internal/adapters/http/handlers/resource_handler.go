package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/AlbinaKonovalova/booking_service/internal/ports/input"
	"github.com/google/uuid"
)

// ResourceHandler — HTTP хендлер для операций с ресурсами.
type ResourceHandler struct {
	service input.ResourceService
}

// NewResourceHandler создаёт новый ResourceHandler.
func NewResourceHandler(service input.ResourceService) *ResourceHandler {
	return &ResourceHandler{service: service}
}

// createResourceRequest — тело запроса на создание ресурса.
type createResourceRequest struct {
	Name string `json:"name"`
}

// resourceResponse — тело ответа с ресурсом.
type resourceResponse struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	RemovedAt *time.Time `json:"removed_at,omitempty"`
}

// Create обрабатывает POST /resource.
func (h *ResourceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createResourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
			"code":  "BAD_REQUEST",
		})
		return
	}

	resource, err := h.service.CreateResource(r.Context(), req.Name)
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondJSON(w, http.StatusCreated, resourceResponse{
		ID:        resource.ID,
		Name:      resource.Name,
		CreatedAt: resource.CreatedAt,
		RemovedAt: resource.RemovedAt,
	})
}

// Delete обрабатывает DELETE /resource/{id}.
func (h *ResourceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid resource id format",
			Code:  "BAD_REQUEST",
		})
		return
	}

	if err := h.service.DeleteResource(r.Context(), id); err != nil {
		RespondError(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, map[string]string{
		"id":     id.String(),
		"status": "removed",
	})
}

package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/AlbinaKonovalova/booking_service/internal/domain"
)

// ErrorResponse — формат ошибки в API.
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

// RespondJSON отправляет JSON-ответ с указанным статусом.
func RespondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// RespondError маппит доменную ошибку в HTTP-ответ.
func RespondError(w http.ResponseWriter, err error) {
	status, code := mapDomainError(err)
	RespondJSON(w, status, ErrorResponse{
		Error: err.Error(),
		Code:  code,
	})
}

func mapDomainError(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrResourceNameEmpty):
		return http.StatusUnprocessableEntity, "INVALID_INPUT"
	case errors.Is(err, domain.ErrResourceNotFound):
		return http.StatusNotFound, "NOT_FOUND"
	case errors.Is(err, domain.ErrResourceAlreadyRemoved):
		return http.StatusGone, "ALREADY_REMOVED"
	default:
		return http.StatusInternalServerError, "INTERNAL_ERROR"
	}
}

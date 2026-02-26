package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/AlbinaKonovalova/booking_service/internal/domain"
	"github.com/AlbinaKonovalova/booking_service/internal/ports/input"
)

// BookingHandler — HTTP хендлер для операций с бронированиями.
type BookingHandler struct {
	service input.BookingService
}

// NewBookingHandler создаёт новый BookingHandler.
func NewBookingHandler(service input.BookingService) *BookingHandler {
	return &BookingHandler{service: service}
}

// createBookingRequest — тело запроса на создание бронирования.
type createBookingRequest struct {
	ResourceID string `json:"resource_id"`
	CheckIn    string `json:"check_in"`
	CheckOut   string `json:"check_out"`
}

// bookingResponse — полное тело ответа с бронированием.
type bookingResponse struct {
	ID         uuid.UUID `json:"id"`
	ResourceID uuid.UUID `json:"resource_id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	CheckIn    time.Time `json:"check_in"`
	CheckOut   time.Time `json:"check_out"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

// bookingStatusResponse — компактный ответ для операций смены статуса.
type bookingStatusResponse struct {
	ID     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

// Create обрабатывает POST /booking.
func (h *BookingHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid request body",
			Code:  "BAD_REQUEST",
		})
		return
	}

	resourceID, err := uuid.Parse(req.ResourceID)
	if err != nil {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid resource_id format",
			Code:  "BAD_REQUEST",
		})
		return
	}

	checkIn, err := time.Parse(time.RFC3339, req.CheckIn)
	if err != nil {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid check_in format, expected RFC3339",
			Code:  "BAD_REQUEST",
		})
		return
	}

	checkOut, err := time.Parse(time.RFC3339, req.CheckOut)
	if err != nil {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid check_out format, expected RFC3339",
			Code:  "BAD_REQUEST",
		})
		return
	}

	booking, err := h.service.CreateBooking(r.Context(), resourceID, checkIn, checkOut)
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondJSON(w, http.StatusCreated, toBookingResponse(booking))
}

// Confirm обрабатывает POST /booking/{id}/confirm.
func (h *BookingHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid booking id format",
			Code:  "BAD_REQUEST",
		})
		return
	}

	booking, err := h.service.ConfirmBooking(r.Context(), id)
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, bookingStatusResponse{
		ID:     booking.ID,
		Status: string(booking.Status),
	})
}

// Cancel обрабатывает POST /booking/{id}/cancel.
func (h *BookingHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid booking id format",
			Code:  "BAD_REQUEST",
		})
		return
	}

	booking, err := h.service.CancelBooking(r.Context(), id)
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, bookingStatusResponse{
		ID:     booking.ID,
		Status: string(booking.Status),
	})
}

func toBookingResponse(b *domain.Booking) bookingResponse {
	return bookingResponse{
		ID:         b.ID,
		ResourceID: b.ResourceID,
		StartTime:  b.StartTime,
		EndTime:    b.EndTime,
		CheckIn:    b.CheckIn,
		CheckOut:   b.CheckOut,
		Status:     string(b.Status),
		CreatedAt:  b.CreatedAt,
	}
}

// ListByResource обрабатывает GET /resource/{id}/bookings.
func (h *BookingHandler) ListByResource(w http.ResponseWriter, r *http.Request) {
	resourceID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid resource id format",
			Code:  "BAD_REQUEST",
		})
		return
	}

	var status *domain.BookingStatus
	if s := r.URL.Query().Get("status"); s != "" {
		bs := domain.BookingStatus(s)
		status = &bs
	}

	bookings, err := h.service.ListBookingsByResource(r.Context(), resourceID, status)
	if err != nil {
		RespondError(w, err)
		return
	}

	resp := make([]bookingResponse, 0, len(bookings))
	for _, b := range bookings {
		resp = append(resp, toBookingResponse(b))
	}

	RespondJSON(w, http.StatusOK, resp)
}

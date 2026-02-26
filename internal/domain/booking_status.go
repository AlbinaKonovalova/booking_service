package domain

// BookingStatus — статус бронирования (value object).
type BookingStatus string

const (
	StatusCreated   BookingStatus = "CREATED"
	StatusConfirmed BookingStatus = "CONFIRMED"
	StatusCancelled BookingStatus = "CANCELLED"
	StatusExpired   BookingStatus = "EXPIRED"
	StatusCompleted BookingStatus = "COMPLETED"
)

// IsActive возвращает true, если бронирование блокирует ресурс.
func (s BookingStatus) IsActive() bool {
	return s == StatusCreated || s == StatusConfirmed
}

// String реализует fmt.Stringer.
func (s BookingStatus) String() string {
	return string(s)
}

package domain

import "time"

// CalculateD вычисляет день периода D из start_time в таймзоне отеля.
//
// Формула:
//   - если time(t) >= 12:00 → D = date(t)
//   - иначе → D = date(t) - 1 day
func CalculateD(startTime time.Time, hotelTZ *time.Location) time.Time {
	t := startTime.In(hotelTZ)
	y, m, d := t.Date()

	if t.Hour() >= 12 {
		return time.Date(y, m, d, 0, 0, 0, 0, hotelTZ)
	}
	return time.Date(y, m, d-1, 0, 0, 0, 0, hotelTZ)
}

// ValidateCheckInWindow проверяет, что start_time попадает в допустимое окно заезда:
// D 12:00 <= start_time < (D+1) 02:00.
func ValidateCheckInWindow(startTime time.Time, D time.Time, hotelTZ *time.Location) error {
	windowStart := time.Date(D.Year(), D.Month(), D.Day(), 12, 0, 0, 0, hotelTZ)
	windowEnd := time.Date(D.Year(), D.Month(), D.Day()+1, 2, 0, 0, 0, hotelTZ)

	t := startTime.In(hotelTZ)
	if t.Before(windowStart) || !t.Before(windowEnd) {
		return ErrBookingNotAvailable
	}
	return nil
}

// CalculateEndTime вычисляет end_time = (D+N) 12:00, где N — минимальное значение
// такое что check_out <= end_time, при 1 <= N <= 365.
func CalculateEndTime(D time.Time, checkOut time.Time, hotelTZ *time.Location) (time.Time, error) {
	for n := 1; n <= 365; n++ {
		endTime := time.Date(D.Year(), D.Month(), D.Day()+n, 12, 0, 0, 0, hotelTZ)
		if !checkOut.After(endTime) { // checkOut <= endTime
			return endTime.UTC(), nil
		}
	}
	return time.Time{}, ErrBookingTooLong
}

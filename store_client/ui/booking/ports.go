package booking

import "apteka_booking/models"

type Service interface {
	Ping() error
	GetBookings() ([]models.Booking, error)
	SetBookingStatus(id int, status string) error
	DeleteBooking(id int) error
}

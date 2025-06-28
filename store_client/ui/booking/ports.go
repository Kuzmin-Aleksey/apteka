package booking

import "store_client/models"

type Service interface {
	Ping() error
	GetBookings() ([]models.Booking, error)
	GetBookingsChan() <-chan models.BookingsWithError
	SetBookingStatus(id int, status string) error
	DeleteBooking(id int) error
}

package booking

import (
	"server/domain/models"
	"time"
)

type CreateBookDTO struct {
	StoreId  int                  `json:"store_id"`
	Phone    string               `json:"phone"`
	Message  string               `json:"message,omitempty"`
	Products []models.BookProduct `json:"products"`
}

type GetBookingResponseDTO struct {
	models.Book
	Delay time.Duration `json:"delay"`
}

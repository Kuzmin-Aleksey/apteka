package booking

import (
	"server/internal/domain/entity"
	"time"
)

type CreateBookDTO struct {
	StoreId  int                  `json:"store_id"`
	Phone    string               `json:"phone"`
	Message  string               `json:"message,omitempty"`
	Products []entity.BookProduct `json:"products"`
}

type GetBookingResponseDTO struct {
	entity.Book
	Delay time.Duration `json:"delay"`
}

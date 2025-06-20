package booking

import (
	"server/internal/domain/aggregate"
	"server/internal/domain/entity"
)

type CreateBookDTO struct {
	Username string               `json:"username"`
	Phone    string               `json:"phone"`
	Message  string               `json:"message,omitempty"`
	Products []entity.BookProduct `json:"products"`
}

type GetBookingResponseDTO struct {
	aggregate.BookWithProducts
	Delay int `json:"delay"`
}

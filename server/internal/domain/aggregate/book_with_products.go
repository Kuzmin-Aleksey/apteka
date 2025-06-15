package aggregate

import (
	"server/internal/domain/entity"
)

type BookWithProducts struct {
	entity.Book
	Products []entity.BookProduct `json:"products"`
}

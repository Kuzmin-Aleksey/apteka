package promotion

import "server/internal/domain/entity"

type PromotionInStock struct {
	entity.Promotion
	InStock bool            `json:"in_stock"`
	Product *entity.Product `json:"product,omitempty"`
}

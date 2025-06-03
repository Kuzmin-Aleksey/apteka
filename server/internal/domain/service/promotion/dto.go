package promotion

import "server/internal/domain/entity"

type PromotionInStock struct {
	entity.Promotion
	PriceWithoutDiscount int  `json:"price_without_discount"`
	InStock              bool `json:"in_stock"`
}

package promotion

import "server/domain/models"

type PromotionInStock struct {
	models.Promotion
	PriceWithoutDiscount int  `json:"price_without_discount"`
	InStock              bool `json:"in_stock"`
}

package products

import (
	"server/internal/domain/entity"
)

func productsToMap(products []entity.Product) map[int]*entity.Product {
	productsMap := make(map[int]*entity.Product)

	for _, product := range products {
		productsMap[product.CodeSTU] = &product
	}

	return productsMap
}

func equalProducts(prod1, prod2 *entity.Product) bool {
	if prod1 == nil || prod2 == nil {
		return false
	}

	return prod1.CodeSTU == prod2.CodeSTU &&
		prod1.StoreId == prod2.StoreId &&
		prod1.Name == prod2.Name &&
		prod1.GTIN == prod2.GTIN &&
		prod1.Description == prod2.Description &&
		prod1.Count == prod2.Count &&
		prod1.Price == prod2.Price &&
		prod1.Country == prod2.Country &&
		prod1.Producer == prod2.Producer
}

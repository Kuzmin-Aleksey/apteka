package utils

import (
	"apteka_client/models"
)

func MergingDuplicates(products []models.Product) []models.Product {
	lastCode := -1
	lastI := 0
	lenProducts := len(products)

	for i := 0; i < lenProducts; {
		if products[i].CodeSTU == lastCode {
			products[lastI].Count += products[i].Count
			if products[i].Price > products[lastI].Price {
				products[lastI].Price = products[i].Price
			}
			products = append(products[:i], products[i+1:]...)
			lenProducts--
		} else {
			lastCode = products[i].CodeSTU
			lastI = i
			i++
		}
	}
	return products
}

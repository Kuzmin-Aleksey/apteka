package utils

import (
	"efarma_integration/models"
	"fmt"
	"testing"
)

var products = []models.Product{
	{CodeSTU: 1, Name: "Apple", GTIN: 1234567890123, Description: "Fresh red apple", Count: 20, Price: 100, Producer: "Farm Fresh"},
	{CodeSTU: 1, Name: "Banana", GTIN: 1234567890124, Description: "Ripe banana", Count: 20, Price: 60, Producer: "Tropical Fruits"},
	{CodeSTU: 3, Name: "Orange Juice", GTIN: 1234567890125, Description: "100% pure orange juice", Count: 40, Price: 150, Producer: "Sunny Days"},
	{CodeSTU: 3, Name: "Chocolate Bar", GTIN: 1234567890126, Description: "Delicious dark chocolate", Count: 100, Price: 200, Producer: "Choco Heaven"},
	{CodeSTU: 5, Name: "Milk", GTIN: 1234567890127, Description: "Fresh milk", Count: 25, Price: 80, Producer: "Dairy Best"},
	{CodeSTU: 6, Name: "Bread", GTIN: 1234567890128, Description: "Whole wheat bread", Count: 30, Price: 50, Producer: "Bakery Delight"},
	{CodeSTU: 9, Name: "Pasta", GTIN: 1234567890129, Description: "Italian pasta", Count: 20, Price: 120, Producer: "Pasta Makers"},
	{CodeSTU: 9, Name: "Rice", GTIN: 1234567890130, Description: "Basmati rice", Count: 10, Price: 250, Producer: "Grains & Seeds"},
	{CodeSTU: 9, Name: "Cheese", GTIN: 1234567890131, Description: "Aged cheddar cheese", Count: 15, Price: 300, Producer: "Cheese Factory"},
	{CodeSTU: 10, Name: "Coffee", GTIN: 1234567890132, Description: "Premium coffee beans", Count: 5, Price: 500, Producer: "Coffee Co."},
}

func TestMergingDuplicates(t *testing.T) {
	products = MergingDuplicates(products)
	for _, product := range products {
		fmt.Printf("%+v \n", product)
	}
}

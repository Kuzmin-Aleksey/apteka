package models

type Promotion struct {
	ProductCode int    `json:"product_code"`
	ProductName string `json:"product_name"`
	Discount    int    `json:"discount"`
}

package entity

type Promotion struct {
	ProductCode int    `json:"product_code"`
	ProductName string `json:"product_name"`
	Discount    int    `json:"discount"`
	IsPercent   bool   `json:"is_percent"`
}

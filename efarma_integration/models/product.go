package models

type Product struct {
	CodeSTU     int    `json:"id"`
	Name        string `json:"name"`
	GTIN        uint64 `json:"gtin"`
	Description string `json:"description"`
	Count       int    `json:"count"`
	Price       int    `json:"price"`
	Country     string `json:"country"`
	Producer    string `json:"producer"`
}

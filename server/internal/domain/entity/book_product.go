package entity

type BookProduct struct {
	CodeSTU  int    `json:"code_stu"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Price    int    `json:"price"`
}

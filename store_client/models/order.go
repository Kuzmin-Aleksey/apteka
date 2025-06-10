package models

import "time"

const (
	OrderStatusCreated   string = "create"
	OrderStatusConfirmed string = "confirm"
	OrderStatusRejected  string = "reject"
	OrderStatusDone      string = "done"
	OrderStatusReceive   string = "receive"
)

type OrderProduct struct {
	CodeSTU  int    `json:"code_stu"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

type Order struct {
	Id        int           `json:"id"`
	CreatedAt time.Time     `json:"created_at"`
	Status    string        `json:"status"`
	Username  string        `json:"username"`
	Phone     string        `json:"phone"`
	Message   string        `json:"message,omitempty"`
	Products  []OrderProduct `json:"products"`
}

package models

import "time"

const (
	BookStatusCreated   string = "create"
	BookStatusConfirmed string = "confirm"
	BookStatusRejected  string = "reject"
	BookStatusDone      string = "done"
	BookStatusReceive   string = "receive"
)

type BookProduct struct {
	CodeSTU  int    `json:"code_stu"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

type Booking struct {
	Id        int           `json:"id"`
	CreatedAt time.Time     `json:"created_at"`
	Status    string        `json:"status"`
	Username  string        `json:"username"`
	Phone     string        `json:"phone"`
	Message   string        `json:"message,omitempty"`
	Products  []BookProduct `json:"products"`
}

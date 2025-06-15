package entity

import "time"

const (
	BookStatusCreated   string = "create"
	BookStatusConfirmed string = "confirm"
	BookStatusRejected  string = "reject"
	BookStatusDone      string = "done"
	BookStatusReceive   string = "receive"
)

type Book struct {
	Id        int       `json:"id"`
	StoreId   int       `json:"store_id"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"`
	Username  string    `json:"username"`
	Phone     string    `json:"phone"`
	Message   string    `json:"message,omitempty"`
}

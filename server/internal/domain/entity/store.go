package entity

import "time"

type Store struct {
	Id         int       `json:"id"`
	Address    string    `json:"address"`
	UploadTime time.Time `json:"upload_time"`
	Position   Position  `json:"position"`
	Contacts   Contacts  `json:"contacts"`
}

type Position struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type Contacts struct {
	Mobile string `json:"mobile"`
	Email  string `json:"email"`
}

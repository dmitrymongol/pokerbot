package model

import "time"

type User struct {
	ID        int64
	FirstName string
	LastName  string
	Username  string
	CreatedAt time.Time
}
package model

import "time"

type Message struct {
	ID        int64
	Text      string
	UserID    int64
	CreatedAt time.Time
	UpdatedAt time.Time
}
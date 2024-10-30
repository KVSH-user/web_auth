package models

import "time"

type Message struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	MessageText string    `json:"message_text"`
	SenderType  string    `json:"sender_type"`
	CreatedAt   time.Time `json:"created_at"`
}

package models

import "time"

type User struct {
	ID             int64     `json:"id"`
	Email          string    `json:"email"`
	PasswordHashed string    `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
	IsActive       bool      `json:"is_active"`
}

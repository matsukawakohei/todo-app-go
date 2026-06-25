package domain

import "time"

type User struct {
	ID uint
	Email string
	Password string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateUser struct {
	ID uint
	Email string
	Password string
}
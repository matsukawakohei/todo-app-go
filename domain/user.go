package domain

import "time"

type UserID uint

type User struct {
	ID        uint
	Name      string
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateUser struct {
	Name     string
	Email    string
	Password string
}

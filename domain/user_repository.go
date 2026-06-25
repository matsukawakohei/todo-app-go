package domain

import "context"

type UserRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	CreateUser(ctx context.Context, user CreateUser) (*User, error)
}
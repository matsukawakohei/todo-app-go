package domain

import "context"

type IUserRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	CreateUser(ctx context.Context, user CreateUser) (*User, error)
}
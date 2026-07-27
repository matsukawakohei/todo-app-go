package domain

import "context"

type UserRepository interface {
	GetUserByID(ctx context.Context, id UserID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	CreateUser(ctx context.Context, user CreateUser) (*User, error)
}

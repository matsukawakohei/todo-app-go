package domain

type IUserRepository interface {
	GetUserByEmail(email string) (User, error)
	CreateUser(user CreateUser) (User, error)
}
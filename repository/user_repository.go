package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/matsukawakohei/todo-app-go/domain"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

type userModel struct {
	ID uint `gorm:"primaryKey"`
	Name string
	Email string `gorm:"unique"`
	Password string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (um *userModel) toEntity() *domain.User {
	return &domain.User{
		ID: um.ID,
		Name: um.Name,
		Email: um.Email,
		Password: um.Password,
		CreatedAt: um.CreatedAt,
		UpdatedAt: um.UpdatedAt,
	}
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{ db }
}

func (ur *userRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	userModel := userModel{}
	if err := ur.db.Where("email=?", email).First(&userModel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("GetUserByEmail: user not found", email, err)
			return nil, domain.ErrNotFound
		}
		slog.Error(fmt.Sprintf("GetUserByEmail: unexpected error: %v", err))
		return nil, domain.ErrUnexpected
	}

	return userModel.toEntity(), nil
}

func (ur *userRepository) CreateUser(ctx context.Context, user domain.CreateUser) (*domain.User, error) {
	userModel := userModel{
		Name: user.Name,
		Email: user.Email,
		Password: user.Password,
	}
	if err := ur.db.Create(&userModel).Error; err != nil {
		slog.Error(fmt.Sprintf("CreateUser: failed to create user %v", err))
		return nil, domain.ErrUnexpected
	}

	return userModel.toEntity(), nil
}
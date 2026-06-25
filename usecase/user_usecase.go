package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/matsukawakohei/todo-app-go/domain"
	"golang.org/x/crypto/bcrypt"

	jwt "github.com/golang-jwt/jwt/v5"
)

type SignUpInput struct {
	Name     string
	Email    string
	Password string
}

type SignUpOutput struct {
	ID    uint
	Email string
}

func (suo *SignUpOutput) FromUser(user *domain.User) {
	suo.ID = user.ID
	suo.Email = user.Email
}

type LoginInput struct {
	Email    string
	Password string
}

type IUserUsecase interface {
	SignUp(ctx context.Context, input SignUpInput) (SignUpOutput, error)
	Login(ctx context.Context, input LoginInput) (string, error)
}

type userUsecase struct {
	ur domain.IUserRepository
	uv IUserValidator
}

func NewUserUsecase(ur domain.IUserRepository, uv IUserValidator) IUserUsecase {
	return &userUsecase{ur, uv}
}

func (uu *userUsecase) SignUp(ctx context.Context, input SignUpInput) (SignUpOutput, error) {
	if err := uu.uv.SignUpValidate(ctx, input); err != nil {
		slog.With("request_id").Error(fmt.Sprintf("failed validation : %v", err))
		return SignUpOutput{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		slog.With("request_id").Error(fmt.Sprintf("failed to hash password : %v", err))
		return SignUpOutput{}, fmt.Errorf("failed to hash password : %w", domain.ErrUnExpected)
	}

	newUser := domain.CreateUser{Name: input.Name, Email: input.Email, Password: string(hash)}
	user, err := uu.ur.CreateUser(ctx, newUser)
	if err != nil {
		return SignUpOutput{}, err
	}
	output := SignUpOutput{}
	output.FromUser(user)

	return output, nil
}

func (uu *userUsecase) Login(ctx context.Context, input LoginInput) (string, error) {
	if err := uu.uv.LoginValidate(ctx, input); err != nil {
		return "", err
	}

	user, err := uu.ur.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 12).Unix(),
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

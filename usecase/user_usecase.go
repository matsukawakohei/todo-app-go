package usecase

import (
	"context"
	"errors"
	"log/slog"
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
	Name  string
	Email string
}

func (suo *SignUpOutput) FromUser(user *domain.User) {
	suo.ID = user.ID
	suo.Name = user.Name
	suo.Email = user.Email
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginOutput struct {
	ID    uint
	Name  string
	Email string
	Token string
}

type UserUsecase interface {
	SignUp(ctx context.Context, input SignUpInput) (*SignUpOutput, error)
	Login(ctx context.Context, input LoginInput) (*LoginOutput, error)
}

type userUsecase struct {
	ur     domain.UserRepository
	uv     UserValidator
	secret string
}

func NewUserUsecase(ur domain.UserRepository, uv UserValidator, secret string) UserUsecase {
	return &userUsecase{ur, uv, secret}
}

func (uu *userUsecase) SignUp(ctx context.Context, input SignUpInput) (*SignUpOutput, error) {
	if err := uu.uv.SignUpValidate(ctx, input); err != nil {
		slog.ErrorContext(ctx, "failed validation", "error", err)
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash password", "error", err)
		return nil, domain.ErrUnexpected
	}

	newUser := domain.CreateUser{Name: input.Name, Email: input.Email, Password: string(hash)}
	user, err := uu.ur.CreateUser(ctx, newUser)
	if err != nil {
		return nil, err
	}
	output := SignUpOutput{}
	output.FromUser(user)

	return &output, nil
}

func (uu *userUsecase) Login(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	if err := uu.uv.LoginValidate(ctx, input); err != nil {
		return nil, err
	}

	user, err := uu.ur.GetUserByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 12).Unix(),
	})

	tokenString, err := token.SignedString([]byte(uu.secret))
	if err != nil {
		slog.ErrorContext(ctx, "failed to sign jwt", "error", err)
		return nil, domain.ErrUnexpected
	}

	output := LoginOutput{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Token: tokenString,
	}

	return &output, nil
}

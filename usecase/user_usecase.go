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
	ID    uint   `json:"id"`
	Email string `json:"email"`
}

func (suo *SignUpOutput) FromUser(user *domain.User) {
	suo.ID = user.ID
	suo.Email = user.Email
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserUsecase interface {
	SignUp(ctx context.Context, input SignUpInput) (SignUpOutput, error)
	Login(ctx context.Context, input LoginInput) (string, error)
}

type userUsecase struct {
	ur     domain.UserRepository
	uv     UserValidator
	secret string
}

func NewUserUsecase(ur domain.UserRepository, uv UserValidator, secret string) UserUsecase {
	return &userUsecase{ur, uv, secret}
}

func (uu *userUsecase) SignUp(ctx context.Context, input SignUpInput) (SignUpOutput, error) {
	if err := uu.uv.SignUpValidate(ctx, input); err != nil {
		slog.ErrorContext(ctx, "failed validation", "error", err)
		return SignUpOutput{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash password", "error", err)
		return SignUpOutput{}, domain.ErrUnexpected
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
		if errors.Is(err, domain.ErrNotFound) {
			return "", domain.ErrInvalidCredentials
		}
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		return "", domain.ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 12).Unix(),
	})

	tokenString, err := token.SignedString([]byte(uu.secret))
	if err != nil {
		slog.ErrorContext(ctx, "failed to sign jwt", "error", err)
		return "", domain.ErrUnexpected
	}

	return tokenString, nil
}

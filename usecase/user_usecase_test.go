package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/matsukawakohei/todo-app-go/domain"
	"golang.org/x/crypto/bcrypt"
)

const testSecret = "test secret"

type mockUserRepository struct {
	createUserFn     func(ctx context.Context, user domain.CreateUser) (*domain.User, error)
	getUserByEmailFn func(ctx context.Context, email string) (*domain.User, error)
}

func (m *mockUserRepository) CreateUser(ctx context.Context, user domain.CreateUser) (*domain.User, error) {
	return m.createUserFn(ctx, user)
}

func (m *mockUserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return m.getUserByEmailFn(ctx, email)
}

type mockUserValidator struct {
	signUpValidateFn func(ctx context.Context, input SignUpInput) error
	loginValidateFn  func(ctx context.Context, input LoginInput) error
}

func (m *mockUserValidator) SignUpValidate(ctx context.Context, input SignUpInput) error {
	return m.signUpValidateFn(ctx, input)
}

func (m *mockUserValidator) LoginValidate(ctx context.Context, input LoginInput) error {
	return m.loginValidateFn(ctx, input)
}

func TestSignUp(t *testing.T) {
	tests := []struct {
		testName     string
		input        SignUpInput
		validateErr  error
		repoUser     *domain.User
		repoErr      error
		wantErr      bool
		wantOutputID uint
	}{
		{
			testName:     "success",
			input:        SignUpInput{Name: "テスト", Email: "test@example.com", Password: "Password1234"},
			repoUser:     &domain.User{ID: 1, Email: "test@example.com", Name: "テスト"},
			wantOutputID: 1,
		},
		{
			testName:    "validationError",
			input:       SignUpInput{Name: "", Email: "test@example.com", Password: "Password1234"},
			validateErr: errors.New("Name: name is required."),
			wantErr:     true,
		},
		{
			testName: "repositoryError",
			input:    SignUpInput{Name: "テスト", Email: "test@example.com", Password: "Password1234"},
			repoErr:  domain.ErrUnexpected,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			var capturedPassword string
			mockRepo := &mockUserRepository{
				createUserFn: func(ctx context.Context, user domain.CreateUser) (*domain.User, error) {
					capturedPassword = user.Password
					return tc.repoUser, tc.repoErr
				},
				getUserByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
					return nil, nil
				},
			}
			mockValidator := &mockUserValidator{
				signUpValidateFn: func(ctx context.Context, input SignUpInput) error {
					return tc.validateErr
				},
				loginValidateFn: func(ctx context.Context, input LoginInput) error {
					return nil
				},
			}

			uu := NewUserUsecase(mockRepo, mockValidator, testSecret)
			output, err := uu.SignUp(context.Background(), tc.input)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if output.ID != tc.wantOutputID {
				t.Errorf("expect output.ID is %d; but got %d", tc.wantOutputID, output.ID)
			}
			if output.Name != tc.repoUser.Name {
				t.Errorf("expect output.Name is %s; but got %s", tc.repoUser.Name, output.Name)
			}
			if output.Email != tc.repoUser.Email {
				t.Errorf("expect output.Email is %s; but got %s", tc.repoUser.Email, output.Email)
			}
			if capturedPassword == tc.input.Password {
				t.Errorf("password should be hashed, but got plain text")
			}
			if err := bcrypt.CompareHashAndPassword([]byte(capturedPassword), []byte(tc.input.Password)); err != nil {
				t.Errorf("stored password is not a valid bcrypt hash: %v", err)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	const rawPassword = "Password1234"
	hash, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("failed to generate bcrypt hash: %v", err)
	}
	hashedPassword := string(hash)

	tests := []struct {
		testName    string
		secret      string
		input       LoginInput
		validateErr error
		repoUser    *domain.User
		repoErr     error
		wantErr     bool
		wantErrIs   error
	}{
		{
			testName: "success",
			secret:   "test-secret",
			input:    LoginInput{Email: "test@example.com", Password: rawPassword},
			repoUser: &domain.User{ID: 1, Name: "テスト", Email: "test@example.com", Password: hashedPassword},
		},
		{
			testName:    "validationError",
			secret:      "test-secret",
			input:       LoginInput{Email: "", Password: rawPassword},
			validateErr: errors.New("Email: email is required."),
			wantErr:     true,
		},
		{
			testName:  "userNotFound",
			secret:    "test-secret",
			input:     LoginInput{Email: "notfound@example.com", Password: rawPassword},
			repoErr:   domain.ErrNotFound,
			wantErr:   true,
			wantErrIs: domain.ErrInvalidCredentials,
		},
		{
			testName:  "passwordMismatch",
			secret:    "test-secret",
			input:     LoginInput{Email: "test@example.com", Password: "WrongPass99"},
			repoUser:  &domain.User{ID: 1, Email: "test@example.com", Password: hashedPassword},
			wantErr:   true,
			wantErrIs: domain.ErrInvalidCredentials,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {

			mockRepo := &mockUserRepository{
				createUserFn: func(ctx context.Context, user domain.CreateUser) (*domain.User, error) {
					return nil, nil
				},
				getUserByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
					return tc.repoUser, tc.repoErr
				},
			}
			mockValidator := &mockUserValidator{
				signUpValidateFn: func(ctx context.Context, input SignUpInput) error {
					return nil
				},
				loginValidateFn: func(ctx context.Context, input LoginInput) error {
					return tc.validateErr
				},
			}

			uu := NewUserUsecase(mockRepo, mockValidator, tc.secret)
			output, err := uu.Login(context.Background(), tc.input)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				if tc.wantErrIs != nil && !errors.Is(err, tc.wantErrIs) {
					t.Errorf("expected error %v; but got %v", tc.wantErrIs, err)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			parsed, parseErr := jwt.Parse(output.Token, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return []byte("test-secret"), nil
			})
			if parseErr != nil {
				t.Errorf("failed to parse token: %v", parseErr)
				return
			}
			claims, ok := parsed.Claims.(jwt.MapClaims)
			if !ok || !parsed.Valid {
				t.Errorf("token is invalid")
				return
			}
			if userID, ok := claims["user_id"]; !ok {
				t.Errorf("user_id claim is missing")
			} else if userID != float64(tc.repoUser.ID) {
				t.Errorf("expect user_id %v; but got %v", float64(tc.repoUser.ID), userID)
			}
			if output.ID != tc.repoUser.ID {
				t.Errorf("expect output.ID is %d; but got %d", tc.repoUser.ID, output.ID)
			}
			if output.Name != tc.repoUser.Name {
				t.Errorf("expect output.Name is %s; but got %s", tc.repoUser.Name, output.Name)
			}
			if output.Email != tc.repoUser.Email {
				t.Errorf("expect output.Email is %s; but got %s", tc.repoUser.Email, output.Email)
			}
		})
	}
}

package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/matsukawakohei/todo-app-go/domain"
	"golang.org/x/crypto/bcrypt"
)

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
			testName: "success",
			input:    SignUpInput{Name: "テスト", Email: "test@example.com", Password: "Password1234"},
			repoUser: &domain.User{ID: 1, Email: "test@example.com"},
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

			uu := NewUserUsecase(mockRepo, mockValidator)
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
		input       LoginInput
		validateErr error
		repoUser    *domain.User
		repoErr     error
		wantErr     bool
	}{
		{
			testName: "success",
			input:    LoginInput{Email: "test@example.com", Password: rawPassword},
			repoUser: &domain.User{ID: 1, Email: "test@example.com", Password: hashedPassword},
		},
		{
			testName:    "validationError",
			input:       LoginInput{Email: "", Password: rawPassword},
			validateErr: errors.New("Email: email is required."),
			wantErr:     true,
		},
		{
			testName: "userNotFound",
			input:    LoginInput{Email: "notfound@example.com", Password: rawPassword},
			repoErr:  domain.ErrNotFound,
			wantErr:  true,
		},
		{
			testName: "passwordMismatch",
			input:    LoginInput{Email: "test@example.com", Password: "WrongPass99"},
			repoUser: &domain.User{ID: 1, Email: "test@example.com", Password: hashedPassword},
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			t.Setenv("SECRET", "test-secret")

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

			uu := NewUserUsecase(mockRepo, mockValidator)
			token, err := uu.Login(context.Background(), tc.input)

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
			if token == "" {
				t.Errorf("expected non-empty token but got empty string")
			}
		})
	}
}

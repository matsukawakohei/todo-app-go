package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/matsukawakohei/todo-app-go/domain"
	"github.com/matsukawakohei/todo-app-go/usecase"
)

type mockUserUsecase struct {
	signUpFn func(ctx context.Context, input usecase.SignUpInput) (usecase.SignUpOutput, error)
	loginFn  func(ctx context.Context, input usecase.LoginInput) (string, error)
}

func (m *mockUserUsecase) SignUp(ctx context.Context, input usecase.SignUpInput) (usecase.SignUpOutput, error) {
	return m.signUpFn(ctx, input)
}

func (m *mockUserUsecase) Login(ctx context.Context, input usecase.LoginInput) (string, error) {
	return m.loginFn(ctx, input)
}

func newTestContext(method, body string) (*echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestSignUp(t *testing.T) {
	tests := []struct {
		testName   string
		body       string
		signUpFn   func(ctx context.Context, input usecase.SignUpInput) (usecase.SignUpOutput, error)
		wantStatus int
		wantID     uint
		wantEmail  string
	}{
		{
			testName: "success",
			body:     `{"Name":"テスト","Email":"test@example.com","Password":"password123"}`,
			signUpFn: func(_ context.Context, _ usecase.SignUpInput) (usecase.SignUpOutput, error) {
				return usecase.SignUpOutput{ID: 1, Email: "test@example.com"}, nil
			},
			wantStatus: http.StatusCreated,
			wantID:     1,
			wantEmail:  "test@example.com",
		},
		{
			testName:   "bindError",
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			testName: "usecaseErrBadRequest",
			body:     `{"Name":"テスト","Email":"test@example.com","Password":"password123"}`,
			signUpFn: func(_ context.Context, _ usecase.SignUpInput) (usecase.SignUpOutput, error) {
				return usecase.SignUpOutput{}, domain.ErrBadRequest
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			testName: "usecaseErrDuplicatedKey",
			body:     `{"Name":"テスト","Email":"test@example.com","Password":"password123"}`,
			signUpFn: func(_ context.Context, _ usecase.SignUpInput) (usecase.SignUpOutput, error) {
				return usecase.SignUpOutput{}, domain.ErrDuplicatedKey
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			testName: "usecaseErrUnexpected",
			body:     `{"Name":"テスト","Email":"test@example.com","Password":"password123"}`,
			signUpFn: func(_ context.Context, _ usecase.SignUpInput) (usecase.SignUpOutput, error) {
				return usecase.SignUpOutput{}, domain.ErrUnexpected
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			c, rec := newTestContext(http.MethodPost, tc.body)
			uc := NewUserController(&mockUserUsecase{signUpFn: tc.signUpFn})
			uc.SignUp(c)

			if rec.Code != tc.wantStatus {
				t.Errorf("expected status %d; got %d", tc.wantStatus, rec.Code)
			}

			if tc.wantID != 0 || tc.wantEmail != "" {
				var output usecase.SignUpOutput
				if err := json.Unmarshal(rec.Body.Bytes(), &output); err != nil {
					t.Fatalf("failed to unmarshal response body: %v", err)
				}
				if output.ID != tc.wantID {
					t.Errorf("expected ID %d; got %d", tc.wantID, output.ID)
				}
				if output.Email != tc.wantEmail {
					t.Errorf("expected Email %q; got %q", tc.wantEmail, output.Email)
				}
			}
		})
	}
}

func TestLogin(t *testing.T) {
	tests := []struct {
		testName   string
		body       string
		loginFn    func(ctx context.Context, input usecase.LoginInput) (string, error)
		wantStatus int
		wantToken  string
	}{
		{
			testName: "success",
			body:     `{"email":"test@example.com","password":"password123"}`,
			loginFn: func(_ context.Context, _ usecase.LoginInput) (string, error) {
				return "jwt.token.string", nil
			},
			wantStatus: http.StatusOK,
			wantToken:  "jwt.token.string",
		},
		{
			testName:   "bindError",
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			testName: "usecaseErrInvalidCredentials",
			body:     `{"email":"test@example.com","password":"wrong"}`,
			loginFn: func(_ context.Context, _ usecase.LoginInput) (string, error) {
				return "", domain.ErrInvalidCredentials
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			testName: "usecaseErrUnexpected",
			body:     `{"email":"test@example.com","password":"password123"}`,
			loginFn: func(_ context.Context, _ usecase.LoginInput) (string, error) {
				return "", domain.ErrUnexpected
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			c, rec := newTestContext(http.MethodPost, tc.body)
			uc := NewUserController(&mockUserUsecase{loginFn: tc.loginFn})
			uc.Login(c)

			if rec.Code != tc.wantStatus {
				t.Errorf("expected status %d; got %d", tc.wantStatus, rec.Code)
			}

			if tc.wantToken != "" {
				var tokenFound bool
				for _, cookie := range rec.Result().Cookies() {
					if cookie.Name == "token" {
						tokenFound = true
						if cookie.Value != tc.wantToken {
							t.Errorf("expected cookie token %q; got %q", tc.wantToken, cookie.Value)
						}
					}
				}
				if !tokenFound {
					t.Errorf("expected cookie 'token' to be set")
				}
			}
		})
	}
}

func TestLogout(t *testing.T) {
	c, rec := newTestContext(http.MethodPost, "")
	uc := NewUserController(&mockUserUsecase{})
	uc.Logout(c)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d; got %d", http.StatusOK, rec.Code)
	}

	var tokenFound bool
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == "token" {
			tokenFound = true
			if cookie.Value != "" {
				t.Errorf("expected cookie token to be empty; got %q", cookie.Value)
			}
		}
	}
	if !tokenFound {
		t.Errorf("expected cookie 'token' to be cleared")
	}
}

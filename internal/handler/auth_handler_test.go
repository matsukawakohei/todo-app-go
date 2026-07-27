package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matsukawakohei/todo-app-go/domain"
	"github.com/matsukawakohei/todo-app-go/internal/api"
	"github.com/matsukawakohei/todo-app-go/usecase"
	"github.com/oapi-codegen/runtime/types"
)

type mockUserUsecase struct {
	signUpFn      func(ctx context.Context, input usecase.SignUpInput) (*usecase.SignUpOutput, error)
	loginFn       func(ctx context.Context, input usecase.LoginInput) (*usecase.LoginOutput, error)
	getUserByIDFn func(ctx context.Context, userID domain.UserID) (*usecase.GetUserByIDOutput, error)
}

func (m *mockUserUsecase) SignUp(ctx context.Context, input usecase.SignUpInput) (*usecase.SignUpOutput, error) {
	return m.signUpFn(ctx, input)
}

func (m *mockUserUsecase) Login(ctx context.Context, input usecase.LoginInput) (*usecase.LoginOutput, error) {
	return m.loginFn(ctx, input)
}

func (m *mockUserUsecase) GetUserByID(ctx context.Context, userID domain.UserID) (*usecase.GetUserByIDOutput, error) {
	return m.getUserByIDFn(ctx, userID)
}

func TestSignUp(t *testing.T) {
	tests := []struct {
		testName   string
		signUpFn   func(ctx context.Context, input usecase.SignUpInput) (*usecase.SignUpOutput, error)
		wantStatus int
		wantID     int
		wantName   string
		wantEmail  string
	}{
		{
			testName: "success",
			signUpFn: func(_ context.Context, _ usecase.SignUpInput) (*usecase.SignUpOutput, error) {
				return &usecase.SignUpOutput{ID: 1, Name: "テスト", Email: "test@example.com"}, nil
			},
			wantStatus: http.StatusCreated,
			wantID:     1,
			wantName:   "テスト",
			wantEmail:  "test@example.com",
		},
		{
			testName: "usecaseErrBadRequest",
			signUpFn: func(_ context.Context, _ usecase.SignUpInput) (*usecase.SignUpOutput, error) {
				return nil, domain.ErrBadRequest
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			testName: "usecaseErrDuplicatedKey",
			signUpFn: func(_ context.Context, _ usecase.SignUpInput) (*usecase.SignUpOutput, error) {
				return nil, domain.ErrDuplicatedKey
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			testName: "usecaseErrUnexpected",
			signUpFn: func(_ context.Context, _ usecase.SignUpInput) (*usecase.SignUpOutput, error) {
				return nil, domain.ErrUnexpected
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			ah := NewAuthHandler(&mockUserUsecase{signUpFn: tc.signUpFn})
			request := api.SignupRequestObject{
				Body: &api.SignupInput{Name: "テスト", Email: types.Email("test@example.com"), Password: "password123"},
			}
			response, err := ah.Signup(context.Background(), request)
			if err != nil {
				t.Fatalf("auth handler SignUp failed: %v", err)
			}

			switch tc.wantStatus {
			case http.StatusCreated:
				got, ok := response.(api.Signup201JSONResponse)
				if !ok {
					t.Fatalf("expected api.Signup201JSONResponse; got %T", response)
				}
				if got.Data.Id != tc.wantID {
					t.Errorf("expected ID %d; got %d", tc.wantID, got.Data.Id)
				}
				if got.Data.Name != tc.wantName {
					t.Errorf("expected Name %s; but got %s", tc.wantName, got.Data.Name)
				}
				if string(got.Data.Email) != tc.wantEmail {
					t.Errorf("expected Email %q; got %q", tc.wantEmail, got.Data.Email)
				}
			case http.StatusBadRequest:
				if _, ok := response.(api.Signup400JSONResponse); !ok {
					t.Fatalf("expected api.Signup400JSONResponse; got %T", response)
				}
			case http.StatusInternalServerError:
				if _, ok := response.(api.Signup500JSONResponse); !ok {
					t.Fatalf("expected api.Signup500JSONResponse; got %T", response)
				}
			}
		})
	}
}

func TestLogin(t *testing.T) {
	tests := []struct {
		testName   string
		loginFn    func(ctx context.Context, input usecase.LoginInput) (*usecase.LoginOutput, error)
		wantStatus int
		wantToken  string
		wantID     int
		wantName   string
		wantEmail  string
	}{
		{
			testName: "success",
			loginFn: func(_ context.Context, _ usecase.LoginInput) (*usecase.LoginOutput, error) {
				return &usecase.LoginOutput{
					ID:    1,
					Name:  "test",
					Email: "test@example.com",
					Token: "jwt.token.string",
				}, nil
			},
			wantStatus: http.StatusOK,
			wantToken:  "jwt.token.string",
			wantID:     1,
			wantName:   "test",
			wantEmail:  "test@example.com",
		},
		{
			testName: "usecaseErrInvalidCredentials",
			loginFn: func(_ context.Context, _ usecase.LoginInput) (*usecase.LoginOutput, error) {
				return nil, domain.ErrInvalidCredentials
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			testName: "usecaseErrUnexpected",
			loginFn: func(_ context.Context, _ usecase.LoginInput) (*usecase.LoginOutput, error) {
				return nil, domain.ErrUnexpected
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			ah := NewAuthHandler(&mockUserUsecase{loginFn: tc.loginFn})
			request := api.LoginRequestObject{
				Body: &api.LoginInput{Email: types.Email("test@example.com"), Password: "password123"},
			}
			response, err := ah.Login(context.Background(), request)
			if err != nil {
				t.Fatalf("auth handler Login failed: %v", err)
			}

			switch tc.wantStatus {
			case http.StatusOK:
				got, ok := response.(loginJSONResponse)
				if !ok {
					t.Fatalf("expected loginJSONResponse; got %T", response)
				}
				if got.cookie == nil || got.cookie.Value != tc.wantToken {
					t.Errorf("expected cookie token %q; got %v", tc.wantToken, got.cookie)
				}
				if got.body.Data.Id != tc.wantID {
					t.Errorf("expected ID %d; got %d", tc.wantID, got.body.Data.Id)
				}
				if got.body.Data.Name != tc.wantName {
					t.Errorf("expected Name %s; but got %s", tc.wantName, got.body.Data.Name)
				}
				if string(got.body.Data.Email) != tc.wantEmail {
					t.Errorf("expected Email %q; got %q", tc.wantEmail, got.body.Data.Email)
				}

				rec := httptest.NewRecorder()
				if err := got.VisitLoginResponse(rec); err != nil {
					t.Fatalf("VisitLoginResponse failed: %v", err)
				}
				if rec.Code != http.StatusOK {
					t.Errorf("expected status %d; got %d", http.StatusOK, rec.Code)
				}
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
			case http.StatusUnauthorized:
				if _, ok := response.(api.Login401JSONResponse); !ok {
					t.Fatalf("expected api.Login401JSONResponse; got %T", response)
				}
			case http.StatusInternalServerError:
				if _, ok := response.(api.Login500JSONResponse); !ok {
					t.Fatalf("expected api.Login500JSONResponse; got %T", response)
				}
			}
		})
	}
}

func TestLogout(t *testing.T) {
	ah := NewAuthHandler(&mockUserUsecase{})
	response, err := ah.Logout(context.Background(), api.LogoutRequestObject{})
	if err != nil {
		t.Fatalf("auth handler Logout failed: %v", err)
	}

	rec := httptest.NewRecorder()
	if err := response.VisitLogoutResponse(rec); err != nil {
		t.Fatalf("VisitLogoutResponse failed: %v", err)
	}

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

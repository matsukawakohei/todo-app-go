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

func TestMe(t *testing.T) {
	tests := []struct {
		testName   string
		mockOutput *usecase.GetUserByIDOutput
		mockErr    error
		wantStatus int
		wantUser   *api.User
	}{
		{
			testName: "success",
			mockOutput: &usecase.GetUserByIDOutput{
				ID:    1,
				Name:  "test",
				Email: "test1@example.com",
			},
			wantStatus: http.StatusOK,
			wantUser: &api.User{
				Id:    1,
				Name:  "test",
				Email: "test1@example.com",
			},
		},
		{
			testName:   "notFound",
			mockErr:    domain.ErrNotFound,
			wantStatus: http.StatusUnauthorized,
		},
		{
			testName:   "internalServerError",
			mockErr:    domain.ErrUnexpected,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			var gotUserID domain.UserID
			ctx := contextWithUser(1)
			ah := &AuthHandler{
				uu: &mockUserUsecase{
					getUserByIDFn: func(ctx context.Context, userID domain.UserID) (*usecase.GetUserByIDOutput, error) {
						gotUserID = userID
						return tc.mockOutput, tc.mockErr
					},
				},
			}

			output, err := ah.Me(ctx, api.MeRequestObject{})
			if err != nil {
				t.Fatalf("want err nil; but got %v", err)
			}
			if gotUserID != 1 {
				t.Errorf("want user ID 1; but got %d", gotUserID)
			}

			switch tc.wantStatus {
			case http.StatusOK:
				got, ok := output.(api.Me200JSONResponse)
				if !ok {
					t.Fatalf("want api.Me200JSONResponse; got %T", output)
				}
				if tc.wantUser != nil && got.Data != *tc.wantUser {
					t.Errorf("want output %+v; but got %+v", tc.wantUser, got.Data)
				}
			case http.StatusUnauthorized:
				_, ok := output.(api.Me401JSONResponse)
				if !ok {
					t.Fatalf("want api.Me401JSONResponse; got %T", output)
				}
			case http.StatusInternalServerError:
				_, ok := output.(api.Me500JSONResponse)
				if !ok {
					t.Fatalf("want api.Me500JSONResponse; got %T", output)
				}
			}
		})
	}
}

package controller

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
	"github.com/matsukawakohei/todo-app-go/domain"
)

func newTestContext(method, body string) (*echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), method, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func setLoginUser(c *echo.Context, userID uint) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": float64(userID)})
	c.Set("user", token)
}

func TestGetLoginUserFromContext(t *testing.T) {
	tests := []struct {
		testName    string
		token       *jwt.Token
		expected    uint
		wantError   bool
		wantErrorIs error
	}{
		{
			testName:    "EmptyJWT",
			wantError:   true,
			wantErrorIs: domain.ErrUnexpected,
		},
		{
			testName:    "ClaimsIsNotMapClaims",
			token:       jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{}),
			wantError:   true,
			wantErrorIs: domain.ErrUnexpected,
		},
		{
			testName:    "ClaimsNotHaveUserID",
			token:       jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{}),
			wantError:   true,
			wantErrorIs: domain.ErrUnexpected,
		},
		{
			testName:    "user_idIsString",
			token:       jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": "a"}),
			wantError:   true,
			wantErrorIs: domain.ErrUnexpected,
		},
		{
			testName:    "user_idIsNegative",
			token:       jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": float64(-1)}),
			wantError:   true,
			wantErrorIs: domain.ErrUnexpected,
		},
		{
			testName:  "success",
			token:     jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": float64(1)}),
			expected:  1,
			wantError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			c, _ := newTestContext("GET", "/")
			if tc.token != nil {
				c.Set("user", tc.token)
			}
			ret, err := getLoginUserFromContext(c)
			if tc.wantError {
				if err == nil {
					t.Fatalf("expect err is not nil; but got nil")
				}
				if !errors.Is(err, tc.wantErrorIs) {
					t.Errorf("expect err is %v; but got %v", tc.wantErrorIs, err)
				}
			} else {
				if err != nil {
					t.Fatalf("expect err is nil; but got %v", err)
				}
				if ret != tc.expected {
					t.Errorf("user_id expected %v; but got %v", tc.expected, ret)
				}
			}
		})
	}
}

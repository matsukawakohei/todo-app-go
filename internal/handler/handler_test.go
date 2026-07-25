package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/matsukawakohei/todo-app-go/domain"
)

func contextWithUser(userID uint) context.Context {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": float64(userID)})
	return ContextWithUserToken(context.Background(), token)
}

func TestGetLoginUserFromContext(t *testing.T) {
	tests := []struct {
		testName    string
		ctx         context.Context
		expected    uint
		wantError   bool
		wantErrorIs error
	}{
		{
			testName:    "EmptyContext",
			ctx:         context.Background(),
			wantError:   true,
			wantErrorIs: domain.ErrUnexpected,
		},
		{
			testName:    "ClaimsIsNotMapClaims",
			ctx:         ContextWithUserToken(context.Background(), jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{})),
			wantError:   true,
			wantErrorIs: domain.ErrUnexpected,
		},
		{
			testName:    "ClaimsNotHaveUserID",
			ctx:         ContextWithUserToken(context.Background(), jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{})),
			wantError:   true,
			wantErrorIs: domain.ErrUnexpected,
		},
		{
			testName:    "user_idIsString",
			ctx:         ContextWithUserToken(context.Background(), jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": "a"})),
			wantError:   true,
			wantErrorIs: domain.ErrUnexpected,
		},
		{
			testName:    "user_idIsNegative",
			ctx:         ContextWithUserToken(context.Background(), jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": float64(-1)})),
			wantError:   true,
			wantErrorIs: domain.ErrUnexpected,
		},
		{
			testName:  "success",
			ctx:       contextWithUser(1),
			expected:  1,
			wantError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			ret, err := getLoginUserFromContext(tc.ctx)
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

package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/matsukawakohei/todo-app-go/domain"
)

type Handler struct {
	*AuthHandler
	*LabelHandler
	*TaskHandler
}

type contextKey int

const userTokenContextKey contextKey = iota

// ContextWithUserToken はJWTトークンを保持したcontext.Contextを返します。
// echo-jwtのSuccessHandlerから呼び出され、認証済みリクエストのcontext.Contextに
// トークンを紐付けるために使用します。
func ContextWithUserToken(ctx context.Context, token *jwt.Token) context.Context {
	return context.WithValue(ctx, userTokenContextKey, token)
}

// CookieSecureは環境変数から実行環境を読み取りCookieのSecure属性を返します
// 環境変数GO_ENVの値がdevの場合にfalseを返し、それ以外はtrueを返します
func CookieSecure() bool {
	return os.Getenv("GO_ENV") != "dev"
}

func getLoginUserFromContext(ctx context.Context) (uint, error) {
	user, ok := ctx.Value(userTokenContextKey).(*jwt.Token)
	if !ok {
		slog.ErrorContext(ctx, "Failed get JWT token from context")
		return 0, domain.ErrUnexpected
	}
	claims, ok := user.Claims.(jwt.MapClaims)
	if !ok {
		slog.ErrorContext(ctx, "Failed get claims from JWT token")
		return 0, domain.ErrUnexpected
	}
	userIDValue, ok := claims["user_id"]
	if !ok {
		slog.ErrorContext(ctx, "Failed get user_id from JWT Claims")
		return 0, domain.ErrUnexpected
	}
	userID, ok := userIDValue.(float64)
	if !ok {
		slog.ErrorContext(ctx, "Failed parse user_id to float64 from any")
		return 0, domain.ErrUnexpected
	}
	if userID < 0 {
		slog.ErrorContext(ctx, "user_id claim is negative")
		return 0, domain.ErrUnexpected
	}

	return uint(userID), nil
}

// errorStatus はdomainエラーを対応するHTTPステータスコードに変換します。
func errorStatus(err error) int {
	if errors.Is(err, domain.ErrDuplicatedKey) {
		return http.StatusBadRequest
	}
	if errors.Is(err, domain.ErrNotFound) {
		return http.StatusNotFound
	}
	if errors.Is(err, domain.ErrUnexpected) {
		return http.StatusInternalServerError
	}
	if errors.Is(err, domain.ErrInvalidCredentials) {
		return http.StatusUnauthorized
	}
	if errors.Is(err, domain.ErrLimitExceeded) {
		return http.StatusBadRequest
	}
	if errors.Is(err, domain.ErrBadRequest) {
		return http.StatusBadRequest
	}

	return http.StatusInternalServerError
}

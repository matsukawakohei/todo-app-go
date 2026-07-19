package controller

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
	"github.com/matsukawakohei/todo-app-go/domain"
)

func buildResponseError(c *echo.Context, err error) error {
	if errors.Is(err, domain.ErrDuplicatedKey) {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if errors.Is(err, domain.ErrNotFound) {
		return c.JSON(http.StatusNotFound, err.Error())
	}
	if errors.Is(err, domain.ErrUnexpected) {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	if errors.Is(err, domain.ErrInvalidCredentials) {
		return c.JSON(http.StatusUnauthorized, err.Error())
	}
	if errors.Is(err, domain.ErrLimitExceeded) {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusBadRequest, err.Error())
}

// CookieSecureは環境変数から実行環境を読み取りCookieのSecure属性を返します
// 環境変数GO_ENVの値がdevの場合にfalseを返し、それ以外はtrueを返します
func CookieSecure() bool {
	return os.Getenv("GO_ENV") != "dev"
}

func getLoginUserFromContext(c *echo.Context) (uint, error) {
	user, ok := c.Get("user").(*jwt.Token)
	if !ok {
		slog.ErrorContext(c.Request().Context(), "Failed get JWT token from context")
		return 0, domain.ErrUnexpected
	}
	claims, ok := user.Claims.(jwt.MapClaims)
	if !ok {
		slog.ErrorContext(c.Request().Context(), "Failed get claims from JWT token")
		return 0, domain.ErrUnexpected
	}
	userIDValue, ok := claims["user_id"]
	if !ok {
		slog.ErrorContext(c.Request().Context(), "Failed get user_id from JWT Claims")
		return 0, domain.ErrUnexpected
	}
	userID, ok := userIDValue.(float64)
	if !ok {
		slog.ErrorContext(c.Request().Context(), "Failed parse user_id to float64 from any")
		return 0, domain.ErrUnexpected
	}
	if userID < 0 {
		slog.ErrorContext(c.Request().Context(), "user_id claim is negative")
		return 0, domain.ErrUnexpected
	}

	return uint(userID), nil
}

func getIDFromContext(c *echo.Context, paramName string) (uint, error) {
	idString := c.Param(paramName)
	ID, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		msg := fmt.Sprintf("Failed parse %s to int from string", paramName)
		slog.ErrorContext(c.Request().Context(), msg, "error", err)
		return 0, domain.ErrBadRequest
	}

	return uint(ID), nil
}

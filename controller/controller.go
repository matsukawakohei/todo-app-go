package controller

import (
	"errors"
	"net/http"

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
	
	return c.JSON(http.StatusBadRequest, err.Error())
}

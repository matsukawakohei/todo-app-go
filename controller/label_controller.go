package controller

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/matsukawakohei/todo-app-go/domain"
	"github.com/matsukawakohei/todo-app-go/usecase"
)

type LabelController interface {
	GetAllLabels(c *echo.Context) error
	CreateLabel(c *echo.Context) error
	DeleteLabel(c *echo.Context) error
}

type labelController struct {
	lu usecase.LabelUsecase
}

func NewLabelController(lu usecase.LabelUsecase) LabelController {
	return &labelController{lu}
}

func (lc *labelController) GetAllLabels(c *echo.Context) error {
	userID, err := getLoginUserFromContext(c)
	if err != nil {
		return buildResponseError(c, err)
	}

	labels, err := lc.lu.GetAllByUserID(c.Request().Context(), domain.UserID(userID))
	if err != nil {
		return buildResponseError(c, err)
	}

	return c.JSON(http.StatusOK, labels)
}

func (lc *labelController) CreateLabel(c *echo.Context) error {
	userID, err := getLoginUserFromContext(c)
	if err != nil {
		return buildResponseError(c, err)
	}

	ctx := c.Request().Context()
	input := usecase.LabelCreateInput{}
	if err := c.Bind(&input); err != nil {
		slog.ErrorContext(ctx, "Failed bind input from create label request", "error", err)
		return buildResponseError(c, err)
	}

	label, err := lc.lu.Create(ctx, domain.UserID(userID), input)
	if err != nil {
		return buildResponseError(c, err)
	}

	return c.JSON(http.StatusCreated, label)
}

func (lc *labelController) DeleteLabel(c *echo.Context) error {
	userID, err := getLoginUserFromContext(c)
	if err != nil {
		return buildResponseError(c, err)
	}
	labelID, err := getLabelIDFromParam(c)
	if err != nil {
		return buildResponseError(c, err)
	}

	ctx := c.Request().Context()
	if err := lc.lu.Delete(ctx, domain.UserID(userID), domain.LabelID(labelID)); err != nil {
		return buildResponseError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

func getLabelIDFromParam(c *echo.Context) (uint, error) {
	labelIDString := c.Param("id")
	labelID, err := strconv.Atoi(labelIDString)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "Failed parse label_id to int from string", "error", err)
		return 0, domain.ErrBadRequest
	}

	return uint(labelID), nil
}

package controller

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/matsukawakohei/todo-app-go/domain"
	"github.com/matsukawakohei/todo-app-go/usecase"
)

type TaskController interface {
	GetAllTasksByUserID(c *echo.Context) error
	CreateTask(c *echo.Context) error
	UpdateTask(c *echo.Context) error
	DeleteTask(c *echo.Context) error
}

type taskController struct {
	tu usecase.TaskUsecase
}

func NewTaskController(tu usecase.TaskUsecase) TaskController {
	return &taskController{ tu }
}

func (tc *taskController) GetAllTasksByUserID(c *echo.Context) error {
	userID, err := getLoginUserFromContext(c)
	if err != nil {
		return buildResponseError(c, err)
	}

	tasks, err := tc.tu.GetAllTasksByUserID(c.Request().Context(), domain.UserID(userID))
	if err != nil {
		return buildResponseError(c, err)
	}

	return c.JSON(http.StatusOK, tasks)
}

func (tc *taskController) CreateTask(c *echo.Context) error {
	userID, err := getLoginUserFromContext(c)
	if err != nil {
		return buildResponseError(c, err)
	}

	ctx := c.Request().Context()
	input := usecase.CreateTaskInput{}
	if err := c.Bind(&input); err != nil {
		slog.ErrorContext(ctx, "failed bind input from create task request", "error", err)
		return buildResponseError(c, err)
	}

	task, err := tc.tu.CreateTask(ctx, domain.UserID(userID), input)
	if err != nil {
		return buildResponseError(c, err)
	}

	return c.JSON(http.StatusCreated, task)
}

func (tc *taskController) UpdateTask(c *echo.Context) error {
	userID, err := getLoginUserFromContext(c)
	if err != nil {
		return buildResponseError(c, err)
	}

	taskID, err := getIDFromContext(c, "id")
	if err != nil {
		return buildResponseError(c, err)
	}

	ctx := c.Request().Context()
	input := usecase.UpdateTaskInput{}
	if err := c.Bind(&input); err != nil {
		slog.ErrorContext(ctx, "failed bind input from update task request", "error", err)
		return buildResponseError(c, err)
	}

	task, err := tc.tu.UpdateTask(ctx, domain.UserID(userID), domain.TaskID(taskID), input)
	if err != nil {
		return buildResponseError(c, err)
	}

	return c.JSON(http.StatusOK, task)
}

func (tc *taskController) DeleteTask(c *echo.Context) error {
	userID, err := getLoginUserFromContext(c)
	if err != nil {
		return buildResponseError(c, err)
	}

	taskID, err := getIDFromContext(c, "id")
	if err != nil {
		return buildResponseError(c, err)
	}

	ctx := c.Request().Context()
	err = tc.tu.DeleteTask(ctx, domain.UserID(userID), domain.TaskID(taskID))
	if err != nil {
		return buildResponseError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}
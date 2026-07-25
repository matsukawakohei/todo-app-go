package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/matsukawakohei/todo-app-go/domain"
	"github.com/matsukawakohei/todo-app-go/internal/api"
	"github.com/matsukawakohei/todo-app-go/usecase"
)

type TaskHandler struct {
	tu usecase.TaskUsecase
}

func NewTaskHandler(tu usecase.TaskUsecase) *TaskHandler {
	return &TaskHandler{tu}
}

func toTaskResponse(task *usecase.Task) api.Task {
	labels := make([]api.Label, 0, len(task.Labels))
	for _, l := range task.Labels {
		labels = append(labels, toLabelResponse(l))
	}

	return api.Task{
		Id:        int(task.ID),
		Title:     task.Title,
		Completed: task.Completed,
		Labels:    labels,
	}
}

func toLabelIDs(ids []int) []uint {
	labelIDs := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id < 0 {
			continue
		}
		labelIDs = append(labelIDs, uint(id))
	}
	return labelIDs
}

func (th *TaskHandler) GetTasks(ctx context.Context, _ api.GetTasksRequestObject) (api.GetTasksResponseObject, error) {
	userID, err := getLoginUserFromContext(ctx)
	if err != nil {
		return api.GetTasks500JSONResponse{Error: err.Error()}, nil
	}

	output, err := th.tu.GetAllTasksByUserID(ctx, domain.UserID(userID))
	if err != nil {
		return api.GetTasks500JSONResponse{Error: err.Error()}, nil
	}

	tasks := make([]api.Task, 0, len(output.Tasks))
	for _, t := range output.Tasks {
		tasks = append(tasks, toTaskResponse(t))
	}

	return api.GetTasks200JSONResponse{Data: tasks}, nil
}

func (th *TaskHandler) CreateTask(ctx context.Context, request api.CreateTaskRequestObject) (api.CreateTaskResponseObject, error) {
	if request.Body == nil {
		slog.ErrorContext(ctx, "request body is nil")
		return api.CreateTask400JSONResponse{Error: domain.ErrBadRequest.Error()}, nil
	}

	userID, err := getLoginUserFromContext(ctx)
	if err != nil {
		return api.CreateTask500JSONResponse{Error: err.Error()}, nil
	}

	input := usecase.CreateTaskInput{
		Title:  request.Body.Title,
		Labels: toLabelIDs(request.Body.Labels),
	}
	output, err := th.tu.CreateTask(ctx, domain.UserID(userID), input)
	if err != nil {
		message := err.Error()
		if errorStatus(err) == http.StatusInternalServerError {
			return api.CreateTask500JSONResponse{Error: message}, nil
		}
		return api.CreateTask400JSONResponse{Error: message}, nil
	}

	return api.CreateTask201JSONResponse{Data: toTaskResponse(output)}, nil
}

func (th *TaskHandler) UpdateTask(ctx context.Context, request api.UpdateTaskRequestObject) (api.UpdateTaskResponseObject, error) {
	if request.Body == nil {
		slog.ErrorContext(ctx, "request body is nil")
		return api.UpdateTask400JSONResponse{Error: domain.ErrBadRequest.Error()}, nil
	}
	userID, err := getLoginUserFromContext(ctx)
	if err != nil {
		return api.UpdateTask500JSONResponse{Error: err.Error()}, nil
	}

	if request.Id <= 0 {
		slog.ErrorContext(ctx, "id less than equals 0")
		return api.UpdateTask400JSONResponse{Error: domain.ErrBadRequest.Error()}, nil
	}

	input := usecase.UpdateTaskInput{
		Title:     request.Body.Title,
		Completed: request.Body.Completed,
		Labels:    toLabelIDs(request.Body.Labels),
	}
	output, err := th.tu.UpdateTask(ctx, domain.UserID(userID), domain.TaskID(request.Id), input)
	if err != nil {
		message := err.Error()
		switch errorStatus(err) {
		case http.StatusNotFound:
			return api.UpdateTask404JSONResponse{Error: message}, nil
		case http.StatusInternalServerError:
			return api.UpdateTask500JSONResponse{Error: message}, nil
		default:
			return api.UpdateTask400JSONResponse{Error: message}, nil
		}
	}

	return api.UpdateTask200JSONResponse{Data: toTaskResponse(output)}, nil
}

func (th *TaskHandler) DeleteTask(ctx context.Context, request api.DeleteTaskRequestObject) (api.DeleteTaskResponseObject, error) {
	userID, err := getLoginUserFromContext(ctx)
	if err != nil {
		return api.DeleteTask500JSONResponse{Error: err.Error()}, nil
	}

	if request.Id <= 0 {
		slog.ErrorContext(ctx, "id less than equals 0")
		return api.DeleteTask400JSONResponse{Error: domain.ErrBadRequest.Error()}, nil
	}

	if err := th.tu.DeleteTask(ctx, domain.UserID(userID), domain.TaskID(request.Id)); err != nil {
		message := err.Error()
		switch errorStatus(err) {
		case http.StatusNotFound:
			return api.DeleteTask404JSONResponse{Error: message}, nil
		case http.StatusInternalServerError:
			return api.DeleteTask500JSONResponse{Error: message}, nil
		default:
			return api.DeleteTask400JSONResponse{Error: message}, nil
		}
	}

	return api.DeleteTask204Response{}, nil
}

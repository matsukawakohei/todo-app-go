package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/matsukawakohei/todo-app-go/domain"
	"github.com/matsukawakohei/todo-app-go/usecase"
)

type mockTaskUsecase struct {
	getAllTasksByUserIDFn func(ctx context.Context, userID domain.UserID) (*usecase.GetAllTasksOutput, error)
	createTaskFn          func(ctx context.Context, userID domain.UserID, input usecase.CreateTaskInput) (*usecase.Task, error)
	updateTaskFn          func(ctx context.Context, userID domain.UserID, taskID domain.TaskID, input usecase.UpdateTaskInput) (*usecase.Task, error)
	deleteTaskFn          func(ctx context.Context, userID domain.UserID, taskID domain.TaskID) error
}

func (m *mockTaskUsecase) GetAllTasksByUserID(ctx context.Context, userID domain.UserID) (*usecase.GetAllTasksOutput, error) {
	return m.getAllTasksByUserIDFn(ctx, userID)
}

func (m *mockTaskUsecase) CreateTask(ctx context.Context, userID domain.UserID, input usecase.CreateTaskInput) (*usecase.Task, error) {
	return m.createTaskFn(ctx, userID, input)
}

func (m *mockTaskUsecase) UpdateTask(ctx context.Context, userID domain.UserID, taskID domain.TaskID, input usecase.UpdateTaskInput) (*usecase.Task, error) {
	return m.updateTaskFn(ctx, userID, taskID, input)
}

func (m *mockTaskUsecase) DeleteTask(ctx context.Context, userID domain.UserID, taskID domain.TaskID) error {
	return m.deleteTaskFn(ctx, userID, taskID)
}

func TestGetAllTasksByUserID(t *testing.T) {
	tests := []struct {
		testName              string
		loggedIn              bool
		getAllTasksByUserIDFn func(ctx context.Context, userID domain.UserID) (*usecase.GetAllTasksOutput, error)
		wantStatus            int
		wantTasks             []*usecase.Task
	}{
		{
			testName:   "getLoginUserFromContextError",
			loggedIn:   false,
			wantStatus: http.StatusInternalServerError,
		},
		{
			testName: "usecaseErrUnexpected",
			loggedIn: true,
			getAllTasksByUserIDFn: func(_ context.Context, _ domain.UserID) (*usecase.GetAllTasksOutput, error) {
				return nil, domain.ErrUnexpected
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			testName: "empty",
			loggedIn: true,
			getAllTasksByUserIDFn: func(_ context.Context, _ domain.UserID) (*usecase.GetAllTasksOutput, error) {
				return &usecase.GetAllTasksOutput{Tasks: []*usecase.Task{}}, nil
			},
			wantStatus: http.StatusOK,
			wantTasks:  []*usecase.Task{},
		},
		{
			testName: "success",
			loggedIn: true,
			getAllTasksByUserIDFn: func(_ context.Context, _ domain.UserID) (*usecase.GetAllTasksOutput, error) {
				return &usecase.GetAllTasksOutput{Tasks: []*usecase.Task{
					{ID: 1, Title: "task1", Labels: []*usecase.Label{}},
					{ID: 2, Title: "task2", Completed: true, Labels: []*usecase.Label{}},
				}}, nil
			},
			wantStatus: http.StatusOK,
			wantTasks: []*usecase.Task{
				{ID: 1, Title: "task1", Labels: []*usecase.Label{}},
				{ID: 2, Title: "task2", Completed: true, Labels: []*usecase.Label{}},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			c, rec := newTestContext(http.MethodGet, "")
			if tc.loggedIn {
				setLoginUser(c, 1)
			}
			tcon := NewTaskController(&mockTaskUsecase{getAllTasksByUserIDFn: tc.getAllTasksByUserIDFn})
			if err := tcon.GetAllTasksByUserID(c); err != nil {
				t.Fatalf("task controller GetAllTasksByUserID failed: %v", err)
			}

			if rec.Code != tc.wantStatus {
				t.Errorf("expected status %d; got %d", tc.wantStatus, rec.Code)
			}

			if tc.wantTasks != nil {
				var output usecase.GetAllTasksOutput
				if err := json.Unmarshal(rec.Body.Bytes(), &output); err != nil {
					t.Fatalf("failed to unmarshal response body: %v", err)
				}
				if len(output.Tasks) != len(tc.wantTasks) {
					t.Fatalf("expected %d tasks; got %d", len(tc.wantTasks), len(output.Tasks))
				}
				for i, tsk := range output.Tasks {
					if tsk.ID != tc.wantTasks[i].ID || tsk.Title != tc.wantTasks[i].Title || tsk.Completed != tc.wantTasks[i].Completed {
						t.Errorf("expected task %+v; got %+v", *tc.wantTasks[i], *tsk)
					}
				}
			}
		})
	}
}

func TestCreateTask(t *testing.T) {
	tests := []struct {
		testName   string
		loggedIn   bool
		body       string
		createFn   func(ctx context.Context, userID domain.UserID, input usecase.CreateTaskInput) (*usecase.Task, error)
		wantStatus int
		wantTask   *usecase.Task
	}{
		{
			testName:   "getLoginUserFromContextError",
			loggedIn:   false,
			body:       `{"Title":"task1"}`,
			wantStatus: http.StatusInternalServerError,
		},
		{
			testName:   "bindError",
			loggedIn:   true,
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			testName: "validationError",
			loggedIn: true,
			body:     `{"Title":""}`,
			createFn: func(_ context.Context, _ domain.UserID, _ usecase.CreateTaskInput) (*usecase.Task, error) {
				return nil, errors.New("Title: title is required.")
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			testName: "labelCountMismatch",
			loggedIn: true,
			body:     `{"Title":"task1","Labels":[1]}`,
			createFn: func(_ context.Context, _ domain.UserID, _ usecase.CreateTaskInput) (*usecase.Task, error) {
				return nil, domain.ErrBadRequest
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			testName: "usecaseErrUnexpected",
			loggedIn: true,
			body:     `{"Title":"task1"}`,
			createFn: func(_ context.Context, _ domain.UserID, _ usecase.CreateTaskInput) (*usecase.Task, error) {
				return nil, domain.ErrUnexpected
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			testName: "success",
			loggedIn: true,
			body:     `{"Title":"task1"}`,
			createFn: func(_ context.Context, _ domain.UserID, _ usecase.CreateTaskInput) (*usecase.Task, error) {
				return &usecase.Task{ID: 1, Title: "task1", Labels: []*usecase.Label{}}, nil
			},
			wantStatus: http.StatusCreated,
			wantTask:   &usecase.Task{ID: 1, Title: "task1", Labels: []*usecase.Label{}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			c, rec := newTestContext(http.MethodPost, tc.body)
			if tc.loggedIn {
				setLoginUser(c, 1)
			}
			tcon := NewTaskController(&mockTaskUsecase{createTaskFn: tc.createFn})
			if err := tcon.CreateTask(c); err != nil {
				t.Fatalf("task controller CreateTask failed: %v", err)
			}

			if rec.Code != tc.wantStatus {
				t.Errorf("expected status %d; got %d", tc.wantStatus, rec.Code)
			}

			if tc.wantTask != nil {
				var output usecase.Task
				if err := json.Unmarshal(rec.Body.Bytes(), &output); err != nil {
					t.Fatalf("failed to unmarshal response body: %v", err)
				}
				if output.ID != tc.wantTask.ID || output.Title != tc.wantTask.Title || output.Completed != tc.wantTask.Completed {
					t.Errorf("expected task %+v; got %+v", *tc.wantTask, output)
				}
			}
		})
	}
}

func TestUpdateTask(t *testing.T) {
	tests := []struct {
		testName   string
		loggedIn   bool
		taskIDStr  string
		body       string
		updateFn   func(ctx context.Context, userID domain.UserID, taskID domain.TaskID, input usecase.UpdateTaskInput) (*usecase.Task, error)
		wantStatus int
		wantTask   *usecase.Task
	}{
		{
			testName:   "getLoginUserFromContextError",
			loggedIn:   false,
			taskIDStr:  "1",
			body:       `{"Title":"task1","Completed":false}`,
			wantStatus: http.StatusInternalServerError,
		},
		{
			testName:   "getTaskIDFromParamError",
			loggedIn:   true,
			taskIDStr:  "abc",
			body:       `{"Title":"task1","Completed":false}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			testName:   "bindError",
			loggedIn:   true,
			taskIDStr:  "1",
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			testName:  "validationError",
			loggedIn:  true,
			taskIDStr: "1",
			body:      `{"Title":"","Completed":false}`,
			updateFn: func(_ context.Context, _ domain.UserID, _ domain.TaskID, _ usecase.UpdateTaskInput) (*usecase.Task, error) {
				return nil, errors.New("Title: title is required.")
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			testName:  "notFound",
			loggedIn:  true,
			taskIDStr: "1",
			body:      `{"Title":"task1","Completed":false}`,
			updateFn: func(_ context.Context, _ domain.UserID, _ domain.TaskID, _ usecase.UpdateTaskInput) (*usecase.Task, error) {
				return nil, domain.ErrNotFound
			},
			wantStatus: http.StatusNotFound,
		},
		{
			testName:  "usecaseErrUnexpected",
			loggedIn:  true,
			taskIDStr: "1",
			body:      `{"Title":"task1","Completed":false}`,
			updateFn: func(_ context.Context, _ domain.UserID, _ domain.TaskID, _ usecase.UpdateTaskInput) (*usecase.Task, error) {
				return nil, domain.ErrUnexpected
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			testName:  "success",
			loggedIn:  true,
			taskIDStr: "1",
			body:      `{"Title":"task1","Completed":true}`,
			updateFn: func(_ context.Context, _ domain.UserID, _ domain.TaskID, _ usecase.UpdateTaskInput) (*usecase.Task, error) {
				return &usecase.Task{ID: 1, Title: "task1", Completed: true, Labels: []*usecase.Label{}}, nil
			},
			wantStatus: http.StatusOK,
			wantTask:   &usecase.Task{ID: 1, Title: "task1", Completed: true, Labels: []*usecase.Label{}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			c, rec := newTestContext(http.MethodPut, tc.body)
			if tc.loggedIn {
				setLoginUser(c, 1)
			}
			c.SetPathValues(echo.PathValues{{Name: "id", Value: tc.taskIDStr}})

			tcon := NewTaskController(&mockTaskUsecase{updateTaskFn: tc.updateFn})
			if err := tcon.UpdateTask(c); err != nil {
				t.Fatalf("task controller UpdateTask failed: %v", err)
			}

			if rec.Code != tc.wantStatus {
				t.Errorf("expected status %d; got %d", tc.wantStatus, rec.Code)
			}

			if tc.wantTask != nil {
				var output usecase.Task
				if err := json.Unmarshal(rec.Body.Bytes(), &output); err != nil {
					t.Fatalf("failed to unmarshal response body: %v", err)
				}
				if output.ID != tc.wantTask.ID || output.Title != tc.wantTask.Title || output.Completed != tc.wantTask.Completed {
					t.Errorf("expected task %+v; got %+v", *tc.wantTask, output)
				}
			}
		})
	}
}

func TestDeleteTask(t *testing.T) {
	tests := []struct {
		testName   string
		loggedIn   bool
		taskIDStr  string
		deleteFn   func(ctx context.Context, userID domain.UserID, taskID domain.TaskID) error
		wantStatus int
	}{
		{
			testName:   "getLoginUserFromContextError",
			loggedIn:   false,
			taskIDStr:  "1",
			wantStatus: http.StatusInternalServerError,
		},
		{
			testName:   "getTaskIDFromParamError",
			loggedIn:   true,
			taskIDStr:  "abc",
			wantStatus: http.StatusBadRequest,
		},
		{
			testName:   "negativeTaskID",
			loggedIn:   true,
			taskIDStr:  "-1",
			wantStatus: http.StatusBadRequest,
		},
		{
			testName:  "notFound",
			loggedIn:  true,
			taskIDStr: "1",
			deleteFn: func(_ context.Context, _ domain.UserID, _ domain.TaskID) error {
				return domain.ErrNotFound
			},
			wantStatus: http.StatusNotFound,
		},
		{
			testName:  "usecaseErrUnexpected",
			loggedIn:  true,
			taskIDStr: "1",
			deleteFn: func(_ context.Context, _ domain.UserID, _ domain.TaskID) error {
				return domain.ErrUnexpected
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			testName:  "success",
			loggedIn:  true,
			taskIDStr: "1",
			deleteFn: func(_ context.Context, _ domain.UserID, _ domain.TaskID) error {
				return nil
			},
			wantStatus: http.StatusNoContent,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			c, rec := newTestContext(http.MethodDelete, "")
			if tc.loggedIn {
				setLoginUser(c, 1)
			}
			c.SetPathValues(echo.PathValues{{Name: "id", Value: tc.taskIDStr}})

			tcon := NewTaskController(&mockTaskUsecase{deleteTaskFn: tc.deleteFn})
			if err := tcon.DeleteTask(c); err != nil {
				t.Fatalf("task controller DeleteTask failed: %v", err)
			}

			if rec.Code != tc.wantStatus {
				t.Errorf("expected status %d; got %d", tc.wantStatus, rec.Code)
			}
		})
	}
}

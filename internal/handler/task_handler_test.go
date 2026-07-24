package handler

import (
	"context"
	"net/http"
	"testing"

	"github.com/matsukawakohei/todo-app-go/domain"
	"github.com/matsukawakohei/todo-app-go/internal/api"
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
		wantTasks             []api.Task
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
			wantTasks:  []api.Task{},
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
			wantTasks: []api.Task{
				{Id: 1, Title: "task1", Labels: []api.Label{}},
				{Id: 2, Title: "task2", Completed: true, Labels: []api.Label{}},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.Background()
			if tc.loggedIn {
				ctx = contextWithUser(1)
			}
			th := NewTaskHandler(&mockTaskUsecase{getAllTasksByUserIDFn: tc.getAllTasksByUserIDFn})
			response, err := th.GetTasks(ctx, api.GetTasksRequestObject{})
			if err != nil {
				t.Fatalf("task handler GetTasks failed: %v", err)
			}

			switch tc.wantStatus {
			case http.StatusOK:
				got, ok := response.(api.GetTasks200JSONResponse)
				if !ok {
					t.Fatalf("expected api.GetTasks200JSONResponse; got %T", response)
				}
				if len(got.Data) != len(tc.wantTasks) {
					t.Fatalf("expected %d tasks; got %d", len(tc.wantTasks), len(got.Data))
				}
				for i, tsk := range got.Data {
					if tsk.Id != tc.wantTasks[i].Id || tsk.Title != tc.wantTasks[i].Title || tsk.Completed != tc.wantTasks[i].Completed {
						t.Errorf("expected task %+v; got %+v", tc.wantTasks[i], tsk)
					}
				}
			case http.StatusInternalServerError:
				if _, ok := response.(api.GetTasks500JSONResponse); !ok {
					t.Fatalf("expected api.GetTasks500JSONResponse; got %T", response)
				}
			}
		})
	}
}

func TestCreateTask(t *testing.T) {
	tests := []struct {
		testName   string
		loggedIn   bool
		title      string
		createFn   func(ctx context.Context, userID domain.UserID, input usecase.CreateTaskInput) (*usecase.Task, error)
		wantStatus int
		wantTask   *api.Task
	}{
		{
			testName:   "getLoginUserFromContextError",
			loggedIn:   false,
			title:      "task1",
			wantStatus: http.StatusInternalServerError,
		},
		{
			testName: "validationError",
			loggedIn: true,
			title:    "",
			createFn: func(_ context.Context, _ domain.UserID, _ usecase.CreateTaskInput) (*usecase.Task, error) {
				return nil, domain.ErrBadRequest
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			testName: "labelCountMismatch",
			loggedIn: true,
			title:    "task1",
			createFn: func(_ context.Context, _ domain.UserID, _ usecase.CreateTaskInput) (*usecase.Task, error) {
				return nil, domain.ErrBadRequest
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			testName: "usecaseErrUnexpected",
			loggedIn: true,
			title:    "task1",
			createFn: func(_ context.Context, _ domain.UserID, _ usecase.CreateTaskInput) (*usecase.Task, error) {
				return nil, domain.ErrUnexpected
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			testName: "success",
			loggedIn: true,
			title:    "task1",
			createFn: func(_ context.Context, _ domain.UserID, _ usecase.CreateTaskInput) (*usecase.Task, error) {
				return &usecase.Task{ID: 1, Title: "task1", Labels: []*usecase.Label{}}, nil
			},
			wantStatus: http.StatusCreated,
			wantTask:   &api.Task{Id: 1, Title: "task1", Labels: []api.Label{}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.Background()
			if tc.loggedIn {
				ctx = contextWithUser(1)
			}
			th := NewTaskHandler(&mockTaskUsecase{createTaskFn: tc.createFn})
			request := api.CreateTaskRequestObject{Body: &api.TaskCreateInput{Title: tc.title, Labels: []int{}}}
			response, err := th.CreateTask(ctx, request)
			if err != nil {
				t.Fatalf("task handler CreateTask failed: %v", err)
			}

			switch tc.wantStatus {
			case http.StatusCreated:
				got, ok := response.(api.CreateTask201JSONResponse)
				if !ok {
					t.Fatalf("expected api.CreateTask201JSONResponse; got %T", response)
				}
				if got.Data.Id != tc.wantTask.Id || got.Data.Title != tc.wantTask.Title || got.Data.Completed != tc.wantTask.Completed {
					t.Errorf("expected task %+v; got %+v", *tc.wantTask, got.Data)
				}
			case http.StatusBadRequest:
				if _, ok := response.(api.CreateTask400JSONResponse); !ok {
					t.Fatalf("expected api.CreateTask400JSONResponse; got %T", response)
				}
			case http.StatusInternalServerError:
				if _, ok := response.(api.CreateTask500JSONResponse); !ok {
					t.Fatalf("expected api.CreateTask500JSONResponse; got %T", response)
				}
			}
		})
	}
}

func TestUpdateTask(t *testing.T) {
	tests := []struct {
		testName   string
		loggedIn   bool
		taskID     int
		title      string
		completed  bool
		updateFn   func(ctx context.Context, userID domain.UserID, taskID domain.TaskID, input usecase.UpdateTaskInput) (*usecase.Task, error)
		wantStatus int
		wantTask   *api.Task
	}{
		{
			testName:   "getLoginUserFromContextError",
			loggedIn:   false,
			taskID:     1,
			title:      "task1",
			wantStatus: http.StatusInternalServerError,
		},
		{
			testName:   "negativeTaskID",
			loggedIn:   true,
			taskID:     -1,
			title:      "task1",
			wantStatus: http.StatusBadRequest,
		},
		{
			testName: "validationError",
			loggedIn: true,
			taskID:   1,
			title:    "",
			updateFn: func(_ context.Context, _ domain.UserID, _ domain.TaskID, _ usecase.UpdateTaskInput) (*usecase.Task, error) {
				return nil, domain.ErrBadRequest
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			testName: "notFound",
			loggedIn: true,
			taskID:   1,
			title:    "task1",
			updateFn: func(_ context.Context, _ domain.UserID, _ domain.TaskID, _ usecase.UpdateTaskInput) (*usecase.Task, error) {
				return nil, domain.ErrNotFound
			},
			wantStatus: http.StatusNotFound,
		},
		{
			testName: "usecaseErrUnexpected",
			loggedIn: true,
			taskID:   1,
			title:    "task1",
			updateFn: func(_ context.Context, _ domain.UserID, _ domain.TaskID, _ usecase.UpdateTaskInput) (*usecase.Task, error) {
				return nil, domain.ErrUnexpected
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			testName:  "success",
			loggedIn:  true,
			taskID:    1,
			title:     "task1",
			completed: true,
			updateFn: func(_ context.Context, _ domain.UserID, _ domain.TaskID, _ usecase.UpdateTaskInput) (*usecase.Task, error) {
				return &usecase.Task{ID: 1, Title: "task1", Completed: true, Labels: []*usecase.Label{}}, nil
			},
			wantStatus: http.StatusOK,
			wantTask:   &api.Task{Id: 1, Title: "task1", Completed: true, Labels: []api.Label{}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.Background()
			if tc.loggedIn {
				ctx = contextWithUser(1)
			}

			th := NewTaskHandler(&mockTaskUsecase{updateTaskFn: tc.updateFn})
			request := api.UpdateTaskRequestObject{
				Id:   tc.taskID,
				Body: &api.TaskUpdateInput{Title: tc.title, Completed: tc.completed, Labels: []int{}},
			}
			response, err := th.UpdateTask(ctx, request)
			if err != nil {
				t.Fatalf("task handler UpdateTask failed: %v", err)
			}

			switch tc.wantStatus {
			case http.StatusOK:
				got, ok := response.(api.UpdateTask200JSONResponse)
				if !ok {
					t.Fatalf("expected api.UpdateTask200JSONResponse; got %T", response)
				}
				if got.Data.Id != tc.wantTask.Id || got.Data.Title != tc.wantTask.Title || got.Data.Completed != tc.wantTask.Completed {
					t.Errorf("expected task %+v; got %+v", *tc.wantTask, got.Data)
				}
			case http.StatusBadRequest:
				if _, ok := response.(api.UpdateTask400JSONResponse); !ok {
					t.Fatalf("expected api.UpdateTask400JSONResponse; got %T", response)
				}
			case http.StatusNotFound:
				if _, ok := response.(api.UpdateTask404JSONResponse); !ok {
					t.Fatalf("expected api.UpdateTask404JSONResponse; got %T", response)
				}
			case http.StatusInternalServerError:
				if _, ok := response.(api.UpdateTask500JSONResponse); !ok {
					t.Fatalf("expected api.UpdateTask500JSONResponse; got %T", response)
				}
			}
		})
	}
}

func TestDeleteTask(t *testing.T) {
	tests := []struct {
		testName   string
		loggedIn   bool
		taskID     int
		deleteFn   func(ctx context.Context, userID domain.UserID, taskID domain.TaskID) error
		wantStatus int
	}{
		{
			testName:   "getLoginUserFromContextError",
			loggedIn:   false,
			taskID:     1,
			wantStatus: http.StatusInternalServerError,
		},
		{
			testName:   "negativeTaskID",
			loggedIn:   true,
			taskID:     -1,
			wantStatus: http.StatusBadRequest,
		},
		{
			testName: "notFound",
			loggedIn: true,
			taskID:   1,
			deleteFn: func(_ context.Context, _ domain.UserID, _ domain.TaskID) error {
				return domain.ErrNotFound
			},
			wantStatus: http.StatusNotFound,
		},
		{
			testName: "usecaseErrUnexpected",
			loggedIn: true,
			taskID:   1,
			deleteFn: func(_ context.Context, _ domain.UserID, _ domain.TaskID) error {
				return domain.ErrUnexpected
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			testName: "success",
			loggedIn: true,
			taskID:   1,
			deleteFn: func(_ context.Context, _ domain.UserID, _ domain.TaskID) error {
				return nil
			},
			wantStatus: http.StatusNoContent,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.Background()
			if tc.loggedIn {
				ctx = contextWithUser(1)
			}

			th := NewTaskHandler(&mockTaskUsecase{deleteTaskFn: tc.deleteFn})
			response, err := th.DeleteTask(ctx, api.DeleteTaskRequestObject{Id: tc.taskID})
			if err != nil {
				t.Fatalf("task handler DeleteTask failed: %v", err)
			}

			switch tc.wantStatus {
			case http.StatusNoContent:
				if _, ok := response.(api.DeleteTask204Response); !ok {
					t.Fatalf("expected api.DeleteTask204Response; got %T", response)
				}
			case http.StatusBadRequest:
				if _, ok := response.(api.DeleteTask400JSONResponse); !ok {
					t.Fatalf("expected api.DeleteTask400JSONResponse; got %T", response)
				}
			case http.StatusNotFound:
				if _, ok := response.(api.DeleteTask404JSONResponse); !ok {
					t.Fatalf("expected api.DeleteTask404JSONResponse; got %T", response)
				}
			case http.StatusInternalServerError:
				if _, ok := response.(api.DeleteTask500JSONResponse); !ok {
					t.Fatalf("expected api.DeleteTask500JSONResponse; got %T", response)
				}
			}
		})
	}
}

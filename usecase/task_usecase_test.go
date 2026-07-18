package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/matsukawakohei/todo-app-go/domain"
)

type MockTaskRepository struct {
	getAllTasksByUserIDFn func(ctx context.Context, userID domain.UserID) ([]*domain.Task, error)
	createTaskFn          func(ctx context.Context, userID domain.UserID, input domain.CreateTask) (*domain.Task, error)
	updateTaskFn          func(ctx context.Context, userID domain.UserID, taskID domain.TaskID, input domain.UpdateTask) (*domain.Task, error)
	deleteTaskFn          func(ctx context.Context, userID domain.UserID, taskID domain.TaskID) error
}

func (m *MockTaskRepository) GetAllTasksByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Task, error) {
	return m.getAllTasksByUserIDFn(ctx, userID)
}

func (m *MockTaskRepository) CreateTask(ctx context.Context, userID domain.UserID, task domain.CreateTask) (*domain.Task, error) {
	return m.createTaskFn(ctx, userID, task)
}

func (m *MockTaskRepository) UpdateTask(ctx context.Context, userID domain.UserID, taskID domain.TaskID, task domain.UpdateTask) (*domain.Task, error) {
	return m.updateTaskFn(ctx, userID, taskID, task)
}

func (m *MockTaskRepository) DeleteTask(ctx context.Context, userID domain.UserID, taskID domain.TaskID) error {
	return m.deleteTaskFn(ctx, userID, taskID)
}

func TestGetAllTasksByUserID(t *testing.T) {
	tests := []struct {
		testName    string
		userID      domain.UserID
		repoTasks   []*domain.Task
		repoErr     error
		wantTaskLen int
		wantTasks   []*Task
		wantErr     bool
		wantErrIs   error
	}{
		{
			testName: "success",
			userID:   1,
			repoTasks: []*domain.Task{
				{ID: 1, Title: "test task 1", Completed: false, Labels: []*domain.Label{}},
				{ID: 2, Title: "test task 2", Completed: true, Labels: []*domain.Label{
					{ID: 1, Name: "test label"},
				}},
			},
			wantTaskLen: 2,
			wantTasks: []*Task{
				{ID: 1, Title: "test task 1", Completed: false, Labels: []*Label{}},
				{ID: 2, Title: "test task 2", Completed: true, Labels: []*Label{
					{ID: 1, Name: "test label"},
				}},
			},
		},
		{
			testName:  "internalServerError",
			userID:    1,
			repoTasks: []*domain.Task{},
			repoErr:   domain.ErrUnexpected,
			wantTasks: []*Task{},
			wantErr:   true,
			wantErrIs: domain.ErrUnexpected,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			repository := MockTaskRepository{
				getAllTasksByUserIDFn: func(ctx context.Context, userID domain.UserID) ([]*domain.Task, error) {
					assertRequestID(t, ctx)
					return tc.repoTasks, tc.repoErr
				},
			}

			validator := NewTaskValidator()
			usecase := NewTaskUsecase(&repository, validator)

			output, err := usecase.GetAllTasksByUserID(makeRequestIDContext(), tc.userID)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("want err; but err is nil")
				}
				if !errors.Is(err, tc.wantErrIs) {
					t.Errorf("want err %v; but got %v", tc.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("err want nil; but got %v", err)
			}
			if len(output.Tasks) != tc.wantTaskLen {
				t.Fatalf("task len want %v; but got %v", tc.wantTaskLen, len(output.Tasks))
			}
			for i, got := range output.Tasks {
				want := tc.wantTasks[i]
				compareTask(t, got, want)
			}
		})
	}
}

func TestCreateTask(t *testing.T) {
	tests := []struct {
		testName  string
		userID    domain.UserID
		input     CreateTaskInput
		repoTask  *domain.Task
		repoErr   error
		wantTask  *Task
		wantErr   bool
		wantErrIs error
	}{
		{
			testName: "success",
			userID:   1,
			input:    CreateTaskInput{Title: "test task", Labels: []uint{1}},
			repoTask: &domain.Task{ID: 1, Title: "test task", Completed: false, Labels: []*domain.Label{
				{ID: 1, Name: "test label"},
			}},
			wantTask: &Task{ID: 1, Title: "test task", Completed: false, Labels: []*Label{
				{ID: 1, Name: "test label"},
			}},
		},
		{
			testName:  "internalServerError",
			userID:    1,
			input:     CreateTaskInput{Title: "test task", Labels: []uint{}},
			repoErr:   domain.ErrUnexpected,
			wantErr:   true,
			wantErrIs: domain.ErrUnexpected,
		},
		{
			testName:  "labelCountMismatch",
			userID:    1,
			input:     CreateTaskInput{Title: "test task", Labels: []uint{999}},
			repoErr:   domain.ErrBadRequest,
			wantErr:   true,
			wantErrIs: domain.ErrBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			repository := MockTaskRepository{
				createTaskFn: func(ctx context.Context, userID domain.UserID, input domain.CreateTask) (*domain.Task, error) {
					assertRequestID(t, ctx)
					return tc.repoTask, tc.repoErr
				},
			}

			validator := NewTaskValidator()
			usecase := NewTaskUsecase(&repository, validator)

			got, err := usecase.CreateTask(makeRequestIDContext(), tc.userID, tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("want err; but err is nil")
				}
				if !errors.Is(err, tc.wantErrIs) {
					t.Errorf("want err %v; but got %v", tc.wantErrIs, err)
				}

				return
			}
			if err != nil {
				t.Fatalf("want err is nil; but got %v", err)
			}
			compareTask(t, got, tc.wantTask)
		})
	}
}

func TestUpdateTask(t *testing.T) {
	tests := []struct {
		testName  string
		userID    domain.UserID
		taskID    domain.TaskID
		input     UpdateTaskInput
		repoTask  *domain.Task
		repoErr   error
		wantTask  *Task
		wantErr   bool
		wantErrIs error
	}{
		{
			testName: "success",
			userID:   1,
			taskID:   1,
			input:    UpdateTaskInput{Title: "updated task", Completed: true, Labels: []uint{1}},
			repoTask: &domain.Task{ID: 1, Title: "updated task", Completed: true, Labels: []*domain.Label{
				{ID: 1, Name: "test label"},
			}},
			wantTask: &Task{ID: 1, Title: "updated task", Completed: true, Labels: []*Label{
				{ID: 1, Name: "test label"},
			}},
		},
		{
			testName:  "internalServerError",
			userID:    1,
			taskID:    1,
			input:     UpdateTaskInput{Title: "updated task", Completed: true, Labels: []uint{}},
			repoErr:   domain.ErrUnexpected,
			wantErr:   true,
			wantErrIs: domain.ErrUnexpected,
		},
		{
			testName:  "taskNotFound",
			userID:    1,
			taskID:    999,
			input:     UpdateTaskInput{Title: "updated task", Completed: true, Labels: []uint{}},
			repoErr:   domain.ErrNotFound,
			wantErr:   true,
			wantErrIs: domain.ErrNotFound,
		},
		{
			testName:  "labelCountMismatch",
			userID:    1,
			taskID:    1,
			input:     UpdateTaskInput{Title: "updated task", Completed: true, Labels: []uint{999}},
			repoErr:   domain.ErrBadRequest,
			wantErr:   true,
			wantErrIs: domain.ErrBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			repository := MockTaskRepository{
				updateTaskFn: func(ctx context.Context, userID domain.UserID, taskID domain.TaskID, input domain.UpdateTask) (*domain.Task, error) {
					assertRequestID(t, ctx)
					return tc.repoTask, tc.repoErr
				},
			}

			validator := NewTaskValidator()
			usecase := NewTaskUsecase(&repository, validator)

			got, err := usecase.UpdateTask(makeRequestIDContext(), tc.userID, tc.taskID, tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("want err; but err is nil")
				}
				if !errors.Is(err, tc.wantErrIs) {
					t.Errorf("want err %v; but got %v", tc.wantErrIs, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("want err is nil; but got %v", err)
			}
			compareTask(t, got, tc.wantTask)
		})
	}
}

func TestDeleteTask(t *testing.T) {
	tests := []struct {
		testName  string
		userID    domain.UserID
		taskID    domain.TaskID
		repoErr   error
		wantErr   bool
		wantErrIs error
	}{
		{
			testName: "success",
			userID:   1,
			taskID:   1,
		},
		{
			testName:  "internalServerError",
			userID:    1,
			taskID:    1,
			repoErr:   domain.ErrUnexpected,
			wantErr:   true,
			wantErrIs: domain.ErrUnexpected,
		},
		{
			testName:  "taskNotFound",
			userID:    1,
			taskID:    999,
			repoErr:   domain.ErrNotFound,
			wantErr:   true,
			wantErrIs: domain.ErrNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			repository := MockTaskRepository{
				deleteTaskFn: func(ctx context.Context, userID domain.UserID, taskID domain.TaskID) error {
					assertRequestID(t, ctx)
					return tc.repoErr
				},
			}

			validator := NewTaskValidator()
			usecase := NewTaskUsecase(&repository, validator)

			err := usecase.DeleteTask(makeRequestIDContext(), tc.userID, tc.taskID)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("want err; but err is nil")
				}
				if !errors.Is(err, tc.wantErrIs) {
					t.Errorf("want err %v; but got %v", tc.wantErrIs, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("want err is nil; but got %v", err)
			}
		})
	}
}

func compareTask(t *testing.T, got, want *Task) {
	t.Helper()

	if got == nil {
		t.Fatalf("got is not nil; but got nil")
	}
	if want == nil {
		t.Fatalf("want is not nil; but got nil")
	}

	if got.ID != want.ID {
		t.Errorf("task id want %v; but got %v", want.ID, got.ID)
	}
	if got.Title != want.Title {
		t.Errorf("task title want %v; but got %v", want.Title, got.Title)
	}
	if got.Completed != want.Completed {
		t.Errorf("task completed want %v, but got %v", want.Completed, got.Completed)
	}
	if len(got.Labels) != len(want.Labels) {
		t.Errorf("task labels len want %v, but got %v", len(want.Labels), len(got.Labels))
	}

	for i, gotLabel := range got.Labels {
		wantLabel := want.Labels[i]
		if gotLabel.ID != wantLabel.ID {
			t.Errorf("task label id want %v; but got %v", wantLabel.ID, gotLabel.ID)
		}
		if gotLabel.Name != wantLabel.Name {
			t.Errorf("task label name want %v; but got %v", wantLabel.Name, gotLabel.Name)
		}
	}
}

package repository

import (
	"errors"
	"testing"

	"github.com/matsukawakohei/todo-app-go/domain"
	"gorm.io/gorm"
)

func TestGetAllTaskByUserID(t *testing.T) {
	tests := []struct {
		testName       string
		seed           bool
		seedUsers      []userModel
		seedLabels     []labelModel
		seedTasks      []taskModel
		seedTaskLabels []taskLabelModel
		userID         domain.UserID
		wantTasks      []*domain.Task
	}{
		{
			testName:  "NotingData",
			seed:      false,
			userID:    domain.UserID(1),
			wantTasks: []*domain.Task{},
		},
		{
			testName: "OtherUserTaskExists",
			seed:     true,
			seedUsers: []userModel{
				{ID: 1, Name: "テスト", Email: "test@example.com"},
			},
			seedLabels: []labelModel{},
			seedTasks: []taskModel{
				{ID: 1, Title: "テスト1", UserID: 1},
			},
			seedTaskLabels: []taskLabelModel{},
			userID:         domain.UserID(2),
			wantTasks:      []*domain.Task{},
		},
		{
			testName: "MyTasksHasNotLabel",
			seed:     true,
			seedUsers: []userModel{
				{ID: 1, Name: "テスト", Email: "test@example.com"},
			},
			seedLabels: []labelModel{},
			seedTasks: []taskModel{
				{ID: 1, Title: "テスト1", UserID: 1},
			},
			seedTaskLabels: []taskLabelModel{},
			userID:         domain.UserID(1),
			wantTasks: []*domain.Task{
				{ID: domain.TaskID(1), Title: "テスト1", Completed: false, Labels: []*domain.Label{}},
			},
		},
		{
			testName: "MyTasksHasLabel",
			seed:     true,
			seedUsers: []userModel{
				{ID: 1, Name: "テスト", Email: "test@example.com"},
			},
			seedLabels: []labelModel{
				{ID: 1, Name: "ラベル", UserID: 1},
				{ID: 2, Name: "ラベル2", UserID: 1},
			},
			seedTasks: []taskModel{
				{ID: 1, Title: "テスト1", UserID: 1},
			},
			seedTaskLabels: []taskLabelModel{
				{UserID: 1, TaskID: 1, LabelID: 1},
				{UserID: 1, TaskID: 1, LabelID: 2},
			},
			userID: domain.UserID(1),
			wantTasks: []*domain.Task{
				{ID: domain.TaskID(1), Title: "テスト1", Completed: false, Labels: []*domain.Label{
					{ID: domain.LabelID(1), Name: "ラベル"},
					{ID: domain.LabelID(2), Name: "ラベル2"},
				}},
			},
		},
		{
			testName: "TaskSortByCompleted",
			seed:     true,
			seedUsers: []userModel{
				{ID: 1, Name: "テスト", Email: "test@example.com"},
			},
			seedLabels: []labelModel{},
			seedTasks: []taskModel{
				{ID: 1, Title: "テスト1", UserID: 1, Completed: true},
				{ID: 2, Title: "テスト2", UserID: 1, Completed: false},
				{ID: 3, Title: "テスト3", UserID: 1, Completed: true},
				{ID: 4, Title: "テスト4", UserID: 1, Completed: false},
			},
			seedTaskLabels: []taskLabelModel{},
			userID:         domain.UserID(1),
			wantTasks: []*domain.Task{
				{ID: domain.TaskID(4), Title: "テスト4", Completed: false, Labels: []*domain.Label{}},
				{ID: domain.TaskID(2), Title: "テスト2", Completed: false, Labels: []*domain.Label{}},
				{ID: domain.TaskID(3), Title: "テスト3", Completed: true, Labels: []*domain.Label{}},
				{ID: domain.TaskID(1), Title: "テスト1", Completed: true, Labels: []*domain.Label{}},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			runInTx(t, func(tx *gorm.DB) {
				if tc.seed {
					execSeed(t, tx, tc.seedUsers, tc.seedLabels, tc.seedTasks, tc.seedTaskLabels)
				}

				repo := NewTaskRepository(tx)
				got, err := repo.GetAllTasksByUserID(t.Context(), tc.userID)
				if err != nil {
					t.Fatalf("don't want err; but got err %v", err)
				}

				if len(tc.wantTasks) != len(got) {
					t.Fatalf("want task length %v; but got length %v", len(tc.wantTasks), len(got))
				}

				if len(tc.wantTasks) == 0 {
					return
				}

				for i, wantTask := range tc.wantTasks {
					gotTask := got[i]
					compareTask(t, wantTask, gotTask, true)
				}
			})
		})
	}
}

func TestCreateTask(t *testing.T) {
	tests := []struct {
		testName   string
		seedUsers  []userModel
		seedLabels []labelModel
		title      string
		userID     uint
		labelIDs   []uint
		wantErr    bool
		wantErrIs  error
		wantTask   *domain.Task
	}{
		{
			testName: "successCreateTaskEmptyLabel",
			seedUsers: []userModel{
				{ID: 1, Name: "テスト", Email: "test@example.com"},
			},
			seedLabels: []labelModel{},
			title:      "テスト1",
			userID:     1,
			labelIDs:   []uint{},
			wantErr:    false,
			wantTask: &domain.Task{
				ID: 1, Title: "テスト1", Completed: false, Labels: []*domain.Label{},
			},
		},
		{
			testName: "successCreateTaskSingleLabel",
			seedUsers: []userModel{
				{ID: 1, Name: "テスト", Email: "test@example.com"},
			},
			seedLabels: []labelModel{
				{ID: 1, Name: "テスト1", UserID: 1},
			},
			title:    "テスト2",
			userID:   1,
			labelIDs: []uint{1},
			wantErr:  false,
			wantTask: &domain.Task{
				ID:        1,
				Title:     "テスト2",
				Completed: false,
				Labels: []*domain.Label{
					{ID: 1, Name: "テスト1"},
				},
			},
		},
		{
			testName: "successCreateTaskMultiLabel",
			seedUsers: []userModel{
				{ID: 1, Name: "テスト", Email: "test@example.com"},
			},
			seedLabels: []labelModel{
				{ID: 1, Name: "テスト1", UserID: 1},
				{ID: 2, Name: "テスト2", UserID: 1},
			},
			title:    "テスト3",
			userID:   1,
			labelIDs: []uint{1, 2},
			wantErr:  false,
			wantTask: &domain.Task{
				ID:        1,
				Title:     "テスト3",
				Completed: false,
				Labels: []*domain.Label{
					{ID: 1, Name: "テスト1"},
					{ID: 2, Name: "テスト2"},
				},
			},
		},
		{
			testName: "errorLabelNotExists",
			seedUsers: []userModel{
				{ID: 1, Name: "テスト", Email: "test@example.com"},
			},
			seedLabels: []labelModel{
				{ID: 1, Name: "テスト1", UserID: 1},
				{ID: 2, Name: "テスト2", UserID: 1},
			},
			title:     "テスト4",
			userID:    1,
			labelIDs:  []uint{3},
			wantErr:   true,
			wantErrIs: domain.ErrBadRequest,
		},
		{
			testName: "errorOtherUserLabel",
			seedUsers: []userModel{
				{ID: 1, Name: "テスト", Email: "test@example.com"},
				{ID: 2, Name: "テスト2", Email: "test2@example.com"},
			},
			seedLabels: []labelModel{
				{ID: 1, Name: "テスト1", UserID: 2},
			},
			title:     "テスト5",
			userID:    1,
			labelIDs:  []uint{1},
			wantErr:   true,
			wantErrIs: domain.ErrBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			runInTx(t, func(tx *gorm.DB) {
				execSeed(t, tx, tc.seedUsers, tc.seedLabels, []taskModel{}, []taskLabelModel{})
				repo := NewTaskRepository(tx)
				labels := make([]domain.LabelID, 0, len(tc.seedLabels))
				for _, labelID := range tc.labelIDs {
					labels = append(labels, domain.LabelID(labelID))
				}
				input := domain.CreateTask{
					Title:    tc.title,
					LabelIDs: labels,
				}
				got, err := repo.CreateTask(t.Context(), domain.UserID(tc.userID), input)
				if tc.wantErr {
					if err == nil {
						t.Fatalf("expected err; but got nil")
					}
					if !errors.Is(err, tc.wantErrIs) {
						t.Errorf("expected err %v; but got %v", tc.wantErrIs, err)
					}
					var taskLabelCount int64
					if err := tx.Model(&taskLabelModel{}).Where("user_id = ?", tc.userID).Count(&taskLabelCount).Error; err != nil {
						t.Fatalf("failed get task labels count %v", err)
					}
					if taskLabelCount != 0 {
						t.Errorf("want task labels count 0; but got %v", taskLabelCount)
					}
					var taskCount int64
					if err := tx.Model(&taskModel{}).Where("user_id = ?", tc.userID).Count(&taskCount).Error; err != nil {
						t.Fatalf("failed task count %v", err)
					}
					if taskCount != 0 {
						t.Errorf("want task count 0; but got %v", taskCount)
					}

					return
				}

				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				compareTask(t, tc.wantTask, got, false)

				storedTask := getTaskByID(t, tx, got.ID)
				compareTask(t, tc.wantTask, storedTask, false)
			})
		})
	}
}

func TestUpdateTask(t *testing.T) {
	tests := []struct {
		testName       string
		seedUsers      []userModel
		seedLabels     []labelModel
		seedTasks      []taskModel
		seedTaskLables []taskLabelModel
		userID         domain.UserID
		taskID         domain.TaskID
		labelIDs       []domain.LabelID
		title          string
		completed      bool
		wantErr        bool
		wantErrIs      error
		wantTask       *domain.Task
	}{
		{
			testName:  "taskNotFound",
			userID:    1,
			taskID:    1,
			title:     "テスト",
			completed: false,
			wantErr:   true,
			wantErrIs: domain.ErrNotFound,
		},
		{
			testName: "taskUnmatchUser",
			seedUsers: []userModel{
				{ID: 1, Name: "テスト", Email: "test@example.com"},
			},
			seedLabels: []labelModel{},
			seedTasks: []taskModel{
				{ID: 1, Title: "テスト1", UserID: 1},
			},
			seedTaskLables: []taskLabelModel{},
			userID:         2,
			taskID:         1,
			title:          "テスト",
			completed:      false,
			wantErr:        true,
			wantErrIs:      domain.ErrNotFound,
		},
		{
			testName: "labelNotFound",
			seedUsers: []userModel{
				{ID: 1, Name: "テスト", Email: "test@example.com"},
			},
			seedLabels: []labelModel{
				{ID: 1, Name: "テスト1", UserID: 1},
				{ID: 2, Name: "テスト2", UserID: 1},
			},
			seedTasks: []taskModel{
				{ID: 1, Title: "テスト", Completed: false, UserID: 1},
			},
			userID:    1,
			taskID:    1,
			title:     "テスト",
			completed: false,
			labelIDs:  []domain.LabelID{3},
			wantErr:   true,
			wantErrIs: domain.ErrBadRequest,
		},
		{
			testName: "otherUserLabel",
			seedUsers: []userModel{
				{ID: 1, Name: "テスト", Email: "test@example.com"},
				{ID: 2, Name: "テスト2", Email: "test2@example.com"},
			},
			seedLabels: []labelModel{
				{ID: 1, Name: "テスト1", UserID: 1},
			},
			seedTasks: []taskModel{
				{ID: 1, Title: "テスト", Completed: false, UserID: 2},
			},
			userID:    2,
			taskID:    1,
			title:     "テスト",
			completed: false,
			labelIDs:  []domain.LabelID{1},
			wantErr:   true,
			wantErrIs: domain.ErrBadRequest,
		},
		{
			testName: "updateTitle",
			seedUsers: []userModel{
				{ID: 1, Name: "テスト", Email: "test@example.com"},
			},
			seedLabels: []labelModel{},
			seedTasks: []taskModel{
				{ID: 1, Title: "テスト1", UserID: 1},
			},
			userID:    1,
			taskID:    1,
			title:     "更新",
			completed: false,
			labelIDs:  []domain.LabelID{},
			wantErr:   false,
			wantTask: &domain.Task{
				ID: 1, Title: "更新", Completed: false, Labels: []*domain.Label{},
			},
		},
		{
			testName: "updateCompleted",
			seedUsers: []userModel{
				{ID: 1, Name: "テスト", Email: "test@example.com"},
			},
			seedLabels: []labelModel{},
			seedTasks: []taskModel{
				{ID: 1, Title: "テスト1", UserID: 1, Completed: false},
			},
			userID:    1,
			taskID:    1,
			title:     "テスト1",
			completed: true,
			labelIDs:  []domain.LabelID{},
			wantErr:   false,
			wantTask: &domain.Task{
				ID: 1, Title: "テスト1", Completed: true, Labels: []*domain.Label{},
			},
		},
		{
			testName: "removeLabel",
			seedUsers: []userModel{
				{ID: 1, Name: "テスト", Email: "test@example.com"},
			},
			seedLabels: []labelModel{
				{ID: 1, Name: "ラベル", UserID: 1},
			},
			seedTasks: []taskModel{
				{ID: 1, Title: "テスト1", UserID: 1, Completed: false},
			},
			seedTaskLables: []taskLabelModel{
				{UserID: 1, TaskID: 1, LabelID: 1},
			},
			userID:    1,
			taskID:    1,
			title:     "テスト1",
			completed: false,
			labelIDs:  []domain.LabelID{},
			wantErr:   false,
			wantTask: &domain.Task{
				ID: 1, Title: "テスト1", Completed: false, Labels: []*domain.Label{},
			},
		},
		{
			testName: "appendLabel",
			seedUsers: []userModel{
				{ID: 1, Name: "テスト", Email: "test@example.com"},
			},
			seedLabels: []labelModel{
				{ID: 1, Name: "ラベル1", UserID: 1},
				{ID: 2, Name: "ラベル2", UserID: 1},
			},
			seedTasks: []taskModel{
				{ID: 1, Title: "テスト1", UserID: 1, Completed: false},
			},
			seedTaskLables: []taskLabelModel{
				{UserID: 1, TaskID: 1, LabelID: 1},
			},
			userID:    1,
			taskID:    1,
			title:     "テスト1",
			completed: false,
			labelIDs:  []domain.LabelID{1, 2},
			wantErr:   false,
			wantTask: &domain.Task{
				ID: 1, Title: "テスト1", Completed: false, Labels: []*domain.Label{
					{ID: 1, Name: "ラベル1"},
					{ID: 2, Name: "ラベル2"},
				},
			},
		},
		{
			testName: "sameLabel",
			seedUsers: []userModel{
				{ID: 1, Name: "テスト", Email: "test@example.com"},
			},
			seedLabels: []labelModel{
				{ID: 1, Name: "ラベル1", UserID: 1},
			},
			seedTasks: []taskModel{
				{ID: 1, Title: "テスト1", UserID: 1, Completed: false},
			},
			seedTaskLables: []taskLabelModel{
				{UserID: 1, TaskID: 1, LabelID: 1},
			},
			userID:    1,
			taskID:    1,
			title:     "テスト1",
			completed: false,
			labelIDs:  []domain.LabelID{1},
			wantErr:   false,
			wantTask: &domain.Task{
				ID: 1, Title: "テスト1", Completed: false, Labels: []*domain.Label{
					{ID: 1, Name: "ラベル1"},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			runInTx(t, func(tx *gorm.DB) {
				execSeed(t, tx, tc.seedUsers, tc.seedLabels, tc.seedTasks, tc.seedTaskLables)
				repo := NewTaskRepository(tx)
				input := domain.UpdateTask{
					Title:     tc.title,
					Completed: tc.completed,
					LabelIDs:  tc.labelIDs,
				}
				got, err := repo.UpdateTask(t.Context(), tc.userID, tc.taskID, input)
				if tc.wantErr {
					if err == nil {
						t.Fatalf("want err; but got nil")
					}
					if !errors.Is(err, tc.wantErrIs) {
						t.Errorf("want err %v; but got err %v", tc.wantErrIs, err)
					}
					return
				}
				if err != nil {
					t.Fatalf("want err is nil; but got %v", err)
				}

				compareTask(t, tc.wantTask, got, true)

				storedTask := getTaskByID(t, tx, got.ID)
				compareTask(t, tc.wantTask, storedTask, true)
			})
		})
	}
}

func TestDeleteTask(t *testing.T) {
	tests := []struct {
		testName                string
		seedUsers               []userModel
		seedLabels              []labelModel
		seedTasks               []taskModel
		seedTaskLables          []taskLabelModel
		userID                  domain.UserID
		taskID                  domain.TaskID
		otherTaskID             domain.TaskID
		wantOtherTaskCount      int64
		wantOtherTaskLabelCount int64
		wantErr                 bool
		wantErrIs               error
	}{
		{
			testName:  "taskNotFound",
			userID:    1,
			taskID:    1,
			wantErr:   true,
			wantErrIs: domain.ErrNotFound,
		},
		{
			testName: "taskUnmatchUser",
			seedUsers: []userModel{
				{ID: 1, Name: "テスト", Email: "test@example.com"},
			},
			seedLabels: []labelModel{},
			seedTasks: []taskModel{
				{ID: 1, Title: "テスト1", UserID: 1},
			},
			seedTaskLables: []taskLabelModel{},
			userID:         2,
			taskID:         1,
			wantErr:        true,
			wantErrIs:      domain.ErrNotFound,
		},
		{
			testName: "deleteTaskHasNotLabel",
			seedUsers: []userModel{
				{ID: 1, Name: "テスト", Email: "test@example.com"},
			},
			seedLabels: []labelModel{
				{ID: 1, Name: "ラベル", UserID: 1},
			},
			seedTasks: []taskModel{
				{ID: 1, Title: "テスト1", UserID: 1, Completed: false},
				{ID: 2, Title: "テスト2", UserID: 1, Completed: false},
			},
			seedTaskLables: []taskLabelModel{
				{UserID: 1, TaskID: 2, LabelID: 1},
			},
			userID:                  1,
			taskID:                  1,
			otherTaskID:             2,
			wantOtherTaskCount:      1,
			wantOtherTaskLabelCount: 1,
			wantErr:                 false,
		},
		{
			testName: "deleteTaskHasLabel",
			seedUsers: []userModel{
				{ID: 1, Name: "テスト", Email: "test@example.com"},
			},
			seedLabels: []labelModel{
				{ID: 1, Name: "ラベル", UserID: 1},
			},
			seedTasks: []taskModel{
				{ID: 1, Title: "テスト1", UserID: 1, Completed: false},
				{ID: 2, Title: "テスト2", UserID: 1, Completed: false},
			},
			seedTaskLables: []taskLabelModel{
				{UserID: 1, TaskID: 1, LabelID: 1},
				{UserID: 1, TaskID: 2, LabelID: 1},
			},
			userID:                  1,
			taskID:                  1,
			otherTaskID:             2,
			wantOtherTaskCount:      1,
			wantOtherTaskLabelCount: 1,
			wantErr:                 false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			runInTx(t, func(tx *gorm.DB) {
				execSeed(t, tx, tc.seedUsers, tc.seedLabels, tc.seedTasks, tc.seedTaskLables)
				repo := NewTaskRepository(tx)
				err := repo.DeleteTask(t.Context(), tc.userID, tc.taskID)
				if tc.wantErr {
					if err == nil {
						t.Fatalf("want err; but got nil")
					}
					if !errors.Is(err, tc.wantErrIs) {
						t.Errorf("want err %v; but got %v", tc.wantErrIs, err)
					}

					return
				}

				if err != nil {
					t.Fatalf("want err is nil; but got err %v", err)
				}

				var taskLabelCount int64
				if err := tx.Model(&taskLabelModel{}).Where("user_id = ? and task_id = ?", tc.userID, tc.taskID).Count(&taskLabelCount).Error; err != nil {
					t.Fatalf("failed get task labels count: %v", err)
				}
				if taskLabelCount != 0 {
					t.Errorf("want task labels count 0; but got count %v", taskLabelCount)
				}

				var taskCount int64
				if err := tx.Model(&taskModel{}).Where("id = ?", tc.taskID).Count(&taskCount).Error; err != nil {
					t.Fatalf("failed get task count: %v", err)
				}
				if taskCount != 0 {
					t.Errorf("want task count 0; but got count %v", taskCount)
				}

				var otherTaskLabelCount int64
				if err := tx.Model(&taskLabelModel{}).Where("user_id = ? and task_id = ?", tc.userID, tc.otherTaskID).Count(&otherTaskLabelCount).Error; err != nil {
					t.Fatalf("failed get other task labels count %v", err)
				}
				if otherTaskLabelCount != tc.wantOtherTaskLabelCount {
					t.Errorf("want other task label count %v; but got %v", tc.wantOtherTaskLabelCount, otherTaskLabelCount)
				}

				var otherTaskCount int64
				if err := tx.Model(&taskModel{}).Where("id = ?", tc.otherTaskID).Count(&otherTaskCount).Error; err != nil {
					t.Fatalf("failed get other task count %v", err)
				}
				if otherTaskCount != tc.wantOtherTaskCount {
					t.Errorf("want other task count %v; but got %v", tc.wantOtherTaskCount, otherTaskCount)
				}
			})
		})
	}
}

func compareTask(t *testing.T, want, got *domain.Task, includeID bool) {
	t.Helper()

	if includeID && got.ID != want.ID {
		t.Errorf("task id want %v; but got %v", want.ID, got.ID)
	}
	if got.Title != want.Title {
		t.Errorf("task title want %v; but got %v", want.Title, got.Title)
	}
	if got.Completed != want.Completed {
		t.Errorf("task completed want %v; but got %v", want.Completed, got.Completed)
	}

	if len(got.Labels) != len(want.Labels) {
		t.Errorf("task labels length want %v; but got %v", len(want.Labels), len(got.Labels))
	}
	if len(want.Labels) == 0 {
		return
	}

	gotLabelMap := make(map[domain.LabelID]*domain.Label)
	for _, l := range got.Labels {
		gotLabelMap[l.ID] = l
	}
	for _, wantLabel := range want.Labels {
		gotLabel, exists := gotLabelMap[wantLabel.ID]
		if !exists {
			t.Errorf("task label want exists id %v; but got not exists", wantLabel.ID)
		}
		if gotLabel.Name != wantLabel.Name {
			t.Errorf("want task label name want %v; but got %v", wantLabel.Name, gotLabel.Name)
		}
	}
}

func execSeed(t *testing.T, tx *gorm.DB, seedUsers []userModel, seedLabels []labelModel, seedTasks []taskModel, seedTaskLabels []taskLabelModel) {
	if len(seedUsers) > 0 {
		if err := tx.Create(&seedUsers).Error; err != nil {
			t.Fatalf("failed create seed users %v", err)
		}
	}
	if len(seedLabels) > 0 {
		if err := tx.Create(&seedLabels).Error; err != nil {
			t.Fatalf("failed create seed labels %v", err)
		}
	}
	if len(seedTasks) > 0 {
		if err := tx.Create(&seedTasks).Error; err != nil {
			t.Fatalf("failed create seed tasks %v", err)
		}
	}
	if len(seedTaskLabels) > 0 {
		if err := tx.Create(seedTaskLabels).Error; err != nil {
			t.Fatalf("failed create seed task labels %v", err)
		}
	}
}

func getTaskByID(t *testing.T, tx *gorm.DB, taskId domain.TaskID) *domain.Task {
	var storedModel taskModel
	if err := tx.Preload("Labels").First(&storedModel, taskId).Error; err != nil {
		t.Fatalf("failed get task by id %v", err)
	}

	return storedModel.toEntity()
}

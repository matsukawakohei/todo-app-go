package repository

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/matsukawakohei/todo-app-go/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type taskModel struct {
	ID        uint `gorm:"primaryKey"`
	Title     string
	Completed bool
	UserID    uint
	Labels    []labelModel `gorm:"many2many:task_labels;joinForeignKey:TaskID;joinReferences:LabelID"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (tm *taskModel) toEntity() *domain.Task {
	labels := make([]*domain.Label, 0, len(tm.Labels))
	for _, l := range tm.Labels {
		labels = append(labels, l.toEntity())
	}
	return &domain.Task{
		ID:        domain.TaskID(tm.ID),
		Title:     tm.Title,
		Completed: tm.Completed,
		Labels:    labels,
		CreatedAt: tm.CreatedAt,
		UpdatedAt: tm.UpdatedAt,
	}
}

func (tm *taskModel) TableName() string {
	return "tasks"
}

type taskLabelModel struct {
	UserID  uint `gorm:"primaryKey"`
	TaskID  uint `gorm:"primaryKey"`
	LabelID uint `gorm:"primaryKey"`
}

func (tlm *taskLabelModel) TableName() string {
	return "task_labels"
}

type taskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) domain.TaskRepository {
	return &taskRepository{db}
}

func (tr *taskRepository) GetAllTasksByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Task, error) {
	var models []taskModel
	if err := tr.db.WithContext(ctx).Preload("Labels").Where("user_id = ?", userID).Order("completed, created_at desc, id desc").Find(&models).Error; err != nil {
		slog.ErrorContext(ctx, "GetAllTasksByUserID: failed get tasks by user_id", "error", err)
		return []*domain.Task{}, domain.ErrUnexpected
	}

	tasks := make([]*domain.Task, 0, len(models))
	for _, m := range models {
		tasks = append(tasks, m.toEntity())
	}
	return tasks, nil
}

func (tr *taskRepository) CreateTask(ctx context.Context, userID domain.UserID, task domain.CreateTask) (*domain.Task, error) {
	model := taskModel{
		Title:  task.Title,
		UserID: uint(userID),
	}

	err := tr.db.Transaction(func(tx *gorm.DB) error {
		tx = tx.WithContext(ctx)

		labelCount, err := tr.getLabelCount(ctx, tx, userID, task.LabelIDs)
		if err != nil {
			slog.ErrorContext(ctx, "Create Task: failed getLabelCount")
			return err
		}
		if labelCount != int64(len(task.LabelIDs)) {
			slog.WarnContext(ctx, "Create Task: label count is unmatch")
			return domain.ErrBadRequest
		}

		if err := tx.Create(&model).Error; err != nil {
			slog.ErrorContext(ctx, "Create Task: failed create task", "error", err)
			return domain.ErrUnexpected
		}

		if err := tr.createTaskLabels(ctx, tx, userID, model.ID, task.LabelIDs); err != nil {
			slog.ErrorContext(ctx, "Create Task: failed createTaskLabels")
			return err
		}

		if err := tr.refreshTask(ctx, tx, &model); err != nil {
			slog.ErrorContext(ctx, "Create Task: failed refreeshTask")
			return domain.ErrUnexpected
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return model.toEntity(), nil
}

func (tr *taskRepository) UpdateTask(ctx context.Context, userID domain.UserID, taskID domain.TaskID, task domain.UpdateTask) (*domain.Task, error) {
	var model taskModel

	err := tr.db.Transaction(func(tx *gorm.DB) error {
		tx = tx.WithContext(ctx)

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ? and id = ?", userID, taskID).First(&model).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrNotFound
			}
			slog.ErrorContext(ctx, "Update Task: failed get task by user_id and id", "error", err)
			return domain.ErrUnexpected
		}

		labelCount, err := tr.getLabelCount(ctx, tx, userID, task.LabelIDs)
		if err != nil {
			slog.ErrorContext(ctx, "Update Task: failed getLabelCount")
			return err
		}
		if labelCount != int64(len(task.LabelIDs)) {
			slog.WarnContext(ctx, "Update Task: label count is unmatch")
			return domain.ErrBadRequest
		}

		model.Title = task.Title
		model.Completed = task.Completed
		if err := tx.Save(&model).Error; err != nil {
			slog.ErrorContext(ctx, "Update Task: failed update task", "error", err)
			return domain.ErrUnexpected
		}

		if err := tx.Where("user_id = ? and task_id = ?", userID, taskID).Delete(&taskLabelModel{}).Error; err != nil {
			slog.ErrorContext(ctx, "Update Task: failed delete task labels", "error", err)
			return domain.ErrUnexpected
		}

		if err := tr.createTaskLabels(ctx, tx, userID, uint(taskID), task.LabelIDs); err != nil {
			slog.ErrorContext(ctx, "Update Task: failed createTaskLabels")
			return err
		}

		if err := tr.refreshTask(ctx, tx, &model); err != nil {
			slog.ErrorContext(ctx, "Update Task: failed refreshTask")
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return model.toEntity(), nil
}

func (tr *taskRepository) DeleteTask(ctx context.Context, userID domain.UserID, taskID domain.TaskID) error {
	return tr.db.Transaction(func(tx *gorm.DB) error {
		tx = tx.WithContext(ctx)

		var model taskModel
		if err := tx.Where("user_id = ? and id = ?", userID, taskID).First(&model).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrNotFound
			}
			slog.ErrorContext(ctx, "Delete Task: failed get task by user_id and id", "error", err)
			return domain.ErrUnexpected
		}

		if err := tx.Where("id = ?", taskID).Delete(&taskModel{}).Error; err != nil {
			slog.ErrorContext(ctx, "Delete Task: failed delete task", "error", err)
			return domain.ErrUnexpected
		}

		// TaskとLabelの紐づけの中間テーブルはON DELETE CASCADEに任せる

		return nil
	})
}

func (tr *taskRepository) getLabelCount(ctx context.Context, db *gorm.DB, userID domain.UserID, labelIDs []domain.LabelID) (int64, error) {
	var labels []labelModel
	if err := db.Clauses(clause.Locking{Strength: "SHARE"}).Select("id").Where("user_id = ? and id in ?", userID, labelIDs).Find(&labels).Error; err != nil {
		slog.ErrorContext(ctx, "get Label Count: failed get label by user_id and ids", "error", err)
		return 0, domain.ErrUnexpected
	}

	return int64(len(labels)), nil
}

func (tr *taskRepository) createTaskLabels(ctx context.Context, db *gorm.DB, userID domain.UserID, taskID uint, labelIDs []domain.LabelID) error {
	taskLabels := make([]taskLabelModel, 0, len(labelIDs))
	for _, labelID := range labelIDs {
		taskLabels = append(taskLabels, taskLabelModel{
			UserID:  uint(userID),
			TaskID:  taskID,
			LabelID: uint(labelID),
		})
	}

	if len(taskLabels) > 0 {
		if err := db.Create(&taskLabels).Error; err != nil {
			slog.ErrorContext(ctx, "create Task Labels: failed create label association", "error", err)
			return domain.ErrUnexpected
		}
	}

	return nil
}

func (tr *taskRepository) refreshTask(ctx context.Context, db *gorm.DB, model *taskModel) error {
	if err := db.Preload("Labels").First(model, model.ID).Error; err != nil {
		slog.ErrorContext(ctx, "refresh Task: failed get created task", "error", err)
		return domain.ErrUnexpected
	}

	return nil
}

package usecase

import (
	"context"
	"log/slog"

	"github.com/matsukawakohei/todo-app-go/domain"
)

type Task struct {
	ID        uint
	Title     string
	Completed bool
	Labels    []*Label
}

func fromTaskEntity(task *domain.Task) *Task {
	labels := make([]*Label, 0, len(task.Labels))
	for _, l := range task.Labels {
		labels = append(labels, fromLabelEntity(l))
	}
	return &Task{
		ID:        uint(task.ID),
		Title:     task.Title,
		Completed: task.Completed,
		Labels:    labels,
	}
}

type GetAllTasksOutput struct {
	Tasks []*Task
}

type CreateTaskInput struct {
	Title  string
	Labels []uint
}

func (cti *CreateTaskInput) toEntityInput() domain.CreateTask {
	labelIDs := make([]domain.LabelID, 0, len(cti.Labels))
	for _, id := range cti.Labels {
		labelIDs = append(labelIDs, domain.LabelID(id))
	}
	return domain.CreateTask{
		Title:    cti.Title,
		LabelIDs: labelIDs,
	}
}

type UpdateTaskInput struct {
	Title     string
	Completed bool
	Labels    []uint
}

func (uti *UpdateTaskInput) toEntityInput() domain.UpdateTask {
	labelIDs := make([]domain.LabelID, 0, len(uti.Labels))
	for _, id := range uti.Labels {
		labelIDs = append(labelIDs, domain.LabelID(id))
	}
	return domain.UpdateTask{
		Title:     uti.Title,
		Completed: uti.Completed,
		LabelIDs:  labelIDs,
	}
}

type TaskUsecase interface {
	GetAllTasksByUserID(ctx context.Context, userID domain.UserID) (*GetAllTasksOutput, error)
	CreateTask(ctx context.Context, userID domain.UserID, input CreateTaskInput) (*Task, error)
	UpdateTask(ctx context.Context, userID domain.UserID, taskID domain.TaskID, input UpdateTaskInput) (*Task, error)
	DeleteTask(ctx context.Context, userID domain.UserID, taskID domain.TaskID) error
}

type taskUsecase struct {
	tr domain.TaskRepository
	tv TaskValidator
}

func NewTaskUsecase(tr domain.TaskRepository, tv TaskValidator) TaskUsecase {
	return &taskUsecase{tr, tv}
}

func (tu *taskUsecase) GetAllTasksByUserID(ctx context.Context, userID domain.UserID) (*GetAllTasksOutput, error) {
	tasks, err := tu.tr.GetAllTasksByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	results := make([]*Task, 0, len(tasks))
	for _, t := range tasks {
		results = append(results, fromTaskEntity(t))
	}

	return &GetAllTasksOutput{Tasks: results}, nil
}

func (tu *taskUsecase) CreateTask(ctx context.Context, userID domain.UserID, input CreateTaskInput) (*Task, error) {
	if err := tu.tv.CreateTaskValidate(ctx, input); err != nil {
		slog.WarnContext(ctx, "failed create task validation", "error", err)
		return nil, err
	}

	task, err := tu.tr.CreateTask(ctx, userID, input.toEntityInput())
	if err != nil {
		return nil, err
	}

	return fromTaskEntity(task), nil
}

func (tu *taskUsecase) UpdateTask(ctx context.Context, userID domain.UserID, taskID domain.TaskID, input UpdateTaskInput) (*Task, error) {
	if err := tu.tv.UpdateTaskValidate(ctx, input); err != nil {
		slog.WarnContext(ctx, "failed update task validation", "error", err)
		return nil, err
	}

	task, err := tu.tr.UpdateTask(ctx, userID, taskID, input.toEntityInput())
	if err != nil {
		return nil, err
	}

	return fromTaskEntity(task), err
}

func (tu *taskUsecase) DeleteTask(ctx context.Context, userID domain.UserID, taskID domain.TaskID) error {
	return tu.tr.DeleteTask(ctx, userID, taskID)
}

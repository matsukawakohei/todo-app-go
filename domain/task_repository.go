package domain

import "context"

type TaskRepository interface {
	GetAllTasksByUserID(ctx context.Context, userID UserID) ([]*Task, error)
	CreateTask(ctx context.Context, userID UserID, task CreateTask) (*Task, error)
	UpdateTask(ctx context.Context, userID UserID, taskID TaskID, task UpdateTask) (*Task, error)
	DeleteTask(ctx context.Context, userID UserID, taskID TaskID) error
}

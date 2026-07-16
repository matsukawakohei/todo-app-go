package domain

import "time"

type TaskID uint

type Task struct {
	ID        TaskID
	Title     string
	Completed bool
	Labels    []*Label
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateTask struct {
	Title    string
	LabelIDs []LabelID
}

type UpdateTask struct {
	Title     string
	Completed bool
	LabelIDs  []LabelID
}

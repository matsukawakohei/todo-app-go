package domain

import "time"

type LabelID uint

type Label struct {
	ID        LabelID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateLabel struct {
	Name string
}

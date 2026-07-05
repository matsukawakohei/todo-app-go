package domain

import "context"

type LabelRepository interface {
	GetAllByUserID(ctx context.Context, userID UserID) ([]*Label, error)
	CreateLabel(ctx context.Context, userID UserID, label CreateLabel) (*Label, error)
	DeleteLabel(ctx context.Context, userID UserID, labelID LabelID) error
}

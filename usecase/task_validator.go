package usecase

import (
	"context"
	"errors"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type TaskValidator interface {
	CreateTaskValidate(ctx context.Context, input CreateTaskInput) error
	UpdateTaskValidate(ctx context.Context, input UpdateTaskInput) error
}

var titleRequiredRule = validation.By(func(value any) error {
	title, _ := value.(string)
	if strings.TrimSpace(title) == "" {
		return errors.New("title is required")
	}
	return nil
})

type taskValidator struct{}

func NewTaskValidator() TaskValidator {
	return &taskValidator{}
}

func (tv *taskValidator) CreateTaskValidate(ctx context.Context, input CreateTaskInput) error {
	return validation.ValidateStructWithContext(ctx,
		&input,
		validation.Field(
			&input.Title,
			titleRequiredRule,
			validation.RuneLength(1, 50).Error("limited max 50 chars"),
		),
		validation.Field(
			&input.Labels,
			validation.Each(validation.By(func(value any) error {
				id, ok := value.(uint)
				if !ok {
					return errors.New("invalid label id")
				}
				if id == 0 {
					return errors.New("label id must be greater than 0")
				}
				return nil
			})),
		),
	)
}

func (tv *taskValidator) UpdateTaskValidate(ctx context.Context, input UpdateTaskInput) error {
	return validation.ValidateStructWithContext(ctx,
		&input,
		validation.Field(
			&input.Title,
			titleRequiredRule,
			validation.RuneLength(1, 50).Error("limited max 50 chars"),
		),
		validation.Field(
			&input.Labels,
			validation.Each(validation.By(func(value any) error {
				id, ok := value.(uint)
				if !ok {
					return errors.New("invalid label id")
				}
				if id == 0 {
					return errors.New("label id must be greater than 0")
				}
				return nil
			})),
		),
	)
}

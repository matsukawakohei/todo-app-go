package usecase

import (
	"context"
	"errors"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type TaskValidator interface {
	CreateTaskValidate(ctx context.Context, input CreateTaskInput) error
	UpdateTaskValidate(ctx context.Context, input UpdateTaskInput) error
}

type taskValidator struct{}

func NewTaskValidator() TaskValidator {
	return &taskValidator{}
}

func (tv *taskValidator) CreateTaskValidate(ctx context.Context, input CreateTaskInput) error {
	return validation.ValidateStructWithContext(ctx,
		&input,
		validation.Field(
			&input.Title,
			validation.Required.Error("title is required"),
			validation.RuneLength(1, 50).Error("limited max 50 chars"),
		),
		validation.Field(
			&input.Labels,
			validation.Each(validation.By(func(value interface{}) error {
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
			validation.Required.Error("title is required"),
			validation.RuneLength(1, 50).Error("limited max 50 chars"),
		),
		validation.Field(
			&input.Labels,
			validation.Each(validation.By(func(value interface{}) error {
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

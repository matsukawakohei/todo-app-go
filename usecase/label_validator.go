package usecase

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type LabelValidator interface {
	CreateValidate(ctx context.Context, input LabelCreateInput) error
}

type labelValidator struct {}

func NewLabelValidator() LabelValidator {
	return &labelValidator{}
}

func (lv *labelValidator) CreateValidate(ctx context.Context, input LabelCreateInput) error {
	return validation.ValidateStructWithContext(ctx,
		&input,
		validation.Field(
			&input.Name,
			validation.Required.Error("name is required"),
			validation.RuneLength(1, 50).Error("limited max 50 char"),
		),
	)
}
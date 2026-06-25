package usecase

import (
	"context"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type IUserValidator interface {
	SignUpValidate(ctx context.Context, input SignUpInput) error
	LoginValidate(ctx context.Context, input LoginInput) error
}

type userValidator struct{}

func NewUserValidator() IUserValidator {
	return &userValidator{}
}

func (uv *userValidator) SignUpValidate(ctx context.Context, input SignUpInput) error {
	return validation.ValidateStructWithContext(ctx,
		&input,
		validation.Field(
			&input.Name,
			validation.Required.Error("name is required"),
			validation.RuneLength(1, 50).Error("limited max 50 char"),
		),
		validation.Field(
			&input.Email,
			validation.Required.Error("email is required"),
			validation.RuneLength(1, 50).Error("limited max 50 char"),
			is.Email.Error("is not valid email format"),
		),
		validation.Field(
			&input.Password,
			validation.Required.Error("password is required"),
			validation.RuneLength(6, 30).Error("limited min 6 max 30 char"),
			validation.Match(regexp.MustCompile(`[a-z]`)).Error("required lower case"),
			validation.Match(regexp.MustCompile(`[A-Z]`)).Error("required upper case"),
			validation.Match(regexp.MustCompile(`[0-9]`)).Error("required number"),
			validation.Match(regexp.MustCompile(`^[a-zA-Z0-9]+$`)).Error("invalid format"),
		),
	)
}

func (uv *userValidator) LoginValidate(ctx context.Context, input LoginInput) error {
	return validation.ValidateStructWithContext(ctx,
		&input,
		validation.Field(
			&input.Email,
			validation.Required.Error("email is required"),
			validation.RuneLength(1, 50).Error("limited max 50 char"),
			is.Email.Error("is not valid email format"),
		),
		validation.Field(
			&input.Password,
			validation.Required.Error("password is required"),
			validation.RuneLength(6, 30).Error("limited min 6 max 30 char"),
		),
	)
}

package usecase

import (
	"context"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

var (
	lowercasePasswordRe = regexp.MustCompile(`[a-z]`)
	uppercasePasswordRe = regexp.MustCompile(`[A-Z]`)
	numberPasswordRe    = regexp.MustCompile(`[0-9]`)
	alnumPasswordRe     = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
)

type UserValidator interface {
	SignUpValidate(ctx context.Context, input SignUpInput) error
	LoginValidate(ctx context.Context, input LoginInput) error
}

type userValidator struct{}

func NewUserValidator() UserValidator {
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
			validation.Match(lowercasePasswordRe).Error("required lower case"),
			validation.Match(uppercasePasswordRe).Error("required upper case"),
			validation.Match(numberPasswordRe).Error("required number"),
			validation.Match(alnumPasswordRe).Error("invalid format"),
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

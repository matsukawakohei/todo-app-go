package domain

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrBadRequest = errors.New("input params are invalid")
	ErrNotFound = errors.New("object not found")
	ErrUnexpected = errors.New("un unexpected error occurred")
)
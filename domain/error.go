package domain

import "errors"

var (
	ErrBadRequest = errors.New("input params are invalid")
	ErrNotFound = errors.New("object not found")
	ErrUnExpected = errors.New("un unexpected error occurred")
)
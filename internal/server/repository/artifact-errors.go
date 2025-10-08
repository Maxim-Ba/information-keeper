package repository

import "errors"

var (
	ErrArtifactNotFound = errors.New("artifact not found")
	ErrForbiddenAction  = errors.New("forbidden action")
	ErrInvalidInput     = errors.New("invalid input")
	ErrNotFound         = errors.New("not found")
)

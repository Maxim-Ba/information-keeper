package services

import "errors"

var (
	ErrUserAllreadyExists    = errors.New("user already exists")
	ErrChangeNotYourPassword = errors.New("you can't change not your password")
	ErrEmptyPassword         = errors.New("password can't be empty")
	ErrInvalidEmail          = errors.New("invalid email format")
)

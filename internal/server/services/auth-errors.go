package services

import "errors"

var ErrUserAllreadyExists = errors.New("user already exists")
var ErrChangeNotYourPassword = errors.New("you can't change not your password")
var ErrEmptyPassword = errors.New("password can't be empty")

package utils

import "errors"

// ErrEmptyPassword is returned when the password is empty.
var ErrEmptyPassword = errors.New("password is empty")
// ErrEmptySecret is returned when the secret is empty.
var ErrEmptySecret = errors.New("secret is empty")

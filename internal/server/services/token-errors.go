package services

import "errors"

var (
	ErrUserNotFound            = errors.New("user not found")
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrTokenExpired            = errors.New("token expired")
	ErrInvalidToken            = errors.New("invalid token")
	ErrTokenRevoked            = errors.New("token revoked")
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
	ErrTokenValidationFailed   = errors.New("token validation failed")
	ErrInvalidTokenClaims      = errors.New("invalid token claims")
	ErrInvalidTokenType        = errors.New("invalid token type")
	ErrAuthTokenNotProvided    = errors.New("authorization token not provided")
	ErrMetaDataNotProvided     = errors.New("meta data token not provided")
)

package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// PasswordManager struct
type PasswordManager struct{}

// HashPassword хеширует пароль с использованием пеппер-секрета и bcrypt
func (pm *PasswordManager) HashPassword(password, secret string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("HashPassword: %w", ErrEmptyPassword)
	}
	if secret == "" {
		return "", fmt.Errorf("HashPassword: %w", ErrEmptySecret)
	}
	pepperedPassword := hmac.New(sha256.New, []byte(secret))
	pepperedPassword.Write([]byte(password))
	peppered := hex.EncodeToString(pepperedPassword.Sum(nil))

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(peppered), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("HashPassword GenerateFromPassword: %w", err)
	}

	return string(hashedBytes), nil
}
// CheckPassword проверяет соответствие пароля хешированному значению
func (pm *PasswordManager) CheckPassword(hashedPassword, password, secret string) error {
	pepperedPassword := hmac.New(sha256.New, []byte(secret))
	pepperedPassword.Write([]byte(password))
	peppered := hex.EncodeToString(pepperedPassword.Sum(nil))

	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(peppered))
}

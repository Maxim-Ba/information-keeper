package repository

import (
	"database/sql"
	"time"
)

type TokenRepository struct {
	db *sql.DB
}

// AddToBlacklist implements services.TokenRepositoryInterface.
func (t *TokenRepository) AddToBlacklist(token string, expiry time.Time) error {
	panic("unimplemented")
}

// IsInBlacklist implements services.TokenRepositoryInterface.
func (t *TokenRepository) IsInBlacklist(token string) (bool, error) {
	panic("unimplemented")
}

func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{
		db: db,
	}
}

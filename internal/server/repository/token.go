package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type TokenRepository struct {
	db DBInterface
}

func (t *TokenRepository) AddToBlacklist(ctx context.Context, token string, expiry time.Time) error {
	query := `
		INSERT INTO token_blacklist (token, expiry)
		VALUES ($1, $2)
		ON CONFLICT (token) DO UPDATE SET expiry = EXCLUDED.expiry
	`

	_, err := t.db.ExecContext(ctx, query, token, expiry)
	if err != nil {
		return fmt.Errorf("failed to add token to blacklist: %w", err)
	}

	return nil
}

func (t *TokenRepository) IsInBlacklist(ctx context.Context, token string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM token_blacklist 
			WHERE token = $1 AND expiry > CURRENT_TIMESTAMP
		)
	`

	var exists bool
	err := t.db.QueryRowContext(ctx, query, token).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check token in blacklist: %w", err)
	}

	return exists, nil
}

func (t *TokenRepository) CleanExpiredTokens(ctx context.Context) error {
	query := `
		DELETE FROM token_blacklist 
		WHERE expiry <= CURRENT_TIMESTAMP
	`

	_, err := t.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to clean expired tokens: %w", err)
	}

	return nil
}

func NewTokenRepository(db DBInterface) *TokenRepository {
	return &TokenRepository{
		db: db,
	}
}

//go:generate mockgen -source=$GOFILE -destination=./mocks/app_config_mock.go -package=mocks AppConfig PasswordManagerInterface
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	config "github.com/Maxim-Ba/information-keeper/config/server"
	"github.com/Maxim-Ba/information-keeper/internal/server/dto"
	"github.com/Maxim-Ba/information-keeper/pkg/logger"
)

type AppConfig interface {
	GetConfig() config.ServerCfg
}
type PasswordManagerInterface interface {
	CheckPassword(hashedPassword, password, secret string) error
	HashPassword(password, secret string) (string, error)
}
type DBInterface interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}
type UserRepository struct {
	db          DBInterface
	cfg         AppConfig
	pswdManager PasswordManagerInterface
}

func (u *UserRepository) Login(ctx context.Context, login string, password string) (*dto.UserRepoDTO, error) {
	logger.Info("Login user", "login", login, "password", password)
	query := `
		SELECT id, login, email, email_confirmed, password
		FROM users 
		WHERE login = $1
	`
	var storedPasswordHash string
	var user dto.UserRepoDTO
	err := u.db.QueryRowContext(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.Email,
		&user.EmailConfirmed,
		&storedPasswordHash,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to login user: %w", err)
	}
	if err := u.pswdManager.CheckPassword(storedPasswordHash, password, u.cfg.GetConfig().PasswordSecret); err != nil {
		return nil, ErrInvalidCredentials
	}

	return &user, nil
}

func (u *UserRepository) ChangePassword(ctx context.Context, login string, oldPassword string, newPassword string) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Проверяем старый пароль
	checkQuery := `
		SELECT id, password FROM users 
		WHERE login = $1 
	`
	var userID string
	var storedPasswordHash string
	err = tx.QueryRowContext(ctx, checkQuery, login).Scan(&userID, &storedPasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidCredentials
		}
		return fmt.Errorf("failed to verify old password: %w", err)
	}
	if err := u.pswdManager.CheckPassword(storedPasswordHash, oldPassword, u.cfg.GetConfig().PasswordSecret); err != nil {
		return ErrInvalidCredentials
	}

	updateQuery := `
		UPDATE users 
		SET password = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`
	newPasswordHash, err := u.pswdManager.HashPassword(newPassword, u.cfg.GetConfig().PasswordSecret)
	if err != nil {
		return fmt.Errorf("failed to hash new password in user repo: %w", err)
	}
	_, err = tx.ExecContext(ctx, updateQuery, newPasswordHash, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (u *UserRepository) GetUserByEmail(ctx context.Context, email string) (*dto.UserRepoDTO, error) {
	query := `
		SELECT id, login, email, email_confirmed
		FROM users 
		WHERE email = $1
	`

	var user dto.UserRepoDTO
	err := u.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Login,
		&user.Email,
		&user.EmailConfirmed,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

func (u *UserRepository) Register(ctx context.Context, login string, email string, password string) (*dto.UserRepoDTO, error) {
	// Проверяем, существует ли пользователь с таким логином или email
	checkQuery := `
		SELECT id FROM users WHERE login = $1 OR email = $2
	`
	var existingID string
	err := u.db.QueryRowContext(ctx, checkQuery, login, email).Scan(&existingID)
	if err == nil {
		return nil, ErrUserAlreadyExists
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// Создаем нового пользователя с использованием gen_random_uuid()
	query := `
		INSERT INTO users (id, login, password, email, email_confirmed)
		VALUES (gen_random_uuid(), $1, $2, $3, $4)
		RETURNING id, login, email, email_confirmed
	`

	var user dto.UserRepoDTO
	err = u.db.QueryRowContext(ctx, query,
		login,
		password,
		email,
		false, // email_confirmed
	).Scan(
		&user.ID,
		&user.Login,
		&user.Email,
		&user.EmailConfirmed,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	return &user, nil
}

func (u *UserRepository) Update(ctx context.Context, user *dto.UserRepoDTO) (*dto.UserRepoDTO, error) {
	query := `
		UPDATE users 
		SET login = $1, email = $2, email_confirmed = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
		RETURNING id, login, email, email_confirmed
	`

	var updatedUser dto.UserRepoDTO
	err := u.db.QueryRowContext(ctx, query,
		user.Login,
		user.Email,
		user.EmailConfirmed,
		user.ID,
	).Scan(
		&updatedUser.ID,
		&updatedUser.Login,
		&updatedUser.Email,
		&updatedUser.EmailConfirmed,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &updatedUser, nil
}

func (u *UserRepository) GetUserByID(ctx context.Context, id string) (*dto.UserRepoDTO, error) {
	query := `
		SELECT id, login, email, email_confirmed
		FROM users 
		WHERE id = $1
	`

	var user dto.UserRepoDTO
	err := u.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Login,
		&user.Email,
		&user.EmailConfirmed,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

func NewUserRepository(db DBInterface, cfg AppConfig, pswdManager PasswordManagerInterface) *UserRepository {
	return &UserRepository{
		db:          db,
		cfg:         cfg,
		pswdManager: pswdManager,
	}
}

// user_test.go
package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	config "github.com/Maxim-Ba/information-keeper/config/server"
	"github.com/Maxim-Ba/information-keeper/internal/server/dto"
	"github.com/Maxim-Ba/information-keeper/internal/server/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// MockDBWrapper обертка для совместимости с sqlmock
type MockDBWrapper struct {
	db *sql.DB
}

func (m *MockDBWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return m.db.QueryRowContext(ctx, query, args...)
}

func (m *MockDBWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return m.db.ExecContext(ctx, query, args...)
}

func (m *MockDBWrapper) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return m.db.BeginTx(ctx, opts)
}

func TestUserRepository_Login(t *testing.T) {

	type args struct {
		login    string
		password string
	}

	type mockBehavior func(mock sqlmock.Sqlmock, args args)
	type passwordManagerBehavior func(mockPM *mocks.MockPasswordManagerInterface, storedHash, password, secret string)
	type testCase struct {
		name                    string
		args                    args
		mockBehavior            mockBehavior
		passwordManagerBehavior passwordManagerBehavior
		wantUser                *dto.UserRepoDTO
		wantErr                 bool
		expectErr               error
	}

	tests := []testCase{
		{
			name: "Success by login",
			args: args{
				login:    "testuser",
				password: "password123",
			},
			mockBehavior: func(mock sqlmock.Sqlmock, args args) {
				rows := sqlmock.NewRows([]string{"id", "login", "email", "email_confirmed", "password"}).
					AddRow("123", "testuser", "test@example.com", true, "hashed_password123")

				mock.ExpectQuery(`SELECT id, login, email, email_confirmed, password FROM users WHERE login = \$1`).
					WithArgs(args.login).
					WillReturnRows(rows)
			},
			passwordManagerBehavior: func(mockPM *mocks.MockPasswordManagerInterface, storedHash, password, secret string) {
				mockPM.EXPECT().CheckPassword("hashed_password123", "password123", "test-secret").Return(nil)
			},
			wantUser: &dto.UserRepoDTO{
				ID:             "123",
				Login:          "testuser",
				Email:          "test@example.com",
				EmailConfirmed: true,
			},
			wantErr:   false,
			expectErr: nil,
		},
		{
			name: "Invalid credentials - no rows",
			args: args{
				login:    "wronguser",
				password: "wrongpass",
			},
			mockBehavior: func(mock sqlmock.Sqlmock, args args) {
				mock.ExpectQuery(`SELECT id, login, email, email_confirmed, password FROM users WHERE login = \$1`).
					WithArgs(args.login).
					WillReturnError(sql.ErrNoRows)
			},
			passwordManagerBehavior: func(mockPM *mocks.MockPasswordManagerInterface, storedHash, password, secret string) {
				// Password manager не должен вызываться при ошибке базы данных
			},
			wantUser:  nil,
			wantErr:   true,
			expectErr: ErrInvalidCredentials,
		},
		{
			name: "Invalid credentials - wrong password",
			args: args{
				login:    "testuser",
				password: "wrongpassword",
			},
			mockBehavior: func(mock sqlmock.Sqlmock, args args) {
				rows := sqlmock.NewRows([]string{"id", "login", "email", "email_confirmed", "password"}).
					AddRow("123", "testuser", "test@example.com", true, "hashed_password123")

				mock.ExpectQuery(`SELECT id, login, email, email_confirmed, password FROM users WHERE login = \$1`).
					WithArgs(args.login).
					WillReturnRows(rows)
			},
			passwordManagerBehavior: func(mockPM *mocks.MockPasswordManagerInterface, storedHash, password, secret string) {
				mockPM.EXPECT().CheckPassword("hashed_password123", "wrongpassword", "test-secret").Return(errors.New("invalid password"))
			},
			wantUser:  nil,
			wantErr:   true,
			expectErr: ErrInvalidCredentials,
		},
		{
			name: "Database error",
			args: args{
				login:    "testuser",
				password: "password123",
			},
			mockBehavior: func(mock sqlmock.Sqlmock, args args) {
				dbError := errors.New("database connection failed")
				mock.ExpectQuery(`SELECT id, login, email, email_confirmed, password FROM users WHERE login = \$1`).
					WithArgs(args.login).
					WillReturnError(dbError)
			},
			passwordManagerBehavior: func(mockPM *mocks.MockPasswordManagerInterface, storedHash, password, secret string) {
				// Password manager не должен вызываться при ошибке базы данных
			},
			wantUser:  nil,
			wantErr:   true,
			expectErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAppConfig := mocks.NewMockAppConfig(ctrl)
			mockAppConfig.EXPECT().GetConfig().Return(config.ServerCfg{
				PasswordSecret: "test-secret",
			}).AnyTimes()

			mockPasswordManager := mocks.NewMockPasswordManagerInterface(ctrl)
			if tt.passwordManagerBehavior != nil {
				// Мы вызовем это позже, после настройки mock'ов
			}

			mockDB := &MockDBWrapper{db: db}
			repo := NewUserRepository(mockDB, mockAppConfig, mockPasswordManager)

			tt.mockBehavior(mock, tt.args)

			// Вызываем passwordManagerBehavior после настройки mock'ов
			if tt.passwordManagerBehavior != nil {
				tt.passwordManagerBehavior(mockPasswordManager, "hashed_password123", tt.args.password, "test-secret")
			}

			ctx := context.Background()
			gotUser, err := repo.Login(ctx, tt.args.login, tt.args.password)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectErr != nil {
					assert.True(t, errors.Is(err, tt.expectErr), "Expected error: %v, got: %v", tt.expectErr, err)
				} else {
					assert.Contains(t, err.Error(), "failed to login user")
				}
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.wantUser, gotUser)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_ChangePassword(t *testing.T) {
	type args struct {
		login       string
		oldPassword string
		newPassword string
	}

	type mockBehavior func(mock sqlmock.Sqlmock, args args)
	type passwordManagerBehavior func(mockPM *mocks.MockPasswordManagerInterface)
	type testCase struct {
		name                    string
		args                    args
		mockBehavior            mockBehavior
		passwordManagerBehavior passwordManagerBehavior
		wantErr                 bool
		expectErr               error
	}

	tests := []testCase{
		{
			name: "Success",
			args: args{
				login:       "testuser",
				oldPassword: "oldpass",
				newPassword: "newpass",
			},
			mockBehavior: func(mock sqlmock.Sqlmock, args args) {
				mock.ExpectBegin()
				// Получаем ID и хеш пароля пользователя
				rows := sqlmock.NewRows([]string{"id", "password"}).AddRow("123", "hashed_oldpass")
				mock.ExpectQuery(`SELECT id, password FROM users WHERE login = \$1`).
					WithArgs(args.login).
					WillReturnRows(rows)
				// Обновляем пароль
				mock.ExpectExec(`UPDATE users SET password = \$1, updated_at = CURRENT_TIMESTAMP WHERE id = \$2`).
					WithArgs("hashed_newpass", "123").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			passwordManagerBehavior: func(mockPM *mocks.MockPasswordManagerInterface) {
				// Проверяем старый пароль
				mockPM.EXPECT().CheckPassword("hashed_oldpass", "oldpass", "test-secret").Return(nil)
				// Хешируем новый пароль
				mockPM.EXPECT().HashPassword("newpass", "test-secret").Return("hashed_newpass", nil)
			},
			wantErr:   false,
			expectErr: nil,
		},
		{
			name: "Invalid old password",
			args: args{
				login:       "testuser",
				oldPassword: "wrongpass",
				newPassword: "newpass",
			},
			mockBehavior: func(mock sqlmock.Sqlmock, args args) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id", "password"}).AddRow("123", "hashed_oldpass")
				mock.ExpectQuery(`SELECT id, password FROM users WHERE login = \$1`).
					WithArgs(args.login).
					WillReturnRows(rows)
				mock.ExpectRollback()
			},
			passwordManagerBehavior: func(mockPM *mocks.MockPasswordManagerInterface) {
				mockPM.EXPECT().CheckPassword("hashed_oldpass", "wrongpass", "test-secret").Return(errors.New("invalid password"))
			},
			wantErr:   true,
			expectErr: ErrInvalidCredentials,
		},
		{
			name: "User not found",
			args: args{
				login:       "nonexistent",
				oldPassword: "oldpass",
				newPassword: "newpass",
			},
			mockBehavior: func(mock sqlmock.Sqlmock, args args) {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT id, password FROM users WHERE login = \$1`).
					WithArgs(args.login).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectRollback()
			},
			passwordManagerBehavior: func(mockPM *mocks.MockPasswordManagerInterface) {
				// Password manager не должен вызываться при ошибке базы данных
			},
			wantErr:   true,
			expectErr: ErrInvalidCredentials,
		},
		{
			name: "Database error on check",
			args: args{
				login:       "testuser",
				oldPassword: "oldpass",
				newPassword: "newpass",
			},
			mockBehavior: func(mock sqlmock.Sqlmock, args args) {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT id, password FROM users WHERE login = \$1`).
					WithArgs(args.login).
					WillReturnError(errors.New("db error"))
				mock.ExpectRollback()
			},
			passwordManagerBehavior: func(mockPM *mocks.MockPasswordManagerInterface) {
				// Password manager не должен вызываться при ошибке базы данных
			},
			wantErr:   true,
			expectErr: nil,
		},
		{
			name: "Error hashing new password",
			args: args{
				login:       "testuser",
				oldPassword: "oldpass",
				newPassword: "newpass",
			},
			mockBehavior: func(mock sqlmock.Sqlmock, args args) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id", "password"}).AddRow("123", "hashed_oldpass")
				mock.ExpectQuery(`SELECT id, password FROM users WHERE login = \$1`).
					WithArgs(args.login).
					WillReturnRows(rows)
				mock.ExpectRollback()
			},
			passwordManagerBehavior: func(mockPM *mocks.MockPasswordManagerInterface) {
				mockPM.EXPECT().CheckPassword("hashed_oldpass", "oldpass", "test-secret").Return(nil)
				mockPM.EXPECT().HashPassword("newpass", "test-secret").Return("", errors.New("hashing error"))
			},
			wantErr:   true,
			expectErr: nil,
		},
		{
			name: "Database error on update",
			args: args{
				login:       "testuser",
				oldPassword: "oldpass",
				newPassword: "newpass",
			},
			mockBehavior: func(mock sqlmock.Sqlmock, args args) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id", "password"}).AddRow("123", "hashed_oldpass")
				mock.ExpectQuery(`SELECT id, password FROM users WHERE login = \$1`).
					WithArgs(args.login).
					WillReturnRows(rows)
				mock.ExpectExec(`UPDATE users SET password = \$1, updated_at = CURRENT_TIMESTAMP WHERE id = \$2`).
					WithArgs("hashed_newpass", "123").
					WillReturnError(errors.New("update error"))
				mock.ExpectRollback()
			},
			passwordManagerBehavior: func(mockPM *mocks.MockPasswordManagerInterface) {
				mockPM.EXPECT().CheckPassword("hashed_oldpass", "oldpass", "test-secret").Return(nil)
				mockPM.EXPECT().HashPassword("newpass", "test-secret").Return("hashed_newpass", nil)
			},
			wantErr:   true,
			expectErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAppConfig := mocks.NewMockAppConfig(ctrl)
			mockAppConfig.EXPECT().GetConfig().Return(config.ServerCfg{
				PasswordSecret: "test-secret",
			}).AnyTimes()

			mockPasswordManager := mocks.NewMockPasswordManagerInterface(ctrl)
			if tt.passwordManagerBehavior != nil {
				tt.passwordManagerBehavior(mockPasswordManager)
			}

			mockDB := &MockDBWrapper{db: db}
			repo := NewUserRepository(mockDB, mockAppConfig, mockPasswordManager)

			tt.mockBehavior(mock, tt.args)

			ctx := context.Background()
			err = repo.ChangePassword(ctx, tt.args.login, tt.args.oldPassword, tt.args.newPassword)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectErr != nil {
					assert.True(t, errors.Is(err, tt.expectErr))
				}
			} else {
				require.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
func TestUserRepository_GetUserByEmail(t *testing.T) {
	type args struct {
		email string
	}

	type mockBehavior func(mock sqlmock.Sqlmock, args args)
	type testCase struct {
		name         string
		args         args
		mockBehavior mockBehavior
		wantUser     *dto.UserRepoDTO
		wantErr      bool
		expectErr    error
	}

	tests := []testCase{
		// ... существующие тестовые случаи без изменений
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAppConfig := mocks.NewMockAppConfig(ctrl)
			mockAppConfig.EXPECT().GetConfig().Return(config.ServerCfg{
				PasswordSecret: "test-secret",
			}).AnyTimes()

			mockPasswordManager := mocks.NewMockPasswordManagerInterface(ctrl)

			mockDB := &MockDBWrapper{db: db}
			repo := NewUserRepository(mockDB, mockAppConfig, mockPasswordManager)

			tt.mockBehavior(mock, tt.args)

			ctx := context.Background()
			gotUser, err := repo.GetUserByEmail(ctx, tt.args.email)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectErr != nil {
					assert.True(t, errors.Is(err, tt.expectErr))
				}
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.wantUser, gotUser)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_Register(t *testing.T) {
	type args struct {
		login    string
		email    string
		password string
	}

	type mockBehavior func(mock sqlmock.Sqlmock, args args)
	type testCase struct {
		name         string
		args         args
		mockBehavior mockBehavior
		wantUser     *dto.UserRepoDTO
		wantErr      bool
		expectErr    error
	}

	tests := []testCase{
		{
			name: "Success",
			args: args{
				login:    "newuser",
				email:    "new@example.com",
				password: "password123",
			},
			mockBehavior: func(mock sqlmock.Sqlmock, args args) {
				mock.ExpectQuery(`SELECT id FROM users WHERE login = \$1 OR email = \$2`).
					WithArgs(args.login, args.email).
					WillReturnError(sql.ErrNoRows)

				rows := sqlmock.NewRows([]string{"id", "login", "email", "email_confirmed"}).
					AddRow("uuid-123", "newuser", "new@example.com", false)
				mock.ExpectQuery(`INSERT INTO users \(id, login, password, email, email_confirmed\) VALUES \(gen_random_uuid\(\), \$1, \$2, \$3, \$4\) RETURNING id, login, email, email_confirmed`).
					WithArgs(args.login, args.password, args.email, false).
					WillReturnRows(rows)
			},
			wantUser: &dto.UserRepoDTO{
				ID:             "uuid-123",
				Login:          "newuser",
				Email:          "new@example.com",
				EmailConfirmed: false,
			},
			wantErr:   false,
			expectErr: nil,
		},
		{
			name: "User already exists",
			args: args{
				login:    "existing",
				email:    "existing@example.com",
				password: "password123",
			},
			mockBehavior: func(mock sqlmock.Sqlmock, args args) {
				rows := sqlmock.NewRows([]string{"id"}).AddRow("existing-id")
				mock.ExpectQuery(`SELECT id FROM users WHERE login = \$1 OR email = \$2`).
					WithArgs(args.login, args.email).
					WillReturnRows(rows)
			},
			wantUser:  nil,
			wantErr:   true,
			expectErr: ErrUserAlreadyExists,
		},
		{
			name: "Database error on check",
			args: args{
				login:    "newuser",
				email:    "new@example.com",
				password: "password123",
			},
			mockBehavior: func(mock sqlmock.Sqlmock, args args) {
				mock.ExpectQuery(`SELECT id FROM users WHERE login = \$1 OR email = \$2`).
					WithArgs(args.login, args.email).
					WillReturnError(errors.New("db error"))
			},
			wantUser:  nil,
			wantErr:   true,
			expectErr: nil,
		},
		{
			name: "Database error on insert",
			args: args{
				login:    "newuser",
				email:    "new@example.com",
				password: "password123",
			},
			mockBehavior: func(mock sqlmock.Sqlmock, args args) {
				mock.ExpectQuery(`SELECT id FROM users WHERE login = \$1 OR email = \$2`).
					WithArgs(args.login, args.email).
					WillReturnError(sql.ErrNoRows)

				mock.ExpectQuery(`INSERT INTO users \(id, login, password, email, email_confirmed\) VALUES \(gen_random_uuid\(\), \$1, \$2, \$3, \$4\) RETURNING id, login, email, email_confirmed`).
					WithArgs(args.login, args.password, args.email, false).
					WillReturnError(errors.New("insert error"))
			},
			wantUser:  nil,
			wantErr:   true,
			expectErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockAppConfig := mocks.NewMockAppConfig(ctrl)
			mockAppConfig.EXPECT().GetConfig().Return(config.ServerCfg{
				PasswordSecret: "test-secret",
			}).AnyTimes()
			mockDB := &MockDBWrapper{db: db}
			mockPasswordManager := mocks.NewMockPasswordManagerInterface(ctrl)

			repo := NewUserRepository(mockDB, mockAppConfig, mockPasswordManager)

			tt.mockBehavior(mock, tt.args)

			ctx := context.Background()
			gotUser, err := repo.Register(ctx, tt.args.login, tt.args.email, tt.args.password)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectErr != nil {
					assert.True(t, errors.Is(err, tt.expectErr))
				}
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.wantUser, gotUser)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
func TestUserRepository_Update(t *testing.T) {
	type args struct {
		user *dto.UserRepoDTO
	}

	type mockBehavior func(mock sqlmock.Sqlmock, args args)
	type testCase struct {
		name         string
		args         args
		mockBehavior mockBehavior
		wantUser     *dto.UserRepoDTO
		wantErr      bool
		expectErr    error
	}

	tests := []testCase{
		{
			name: "Success",
			args: args{
				user: &dto.UserRepoDTO{
					ID:             "123",
					Login:          "updateduser",
					Email:          "updated@example.com",
					EmailConfirmed: true,
				},
			},
			mockBehavior: func(mock sqlmock.Sqlmock, args args) {
				rows := sqlmock.NewRows([]string{"id", "login", "email", "email_confirmed"}).
					AddRow("123", "updateduser", "updated@example.com", true)
				mock.ExpectQuery(`UPDATE users SET login = \$1, email = \$2, email_confirmed = \$3, updated_at = CURRENT_TIMESTAMP WHERE id = \$4 RETURNING id, login, email, email_confirmed`).
					WithArgs(args.user.Login, args.user.Email, args.user.EmailConfirmed, args.user.ID).
					WillReturnRows(rows)
			},
			wantUser: &dto.UserRepoDTO{
				ID:             "123",
				Login:          "updateduser",
				Email:          "updated@example.com",
				EmailConfirmed: true,
			},
			wantErr:   false,
			expectErr: nil,
		},
		{
			name: "User not found",
			args: args{
				user: &dto.UserRepoDTO{
					ID:             "nonexistent",
					Login:          "nonexistent",
					Email:          "nonexistent@example.com",
					EmailConfirmed: false,
				},
			},
			mockBehavior: func(mock sqlmock.Sqlmock, args args) {
				mock.ExpectQuery(`UPDATE users SET login = \$1, email = \$2, email_confirmed = \$3, updated_at = CURRENT_TIMESTAMP WHERE id = \$4 RETURNING id, login, email, email_confirmed`).
					WithArgs(args.user.Login, args.user.Email, args.user.EmailConfirmed, args.user.ID).
					WillReturnError(sql.ErrNoRows)
			},
			wantUser:  nil,
			wantErr:   true,
			expectErr: ErrUserNotFound,
		},
		{
			name: "Database error",
			args: args{
				user: &dto.UserRepoDTO{
					ID:             "123",
					Login:          "testuser",
					Email:          "test@example.com",
					EmailConfirmed: true,
				},
			},
			mockBehavior: func(mock sqlmock.Sqlmock, args args) {
				mock.ExpectQuery(`UPDATE users SET login = \$1, email = \$2, email_confirmed = \$3, updated_at = CURRENT_TIMESTAMP WHERE id = \$4 RETURNING id, login, email, email_confirmed`).
					WithArgs(args.user.Login, args.user.Email, args.user.EmailConfirmed, args.user.ID).
					WillReturnError(errors.New("db error"))
			},
			wantUser:  nil,
			wantErr:   true,
			expectErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockAppConfig := mocks.NewMockAppConfig(ctrl)
			mockAppConfig.EXPECT().GetConfig().Return(config.ServerCfg{
				PasswordSecret: "test-secret",
			}).AnyTimes()
			mockDB := &MockDBWrapper{db: db}
			mockPasswordManager := mocks.NewMockPasswordManagerInterface(ctrl)

			repo := NewUserRepository(mockDB, mockAppConfig, mockPasswordManager)

			tt.mockBehavior(mock, tt.args)

			ctx := context.Background()
			gotUser, err := repo.Update(ctx, tt.args.user)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectErr != nil {
					assert.True(t, errors.Is(err, tt.expectErr))
				}
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.wantUser, gotUser)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_GetUserByID(t *testing.T) {
	type args struct {
		id string
	}

	type mockBehavior func(mock sqlmock.Sqlmock, args args)
	type testCase struct {
		name         string
		args         args
		mockBehavior mockBehavior
		wantUser     *dto.UserRepoDTO
		wantErr      bool
		expectErr    error
	}

	tests := []testCase{
		{
			name: "Success",
			args: args{
				id: "123",
			},
			mockBehavior: func(mock sqlmock.Sqlmock, args args) {
				rows := sqlmock.NewRows([]string{"id", "login", "email", "email_confirmed"}).
					AddRow("123", "testuser", "test@example.com", true)
				mock.ExpectQuery(`SELECT id, login, email, email_confirmed FROM users WHERE id = \$1`).
					WithArgs(args.id).
					WillReturnRows(rows)
			},
			wantUser: &dto.UserRepoDTO{
				ID:             "123",
				Login:          "testuser",
				Email:          "test@example.com",
				EmailConfirmed: true,
			},
			wantErr:   false,
			expectErr: nil,
		},
		{
			name: "User not found",
			args: args{
				id: "nonexistent",
			},
			mockBehavior: func(mock sqlmock.Sqlmock, args args) {
				mock.ExpectQuery(`SELECT id, login, email, email_confirmed FROM users WHERE id = \$1`).
					WithArgs(args.id).
					WillReturnError(sql.ErrNoRows)
			},
			wantUser:  nil,
			wantErr:   true,
			expectErr: ErrUserNotFound,
		},
		{
			name: "Database error",
			args: args{
				id: "123",
			},
			mockBehavior: func(mock sqlmock.Sqlmock, args args) {
				mock.ExpectQuery(`SELECT id, login, email, email_confirmed FROM users WHERE id = \$1`).
					WithArgs(args.id).
					WillReturnError(errors.New("db error"))
			},
			wantUser:  nil,
			wantErr:   true,
			expectErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockAppConfig := mocks.NewMockAppConfig(ctrl)
			mockAppConfig.EXPECT().GetConfig().Return(config.ServerCfg{
				PasswordSecret: "test-secret",
			}).AnyTimes()
			mockDB := &MockDBWrapper{db: db}
			mockPasswordManager := mocks.NewMockPasswordManagerInterface(ctrl)
			repo := NewUserRepository(mockDB, mockAppConfig, mockPasswordManager)

			tt.mockBehavior(mock, tt.args)

			ctx := context.Background()
			gotUser, err := repo.GetUserByID(ctx, tt.args.id)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectErr != nil {
					assert.True(t, errors.Is(err, tt.expectErr))
				}
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.wantUser, gotUser)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestNewUserRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAppConfig := mocks.NewMockAppConfig(ctrl)
	mockAppConfig.EXPECT().GetConfig().Return(config.ServerCfg{
		PasswordSecret: "test-secret",
	}).AnyTimes()

	mockPasswordManager := mocks.NewMockPasswordManagerInterface(ctrl)

	type args struct {
		db          DBInterface
		cfg         AppConfig
		pswdManager PasswordManagerInterface
	}

	type testCase struct {
		name    string
		args    args
		want    *UserRepository
		wantErr bool
	}

	tests := []testCase{
		{
			name: "Success",
			args: args{
				db:  &MockDBWrapper{},
				cfg: mockAppConfig,
			},
			want: &UserRepository{
				db:  &MockDBWrapper{},
				cfg: mockAppConfig,
			},
			wantErr: false,
		},
		{
			name: "Nil DB",
			args: args{
				db:  nil,
				cfg: mockAppConfig,
			},
			want: &UserRepository{
				db:  nil,
				cfg: mockAppConfig,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewUserRepository(tt.args.db, tt.args.cfg, mockPasswordManager)
			assert.Equal(t, tt.want.db, got.db)
		})
	}
}

package services

import (
	"context"
	"errors"
	"testing"

	config "github.com/Maxim-Ba/information-keeper/config/server"
	"github.com/Maxim-Ba/information-keeper/internal/server/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthService_Login(t *testing.T) {
	t.Run("successful login", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockTokenService := new(MockTokenService)
		mockConfig := new(MockAppConfig)

		authService := New(mockUserRepo, mockTokenService, mockConfig)

		login, password := "testuser", "password123"
		expectedUser := &dto.UserRepoDTO{
			ID:    "1",
			Login: login,
			Email: "test@example.com",
		}
		expectedToken := &JWTToken{
			AcssToken:    "access_token",
			RefreshToken: "refresh_token",
		}

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			PasswordSecret: "test_secret",
		})

		mockUserRepo.On("Login", login, mock.AnythingOfType("string")).
			Return(expectedUser, nil)

		mockTokenService.On("GenerateToken", expectedUser).
			Return(expectedToken, nil)
		ctx := context.Background()
		result, err := authService.Login(ctx, login, password)

		assert.NoError(t, err)
		assert.Equal(t, expectedToken, result)
		mockUserRepo.AssertExpectations(t)
		mockTokenService.AssertExpectations(t)
		mockConfig.AssertExpectations(t)
	})

	t.Run("hash password error", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockTokenService := new(MockTokenService)
		mockConfig := new(MockAppConfig)

		authService := New(mockUserRepo, mockTokenService, mockConfig)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			PasswordSecret: "",
		})
		ctx := context.Background()

		result, err := authService.Login(ctx, "testuser", "password")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "HashPassword")
		mockConfig.AssertExpectations(t)
	})

	t.Run("user repository error", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockTokenService := new(MockTokenService)
		mockConfig := new(MockAppConfig)

		authService := New(mockUserRepo, mockTokenService, mockConfig)

		login, password := "testuser", "password123"
		expectedError := errors.New("user not found")

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			PasswordSecret: "test_secret",
		})

		mockUserRepo.On("Login", login, mock.AnythingOfType("string")).
			Return((*dto.UserRepoDTO)(nil), expectedError)
		ctx := context.Background()

		result, err := authService.Login(ctx, login, password)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "AuthService Login")
		mockUserRepo.AssertExpectations(t)
		mockConfig.AssertExpectations(t)
	})

	t.Run("token generation error", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockTokenService := new(MockTokenService)
		mockConfig := new(MockAppConfig)

		authService := New(mockUserRepo, mockTokenService, mockConfig)

		login, password := "testuser", "password123"
		expectedUser := &dto.UserRepoDTO{
			ID:    "1",
			Login: login,
			Email: "test@example.com",
		}
		expectedError := errors.New("token generation failed")

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			PasswordSecret: "test_secret",
		})

		mockUserRepo.On("Login", login, mock.AnythingOfType("string")).
			Return(expectedUser, nil)

		mockTokenService.On("GenerateToken", expectedUser).
			Return((*JWTToken)(nil), expectedError)
		ctx := context.Background()

		result, err := authService.Login(ctx, login, password)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "GenerateToken")
		mockUserRepo.AssertExpectations(t)
		mockTokenService.AssertExpectations(t)
		mockConfig.AssertExpectations(t)
	})

	t.Run("empty password", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockTokenService := new(MockTokenService)
		mockConfig := new(MockAppConfig)

		authService := New(mockUserRepo, mockTokenService, mockConfig)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			PasswordSecret: "test_secret",
		})
		ctx := context.Background()

		result, err := authService.Login(ctx, "testuser", "")

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

// Help function for testing
func setupTestAuthService() (*AuthService, *MockUserRepository, *MockTokenService, *MockAppConfig) {
	mockUserRepo := new(MockUserRepository)
	mockTokenService := new(MockTokenService)
	mockConfig := new(MockAppConfig)

	authService := New(mockUserRepo, mockTokenService, mockConfig)

	return authService, mockUserRepo, mockTokenService, mockConfig
}

func TestAuthService_Login_TableDriven(t *testing.T) {
	testCases := []struct {
		name          string
		login         string
		password      string
		setupMocks    func(*MockUserRepository, *MockTokenService, *MockAppConfig)
		expectedError bool
		errorContains string
	}{
		{
			name:     "invalid credentials",
			login:    "user",
			password: "wrongpass",
			setupMocks: func(repo *MockUserRepository, token *MockTokenService, cfg *MockAppConfig) {
				cfg.On("GetConfig").Return(config.ServerCfg{PasswordSecret: "secret"})
				repo.On("Login", "user", mock.AnythingOfType("string")).
					Return((*dto.UserRepoDTO)(nil), errors.New("invalid credentials"))
			},
			expectedError: true,
			errorContains: "AuthService Login",
		},
		{
			name:     "empty login",
			login:    "",
			password: "password",
			setupMocks: func(repo *MockUserRepository, token *MockTokenService, cfg *MockAppConfig) {
				cfg.On("GetConfig").Return(config.ServerCfg{PasswordSecret: "secret"})

			},
			expectedError: true,
			errorContains: "login cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authService, mockRepo, mockToken, mockConfig := setupTestAuthService()

			if tc.setupMocks != nil {
				tc.setupMocks(mockRepo, mockToken, mockConfig)
			}
			ctx := context.Background()

			result, err := authService.Login(ctx, tc.login, tc.password)

			if tc.expectedError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
			mockToken.AssertExpectations(t)
			mockConfig.AssertExpectations(t)
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	t.Run("successful token refresh", func(t *testing.T) {
		authService, _, mockToken, _ := setupTestAuthService()

		refreshToken := "valid_refresh_token"
		expectedToken := &JWTToken{
			AcssToken:    "new_access_token",
			RefreshToken: "new_refresh_token",
		}

		mockToken.On("RefreshToken", refreshToken).
			Return(expectedToken, nil)
		ctx := context.Background()

		result, err := authService.RefreshToken(ctx, refreshToken)

		assert.NoError(t, err)
		assert.Equal(t, expectedToken, result)
		mockToken.AssertExpectations(t)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		authService, _, mockToken, _ := setupTestAuthService()

		refreshToken := "invalid_refresh_token"
		expectedError := errors.New("invalid token")

		mockToken.On("RefreshToken", refreshToken).
			Return((*JWTToken)(nil), expectedError)
		ctx := context.Background()

		result, err := authService.RefreshToken(ctx, refreshToken)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "AuthService RefreshToken")
		mockToken.AssertExpectations(t)
	})

	t.Run("empty refresh token", func(t *testing.T) {
		authService, _, mockToken, _ := setupTestAuthService()

		refreshToken := ""
		expectedError := errors.New("token is empty")

		mockToken.On("RefreshToken", refreshToken).
			Return((*JWTToken)(nil), expectedError)
		ctx := context.Background()

		result, err := authService.RefreshToken(ctx, refreshToken)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockToken.AssertExpectations(t)
	})
}

func TestAuthService_RefreshToken_TableDriven(t *testing.T) {
	testCases := []struct {
		name           string
		refreshToken   string
		setupMocks     func(*MockTokenService)
		expectedError  bool
		errorContains  string
		expectedResult *JWTToken
	}{
		{
			name:         "successful refresh",
			refreshToken: "valid_token",
			setupMocks: func(token *MockTokenService) {
				token.On("RefreshToken", "valid_token").
					Return(&JWTToken{
						AcssToken:    "new_access",
						RefreshToken: "new_refresh",
					}, nil)
			},
			expectedError: false,
			expectedResult: &JWTToken{
				AcssToken:    "new_access",
				RefreshToken: "new_refresh",
			},
		},
		{
			name:         "expired token",
			refreshToken: "expired_token",
			setupMocks: func(token *MockTokenService) {
				token.On("RefreshToken", "expired_token").
					Return((*JWTToken)(nil), errors.New("token expired"))
			},
			expectedError: true,
			errorContains: "AuthService RefreshToken",
		},
		{
			name:         "malformed token",
			refreshToken: "malformed",
			setupMocks: func(token *MockTokenService) {
				token.On("RefreshToken", "malformed").
					Return((*JWTToken)(nil), errors.New("malformed token"))
			},
			expectedError: true,
			errorContains: "AuthService RefreshToken",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authService, _, mockToken, _ := setupTestAuthService()

			if tc.setupMocks != nil {
				tc.setupMocks(mockToken)
			}
			ctx := context.Background()

			result, err := authService.RefreshToken(ctx, tc.refreshToken)

			if tc.expectedError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			}

			mockToken.AssertExpectations(t)
		})
	}
}

func TestAuthService_Logout(t *testing.T) {
	t.Run("successful logout", func(t *testing.T) {
		authService, _, mockTokenService, _ := setupTestAuthService()

		token := &JWTToken{
			AcssToken:    "valid_access_token",
			RefreshToken: "valid_refresh_token",
		}

		mockTokenService.On("Remove", token).
			Return(nil)
		ctx := context.Background()

		err := authService.Logout(ctx, token)

		assert.NoError(t, err)
		mockTokenService.AssertExpectations(t)
	})

	t.Run("logout with repository error", func(t *testing.T) {
		authService, _, mockTokenService, _ := setupTestAuthService()

		token := &JWTToken{
			AcssToken:    "valid_access_token",
			RefreshToken: "valid_refresh_token",
		}
		expectedError := errors.New("logout failed")

		mockTokenService.On("Remove", token).
			Return(expectedError)
		ctx := context.Background()

		err := authService.Logout(ctx, token)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "AuthService Logout")
		mockTokenService.AssertExpectations(t)
	})

	t.Run("logout with empty token", func(t *testing.T) {
		authService, _, mockTokenService, _ := setupTestAuthService()

		token := &JWTToken{
			AcssToken:    "",
			RefreshToken: "",
		}
		expectedError := errors.New("token is empty")

		mockTokenService.On("Remove", token).
			Return(expectedError)
		ctx := context.Background()

		err := authService.Logout(ctx, token)

		assert.Error(t, err)
		mockTokenService.AssertExpectations(t)
	})
}

func TestAuthService_Logout_TableDriven(t *testing.T) {
	token := &JWTToken{
		AcssToken:    "valid_access_token",
		RefreshToken: "valid_refresh_token",
	}
	testCases := []struct {
		name          string
		accessToken   *JWTToken
		setupMocks    func(*MockTokenService)
		expectedError bool
		errorContains string
	}{
		{
			name:        "successful logout",
			accessToken: token,
			setupMocks: func(repo *MockTokenService) {
				repo.On("Remove", token).Return(nil)
			},
			expectedError: false,
		},
		{
			name:        "token not found",
			accessToken: token,
			setupMocks: func(repo *MockTokenService) {
				repo.On("Remove", token).
					Return(errors.New("token not found"))
			},
			expectedError: true,
			errorContains: "AuthService Logout",
		},
		{
			name:        "already logged out",
			accessToken: token,
			setupMocks: func(repo *MockTokenService) {
				repo.On("Remove", token).
					Return(errors.New("already logged out"))
			},
			expectedError: true,
			errorContains: "AuthService Logout",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authService, _, tokenService, _ := setupTestAuthService()

			if tc.setupMocks != nil {
				tc.setupMocks(tokenService)
			}
			ctx := context.Background()

			err := authService.Logout(ctx, tc.accessToken)

			if tc.expectedError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			tokenService.AssertExpectations(t)
		})
	}
}

func TestAuthService_Register(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		authService, mockRepo, mockToken, mockConfig := setupTestAuthService()

		expectedConfig := config.ServerCfg{
			PasswordSecret: "test-secret",
		}

		login, email, password := "testuser", "test@example.com", "validPassword123"
		expectedUser := &dto.UserRepoDTO{
			ID:    "1",
			Login: login,
			Email: email,
		}
		expectedToken := &JWTToken{
			AcssToken:    "access_token",
			RefreshToken: "refresh_token",
		}
		updatedUser := &dto.UserRepoDTO{
			ID:             "1",
			Login:          login,
			Email:          email,
			EmailConfirmed: true,
		}

		mockConfig.On("GetConfig").Return(expectedConfig)

		mockRepo.On("Register", login, email, mock.AnythingOfType("string")).
			Return(expectedUser, nil)

		mockToken.On("GenerateToken", expectedUser).
			Return(expectedToken, nil)

		mockRepo.On("GetUserByEmail", email).
			Return(expectedUser, nil)
		mockRepo.On("Update", mock.MatchedBy(func(user *dto.UserRepoDTO) bool {
			return user.EmailConfirmed == true
		})).Return(updatedUser, nil)
		ctx := context.Background()

		result, err := authService.Register(ctx, login, email, password)

		assert.NoError(t, err)
		assert.Equal(t, expectedToken, result)
		mockRepo.AssertExpectations(t)
		mockToken.AssertExpectations(t)
		mockConfig.AssertExpectations(t)
	})

	t.Run("user already exists", func(t *testing.T) {
		authService, mockRepo, _, mockConfig := setupTestAuthService()

		login, email, password := "existinguser", "existing@example.com", "password123"

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			PasswordSecret: "test_secret",
		})

		mockRepo.On("Register", login, email, mock.AnythingOfType("string")).
			Return((*dto.UserRepoDTO)(nil), errors.New("user already exists"))
		ctx := context.Background()

		result, err := authService.Register(ctx, login, email, password)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrUserAllreadyExists, err)
		mockRepo.AssertExpectations(t)
		mockConfig.AssertExpectations(t)
	})

	t.Run("password too short", func(t *testing.T) {
		authService, mockRepo, _, mockConfig := setupTestAuthService()

		login, email, password := "testuser", "test@example.com", "short"

		ctx := context.Background()

		result, err := authService.Register(ctx, login, email, password)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "TODO error")
		mockConfig.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "Register")
		mockRepo.AssertNotCalled(t, "GetUserByEmail")
		mockRepo.AssertNotCalled(t, "Update")
	})

	t.Run("repository error during registration", func(t *testing.T) {
		authService, mockRepo, _, mockConfig := setupTestAuthService()

		login, email, password := "testuser", "test@example.com", "validPassword123"
		expectedError := errors.New("database error")

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			PasswordSecret: "test_secret",
		})

		mockRepo.On("Register", login, email, mock.AnythingOfType("string")).
			Return((*dto.UserRepoDTO)(nil), expectedError)
		ctx := context.Background()

		result, err := authService.Register(ctx, login, email, password)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "AuthService Register")
		mockRepo.AssertExpectations(t)
		mockConfig.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "GetUserByEmail")
		mockRepo.AssertNotCalled(t, "Update")
	})

	t.Run("token generation error", func(t *testing.T) {
		authService, mockRepo, mockToken, mockConfig := setupTestAuthService()

		login, email, password := "testuser", "test@example.com", "validPassword123"
		expectedUser := &dto.UserRepoDTO{
			ID:    "1",
			Login: login,
			Email: email,
		}
		expectedError := errors.New("token generation failed")

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			PasswordSecret: "test_secret",
		})

		mockRepo.On("Register", login, email, mock.AnythingOfType("string")).
			Return(expectedUser, nil)

		mockToken.On("GenerateToken", expectedUser).
			Return((*JWTToken)(nil), expectedError)
		ctx := context.Background()

		result, err := authService.Register(ctx, login, email, password)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "GenerateToken")
		mockRepo.AssertExpectations(t)
		mockToken.AssertExpectations(t)
		mockConfig.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "GetUserByEmail")
		mockRepo.AssertNotCalled(t, "Update")
	})

	t.Run("error in SendEmailConfirmation", func(t *testing.T) {
		authService, mockRepo, mockToken, mockConfig := setupTestAuthService()

		login, email, password := "testuser", "test@example.com", "validPassword123"
		expectedUser := &dto.UserRepoDTO{
			ID:    "1",
			Login: login,
			Email: email,
		}
		expectedToken := &JWTToken{
			AcssToken:    "access_token",
			RefreshToken: "refresh_token",
		}

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			PasswordSecret: "test_secret",
		})

		mockRepo.On("Register", login, email, mock.AnythingOfType("string")).
			Return(expectedUser, nil)

		mockToken.On("GenerateToken", expectedUser).
			Return(expectedToken, nil)

		// Моки для SendEmailConfirmation с ошибкой
		mockRepo.On("GetUserByEmail", email).
			Return((*dto.UserRepoDTO)(nil), errors.New("email service unavailable"))
		ctx := context.Background()

		result, err := authService.Register(ctx, login, email, password)

		// Ошибка в SendEmailConfirmation не должна влиять на основной поток регистрации
		assert.NoError(t, err)
		assert.Equal(t, expectedToken, result)
		mockRepo.AssertExpectations(t)
		mockToken.AssertExpectations(t)
		mockConfig.AssertExpectations(t)
		// Update не должен вызываться если GetUserByEmail вернул ошибку
		mockRepo.AssertNotCalled(t, "Update")
	})
}
func TestAuthService_Register_TableDriven(t *testing.T) {
	testCases := []struct {
		name          string
		login         string
		email         string
		password      string
		setupMocks    func(*MockUserRepository, *MockTokenService, *MockAppConfig)
		expectedError bool
		errorContains string
		expectedErr   error
	}{
		{
			name:     "successful registration",
			login:    "newuser",
			email:    "new@example.com",
			password: "ValidPass123",
			setupMocks: func(repo *MockUserRepository, token *MockTokenService, cfg *MockAppConfig) {
				cfg.On("GetConfig").Return(config.ServerCfg{PasswordSecret: "secret"})
				user := &dto.UserRepoDTO{ID: "1", Login: "newuser", Email: "new@example.com"}
				updatedUser := &dto.UserRepoDTO{ID: "1", Login: "newuser", Email: "new@example.com", EmailConfirmed: true}

				repo.On("Register", "newuser", "new@example.com", mock.AnythingOfType("string")).
					Return(user, nil)
				token.On("GenerateToken", user).
					Return(&JWTToken{AcssToken: "access", RefreshToken: "refresh"}, nil)
				repo.On("GetUserByEmail", "new@example.com").Return(user, nil)
				repo.On("Update", mock.AnythingOfType("*dto.UserRepoDTO")).Return(updatedUser, nil)
			},
			expectedError: false,
		},
		{
			name:     "user already exists",
			login:    "existing",
			email:    "existing@example.com",
			password: "password123",
			setupMocks: func(repo *MockUserRepository, token *MockTokenService, cfg *MockAppConfig) {
				cfg.On("GetConfig").Return(config.ServerCfg{PasswordSecret: "secret"})
				repo.On("Register", "existing", "existing@example.com", mock.AnythingOfType("string")).
					Return((*dto.UserRepoDTO)(nil), errors.New("user already exists"))
			},
			expectedError: true,
			expectedErr:   ErrUserAllreadyExists,
		},
		{
			name:     "database error in registration",
			login:    "user",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func(repo *MockUserRepository, token *MockTokenService, cfg *MockAppConfig) {
				cfg.On("GetConfig").Return(config.ServerCfg{PasswordSecret: "secret"})
				repo.On("Register", "user", "test@example.com", mock.AnythingOfType("string")).
					Return((*dto.UserRepoDTO)(nil), errors.New("database connection failed"))
			},
			expectedError: true,
			errorContains: "AuthService Register",
		},
		{
			name:     "email confirmation fails but registration succeeds",
			login:    "user",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func(repo *MockUserRepository, token *MockTokenService, cfg *MockAppConfig) {
				cfg.On("GetConfig").Return(config.ServerCfg{PasswordSecret: "secret"})
				user := &dto.UserRepoDTO{ID: "1", Login: "user", Email: "test@example.com"}

				repo.On("Register", "user", "test@example.com", mock.AnythingOfType("string")).
					Return(user, nil)
				token.On("GenerateToken", user).
					Return(&JWTToken{AcssToken: "access", RefreshToken: "refresh"}, nil)
				repo.On("GetUserByEmail", "test@example.com").
					Return((*dto.UserRepoDTO)(nil), errors.New("email service down"))

			},
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authService, mockRepo, mockToken, mockConfig := setupTestAuthService()

			if tc.setupMocks != nil {
				tc.setupMocks(mockRepo, mockToken, mockConfig)
			}
			ctx := context.Background()

			result, err := authService.Register(ctx, tc.login, tc.email, tc.password)

			if tc.expectedError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
				if tc.expectedErr != nil {
					assert.Equal(t, tc.expectedErr, err)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
			mockToken.AssertExpectations(t)
			mockConfig.AssertExpectations(t)
		})
	}
}

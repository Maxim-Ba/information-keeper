package services

import (
	"context"
	"errors"
	"testing"
	"time"

	config "github.com/Maxim-Ba/information-keeper/config/server"
	"github.com/Maxim-Ba/information-keeper/internal/server/dto"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTokenService_AddToBlacklist(t *testing.T) {
	t.Run("successful add to blacklist", func(t *testing.T) {
		mockTokenRepo := new(MockTokenRepository)
		mockConfig := new(MockAppConfig)
		mockUserRepo := new(MockUserRepositoryInterface)

		tokenService := NewTokenService(mockConfig, mockUserRepo, mockTokenRepo)

		token := "test_token"
		expiry := time.Now().Add(time.Hour)

		mockTokenRepo.On("AddToBlacklist", token, expiry).
			Return(nil)

		err := tokenService.AddToBlacklist(context.Background(),token, expiry)

		assert.NoError(t, err)
		mockTokenRepo.AssertExpectations(t)
	})

	t.Run("token repository not configured", func(t *testing.T) {
		mockConfig := new(MockAppConfig)
		mockUserRepo := new(MockUserRepositoryInterface)

		// Создаем сервис без tokenRepository
		tokenService := &TokenService{
			config:         mockConfig,
			userRepository: mockUserRepo,
			// tokenRepository: nil - специально не устанавливаем
		}

		token := "test_token"
		expiry := time.Now().Add(time.Hour)

		err := tokenService.AddToBlacklist(context.Background(),token, expiry)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token repository not configured")
	})

	t.Run("repository error", func(t *testing.T) {
		mockTokenRepo := new(MockTokenRepository)
		mockConfig := new(MockAppConfig)
		mockUserRepo := new(MockUserRepositoryInterface)

		tokenService := NewTokenService(mockConfig, mockUserRepo, mockTokenRepo)

		token := "test_token"
		expiry := time.Now().Add(time.Hour)
		expectedError := errors.New("database error")

		mockTokenRepo.On("AddToBlacklist", token, expiry).
			Return(expectedError)

		err := tokenService.AddToBlacklist(context.Background(),token, expiry)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		mockTokenRepo.AssertExpectations(t)
	})
}

func TestTokenService_IsInBlacklist(t *testing.T) {
	t.Run("token is in blacklist", func(t *testing.T) {
		mockTokenRepo := new(MockTokenRepository)
		mockConfig := new(MockAppConfig)
		mockUserRepo := new(MockUserRepositoryInterface)

		tokenService := NewTokenService(mockConfig, mockUserRepo, mockTokenRepo)

		token := "blacklisted_token"

		mockTokenRepo.On("IsInBlacklist", token).
			Return(true, nil)

		result, err := tokenService.IsInBlacklist(context.Background(),token)

		assert.NoError(t, err)
		assert.True(t, result)
		mockTokenRepo.AssertExpectations(t)
	})

	t.Run("token is not in blacklist", func(t *testing.T) {
		mockTokenRepo := new(MockTokenRepository)
		mockConfig := new(MockAppConfig)
		mockUserRepo := new(MockUserRepositoryInterface)

		tokenService := NewTokenService(mockConfig, mockUserRepo, mockTokenRepo)

		token := "valid_token"

		mockTokenRepo.On("IsInBlacklist", token).
			Return(false, nil)

		result, err := tokenService.IsInBlacklist(context.Background(),token)

		assert.NoError(t, err)
		assert.False(t, result)
		mockTokenRepo.AssertExpectations(t)
	})

	t.Run("token repository not configured", func(t *testing.T) {
		mockConfig := new(MockAppConfig)
		mockUserRepo := new(MockUserRepositoryInterface)

		tokenService := &TokenService{
			config:         mockConfig,
			userRepository: mockUserRepo,
			// tokenRepository: nil - специально не устанавливаем
		}

		token := "test_token"

		result, err := tokenService.IsInBlacklist(context.Background(),token)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token repository not configured")
		assert.False(t, result)
	})

	t.Run("repository error", func(t *testing.T) {
		mockTokenRepo := new(MockTokenRepository)
		mockConfig := new(MockAppConfig)
		mockUserRepo := new(MockUserRepositoryInterface)

		tokenService := NewTokenService(mockConfig, mockUserRepo, mockTokenRepo)

		token := "test_token"
		expectedError := errors.New("connection failed")

		mockTokenRepo.On("IsInBlacklist", token).
			Return(false, expectedError)

		result, err := tokenService.IsInBlacklist(context.Background(),token)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.False(t, result)
		mockTokenRepo.AssertExpectations(t)
	})
}
func createTestToken(secret string, expiry time.Time) string {
	claims := CustomClaims{
		UserID: "test_user_id",
		Login:  "test_user",
		Email:  "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "test_user_id",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}
func TestTokenService_InvalidateToken(t *testing.T) {
	t.Run("successful token invalidation", func(t *testing.T) {
		mockTokenRepo := new(MockTokenRepository)
		mockConfig := new(MockAppConfig)
		mockUserRepo := new(MockUserRepositoryInterface)

		tokenService := NewTokenService(mockConfig, mockUserRepo, mockTokenRepo)
		secret := "test_secret"
		expiry := time.Now().Add(time.Hour)
		token := createTestToken(secret, expiry)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})

		// mock.AnythingOfType вместо точного времени
		mockTokenRepo.On("AddToBlacklist", token, mock.AnythingOfType("time.Time")).
			Return(nil)

		err := tokenService.InvalidateToken(context.Background(),token)

		assert.NoError(t, err)
		mockTokenRepo.AssertExpectations(t)
	})

	t.Run("blacklist error during invalidation", func(t *testing.T) {
		mockTokenRepo := new(MockTokenRepository)
		mockConfig := new(MockAppConfig)
		mockUserRepo := new(MockUserRepositoryInterface)

		tokenService := NewTokenService(mockConfig, mockUserRepo, mockTokenRepo)

		secret := "test_secret"
		expiry := time.Now().Add(time.Hour)
		validToken := createTestToken(secret, expiry)
		expectedError := errors.New("blacklist storage error")

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})

		mockTokenRepo.On("AddToBlacklist", validToken, mock.AnythingOfType("time.Time")).
			Return(expectedError)

		err := tokenService.InvalidateToken(context.Background(),validToken)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to add token to blacklist")
		mockTokenRepo.AssertExpectations(t)
	})

	t.Run("blacklist error during invalidation", func(t *testing.T) {
		mockTokenRepo := new(MockTokenRepository)
		mockConfig := new(MockAppConfig)
		mockUserRepo := new(MockUserRepositoryInterface)

		tokenService := NewTokenService(mockConfig, mockUserRepo, mockTokenRepo)

		secret := "test_secret"
		expiry := time.Now().Add(time.Hour)
		validToken := createTestToken(secret, expiry)
		expectedError := errors.New("blacklist storage error")

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})

		mockTokenRepo.On("AddToBlacklist", validToken, mock.AnythingOfType("time.Time")).
			Return(expectedError)

		err := tokenService.InvalidateToken(context.Background(),validToken)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to add token to blacklist")
	})
}

// Вспомогательные моки и функции

func setupTestTokenService() (*TokenService, *MockTokenRepository, *MockAppConfig, *MockUserRepositoryInterface) {
	mockTokenRepo := new(MockTokenRepository)
	mockConfig := new(MockAppConfig)
	mockUserRepo := new(MockUserRepositoryInterface)

	tokenService := NewTokenService(mockConfig, mockUserRepo, mockTokenRepo)

	return tokenService, mockTokenRepo, mockConfig, mockUserRepo
}

// Table-driven tests
func TestTokenService_AddToBlacklist_TableDriven(t *testing.T) {
	testCases := []struct {
		name          string
		token         string
		expiry        time.Time
		setupMocks    func(*MockTokenRepository)
		expectedError bool
		errorContains string
		expectedErr   error
	}{
		{
			name:   "successful addition",
			token:  "token123",
			expiry: time.Now().Add(time.Hour),
			setupMocks: func(repo *MockTokenRepository) {
				repo.On("AddToBlacklist", "token123", mock.AnythingOfType("time.Time")).
					Return(nil)
			},
			expectedError: false,
		},
		{
			name:   "database error",
			token:  "token456",
			expiry: time.Now().Add(time.Hour),
			setupMocks: func(repo *MockTokenRepository) {
				repo.On("AddToBlacklist", "token456", mock.AnythingOfType("time.Time")).
					Return(errors.New("database connection failed"))
			},
			expectedError: true,
			errorContains: "database connection failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokenService, mockRepo, _, _ := setupTestTokenService()

			if tc.setupMocks != nil {
				tc.setupMocks(mockRepo)
			}

			err := tokenService.AddToBlacklist(context.Background(),tc.token, tc.expiry)

			if tc.expectedError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
				if tc.expectedErr != nil {
					assert.Equal(t, tc.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTokenService_IsInBlacklist_TableDriven(t *testing.T) {
	testCases := []struct {
		name           string
		token          string
		setupMocks     func(*MockTokenRepository)
		expectedResult bool
		expectedError  bool
		errorContains  string
	}{
		{
			name:  "token in blacklist",
			token: "blacklisted_token",
			setupMocks: func(repo *MockTokenRepository) {
				repo.On("IsInBlacklist", "blacklisted_token").
					Return(true, nil)
			},
			expectedResult: true,
			expectedError:  false,
		},
		{
			name:  "token not in blacklist",
			token: "valid_token",
			setupMocks: func(repo *MockTokenRepository) {
				repo.On("IsInBlacklist", "valid_token").
					Return(false, nil)
			},
			expectedResult: false,
			expectedError:  false,
		},
		{
			name:  "repository error",
			token: "test_token",
			setupMocks: func(repo *MockTokenRepository) {
				repo.On("IsInBlacklist", "test_token").
					Return(false, errors.New("connection error"))
			},
			expectedResult: false,
			expectedError:  true,
			errorContains:  "connection error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokenService, mockRepo, _, _ := setupTestTokenService()

			if tc.setupMocks != nil {
				tc.setupMocks(mockRepo)
			}

			result, err := tokenService.IsInBlacklist(context.Background(),tc.token)

			if tc.expectedError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTokenService_ValidateTokenWithBlacklist(t *testing.T) {
	t.Run("valid token not in blacklist", func(t *testing.T) {
		tokenService, mockTokenRepo, mockConfig, _ := setupTestTokenService()

		secret := "test_secret"
		expiry := time.Now().Add(time.Hour)
		token := createTestToken(secret, expiry)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})
		mockTokenRepo.On("IsInBlacklist", token).Return(false, nil)

		err := tokenService.ValidateTokenWithBlacklist(context.Background(),token)

		assert.NoError(t, err)
		mockTokenRepo.AssertExpectations(t)
		mockConfig.AssertExpectations(t)
	})

	t.Run("token in blacklist", func(t *testing.T) {
		tokenService, mockTokenRepo, _, _ := setupTestTokenService()

		secret := "test_secret"
		expiry := time.Now().Add(time.Hour)
		token := createTestToken(secret, expiry)

		mockTokenRepo.On("IsInBlacklist", token).Return(true, nil)

		err := tokenService.ValidateTokenWithBlacklist(context.Background(),token)

		assert.Error(t, err)
		assert.Equal(t, ErrTokenRevoked, err)
		mockTokenRepo.AssertExpectations(t)
	})

	t.Run("blacklist check error", func(t *testing.T) {
		tokenService, mockTokenRepo, _, _ := setupTestTokenService()

		token := "test_token"
		expectedError := errors.New("database error")

		mockTokenRepo.On("IsInBlacklist", token).Return(false, expectedError)

		err := tokenService.ValidateTokenWithBlacklist(context.Background(),token)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check blacklist")
		mockTokenRepo.AssertExpectations(t)
	})

	t.Run("invalid token", func(t *testing.T) {
		tokenService, mockTokenRepo, mockConfig, _ := setupTestTokenService()

		invalidToken := "invalid.token.here"

		// mockConfig.On("GetConfig").Return(config.ServerCfg{
		// 	JWTSecret: "test_secret",
		// })
		mockTokenRepo.On("IsInBlacklist", invalidToken).Return(false, nil)

		err := tokenService.ValidateTokenWithBlacklist(context.Background(),invalidToken)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token validation failed")
		mockTokenRepo.AssertExpectations(t)
		mockConfig.AssertExpectations(t)
	})
}

func TestTokenService_ValidateRefreshTokenWithBlacklist(t *testing.T) {
	t.Run("valid refresh token not in blacklist", func(t *testing.T) {
		tokenService, mockTokenRepo, mockConfig, _ := setupTestTokenService()

		secret := "test_secret"
		expiry := time.Now().Add(time.Hour * 24 * 8)
		token := createTestToken(secret, expiry)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})
		mockTokenRepo.On("IsInBlacklist", token).Return(false, nil)

		err := tokenService.ValidateRefreshTokenWithBlacklist(context.Background(),token)

		assert.NoError(t, err)
		mockTokenRepo.AssertExpectations(t)
		mockConfig.AssertExpectations(t)
	})

	t.Run("refresh token in blacklist", func(t *testing.T) {
		tokenService, mockTokenRepo, _, _ := setupTestTokenService()

		token := "blacklisted_refresh_token"

		mockTokenRepo.On("IsInBlacklist", token).Return(true, nil)

		err := tokenService.ValidateRefreshTokenWithBlacklist(context.Background(),token)

		assert.Error(t, err)
		assert.Equal(t, ErrTokenRevoked, err)
		mockTokenRepo.AssertExpectations(t)
	})

	t.Run("blacklist check error for refresh token", func(t *testing.T) {
		tokenService, mockTokenRepo, _, _ := setupTestTokenService()

		token := "test_refresh_token"
		expectedError := errors.New("connection failed")

		mockTokenRepo.On("IsInBlacklist", token).Return(false, expectedError)

		err := tokenService.ValidateRefreshTokenWithBlacklist(context.Background(),token)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check blacklist")
		mockTokenRepo.AssertExpectations(t)
	})

	t.Run("invalid refresh token format", func(t *testing.T) {
		tokenService, mockTokenRepo, mockConfig, _ := setupTestTokenService()

		invalidToken := "invalid.refresh.token"

		mockTokenRepo.On("IsInBlacklist", invalidToken).Return(false, nil)

		err := tokenService.ValidateRefreshTokenWithBlacklist(context.Background(),invalidToken)

		assert.Error(t, err)
		mockTokenRepo.AssertExpectations(t)
		mockConfig.AssertExpectations(t)
	})
}

func TestTokenService_GenerateToken(t *testing.T) {
	t.Run("successful token generation", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		user := &dto.UserRepoDTO{
			ID:    "user123",
			Login: "testuser",
			Email: "test@example.com",
		}

		secret := "test_secret"
		// Ожидаем 4 вызова GetConfig: для генерации access, refresh и двух валидаций
		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		}).Times(4)

		tokens, err := tokenService.GenerateToken(user)

		assert.NoError(t, err)
		assert.NotNil(t, tokens)
		assert.NotEmpty(t, tokens.AcssToken)
		assert.NotEmpty(t, tokens.RefreshToken)

		// Проверяем, что токены валидны
		err = tokenService.ValidateToken(tokens.AcssToken)
		assert.NoError(t, err)

		err = tokenService.ValidateRefreshToken(tokens.RefreshToken)
		assert.NoError(t, err)

		mockConfig.AssertExpectations(t)
	})

	t.Run("refresh token contains only user ID", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		user := &dto.UserRepoDTO{
			ID:    "user123",
			Login: "testuser",
			Email: "test@example.com",
		}

		secret := "test_secret"
		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		}).Times(3)

		tokens, err := tokenService.GenerateToken(user)
		assert.NoError(t, err)

		userFromRefresh, err := tokenService.GetUserFromRefreshToken(tokens.RefreshToken)
		assert.NoError(t, err)
		assert.Equal(t, user.ID, userFromRefresh.ID)
		// Login и Email будут пустыми для refresh token
		assert.Empty(t, userFromRefresh.Login)
		assert.Empty(t, userFromRefresh.Email)

		mockConfig.AssertExpectations(t)
	})

	t.Run("tokens have correct expiration times", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		user := &dto.UserRepoDTO{
			ID:    "user123",
			Login: "testuser",
			Email: "test@example.com",
		}

		secret := "test_secret"
		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		}).Times(2) // Только для генерации

		startTime := time.Now()
		tokens, err := tokenService.GenerateToken(user)
		assert.NoError(t, err)

		accessToken, _ := jwt.ParseWithClaims(tokens.AcssToken, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		accessClaims := accessToken.Claims.(*CustomClaims)

		// Access token должен истечь через ~15 минут
		expectedAccessExpiry := startTime.Add(15 * time.Minute)
		assert.WithinDuration(t, expectedAccessExpiry, accessClaims.ExpiresAt.Time, time.Second*5)

		refreshToken, _ := jwt.ParseWithClaims(tokens.RefreshToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		refreshClaims := refreshToken.Claims.(*jwt.RegisteredClaims)

		// Refresh token должен истечь через ~7 дней
		expectedRefreshExpiry := startTime.Add(7 * 24 * time.Hour)
		assert.WithinDuration(t, expectedRefreshExpiry, refreshClaims.ExpiresAt.Time, time.Second*5)

		mockConfig.AssertExpectations(t)
	})
}

func TestTokenService_RefreshToken(t *testing.T) {
	t.Run("successful token refresh", func(t *testing.T) {
		tokenService, mockTokenRepo, mockConfig, mockUserRepo := setupTestTokenService()

		secret := "test_secret"
		expiry := time.Now().Add(time.Hour * 24 * 8) // 8 дней для refresh token
		refreshToken := createTestToken(secret, expiry)

		user := &dto.UserRepoDTO{
			ID:    "user123",
			Login: "testuser",
			Email: "test@example.com",
		}

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})
		mockTokenRepo.On("IsInBlacklist", refreshToken).Return(false, nil)
		mockTokenRepo.On("AddToBlacklist", refreshToken, mock.AnythingOfType("time.Time")).Return(nil)
		mockUserRepo.On("GetUserByID", "test_user_id").Return(user, nil)

		newTokens, err := tokenService.RefreshToken(context.Background(),refreshToken)

		assert.NoError(t, err)
		assert.NotNil(t, newTokens)
		assert.NotEmpty(t, newTokens.AcssToken)
		assert.NotEmpty(t, newTokens.RefreshToken)

		mockTokenRepo.AssertExpectations(t)
		mockConfig.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("refresh token in blacklist", func(t *testing.T) {
		tokenService, mockTokenRepo, _, _ := setupTestTokenService()

		refreshToken := "blacklisted_token"

		mockTokenRepo.On("IsInBlacklist", refreshToken).Return(true, nil)

		newTokens, err := tokenService.RefreshToken(context.Background(),refreshToken)

		assert.Error(t, err)
		assert.Nil(t, newTokens)
		assert.Equal(t, ErrTokenRevoked, err)

		mockTokenRepo.AssertExpectations(t)
	})

	t.Run("blacklist check error", func(t *testing.T) {
		tokenService, mockTokenRepo, _, _ := setupTestTokenService()

		refreshToken := "test_token"
		expectedError := errors.New("database error")

		mockTokenRepo.On("IsInBlacklist", refreshToken).Return(false, expectedError)

		newTokens, err := tokenService.RefreshToken(context.Background(),refreshToken)

		assert.Error(t, err)
		assert.Nil(t, newTokens)
		assert.Contains(t, err.Error(), "failed to check blacklist")

		mockTokenRepo.AssertExpectations(t)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		tokenService, mockTokenRepo, mockConfig, _ := setupTestTokenService()

		invalidToken := "invalid.token.here"

		mockTokenRepo.On("IsInBlacklist", invalidToken).Return(false, nil)

		newTokens, err := tokenService.RefreshToken(context.Background(),invalidToken)

		assert.Error(t, err)
		assert.Nil(t, newTokens)
		assert.Contains(t, err.Error(), "invalid refresh token")

		mockTokenRepo.AssertExpectations(t)
		mockConfig.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		tokenService, mockTokenRepo, mockConfig, mockUserRepo := setupTestTokenService()

		secret := "test_secret"
		expiry := time.Now().Add(time.Hour * 24 * 8)
		refreshToken := createTestToken(secret, expiry)
		expectedError := errors.New("user not found")

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})

		mockTokenRepo.On("IsInBlacklist", refreshToken).Return(false, nil)
		mockTokenRepo.On("AddToBlacklist", refreshToken, mock.AnythingOfType("time.Time")).Return(nil)
		mockUserRepo.On("GetUserByID", "test_user_id").Return(nil, expectedError)

		newTokens, err := tokenService.RefreshToken(context.Background(),refreshToken)

		assert.Error(t, err)
		assert.Nil(t, newTokens)
		assert.Contains(t, err.Error(), "user not found")

		mockTokenRepo.AssertExpectations(t)
		mockConfig.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("failed to blacklist used token", func(t *testing.T) {
		tokenService, mockTokenRepo, mockConfig, mockUserRepo := setupTestTokenService()

		secret := "test_secret"
		expiry := time.Now().Add(time.Hour * 24 * 8)
		refreshToken := createTestToken(secret, expiry)

		blacklistError := errors.New("blacklist error")

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})

		mockTokenRepo.On("IsInBlacklist", refreshToken).Return(false, nil)
		mockTokenRepo.On("AddToBlacklist", refreshToken, mock.AnythingOfType("time.Time")).Return(blacklistError)

		newTokens, err := tokenService.RefreshToken(context.Background(),refreshToken)

		assert.Error(t, err)
		assert.Nil(t, newTokens)
		assert.Contains(t, err.Error(), "failed to blacklist used refresh token")

		mockTokenRepo.AssertExpectations(t)
		mockConfig.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestTokenService_ValidateToken(t *testing.T) {
	t.Run("valid access token", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		secret := "test_secret"
		expiry := time.Now().Add(time.Hour) // 1 час для access token
		token := createTestToken(secret, expiry)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})

		err := tokenService.ValidateToken(token)

		assert.NoError(t, err)
		mockConfig.AssertExpectations(t)
	})

	t.Run("invalid token signature", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		// Токен подписан другим секретом
		wrongSecret := "wrong_secret"
		expiry := time.Now().Add(time.Hour)
		token := createTestToken(wrongSecret, expiry)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: "correct_secret",
		})

		err := tokenService.ValidateToken(token)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token validation failed")
		mockConfig.AssertExpectations(t)
	})

	t.Run("expired token", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		secret := "test_secret"
		expiry := time.Now().Add(-time.Hour) // Токен истек час назад
		token := createTestToken(secret, expiry)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})

		err := tokenService.ValidateToken(token)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token validation failed")
		mockConfig.AssertExpectations(t)
	})

	t.Run("malformed token", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		malformedToken := "not.a.valid.jwt.token"

		// mockConfig.On("GetConfig").Return(config.ServerCfg{
		// 	JWTSecret: "test_secret",
		// })

		err := tokenService.ValidateToken(malformedToken)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token validation failed")
		mockConfig.AssertExpectations(t)
	})

	t.Run("access token with refresh token expiration", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		secret := "test_secret"
		expiry := time.Now().Add(time.Hour * 24 * 8) // 8 дней - как у refresh token
		token := createTestToken(secret, expiry)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})

		err := tokenService.ValidateToken(token)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token type")
		mockConfig.AssertExpectations(t)
	})
}

func TestTokenService_ValidateRefreshToken(t *testing.T) {
	t.Run("valid refresh token", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		secret := "test_secret"
		expiry := time.Now().Add(time.Hour * 24 * 8) // 8 дней для refresh token
		token := createTestToken(secret, expiry)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})

		err := tokenService.ValidateRefreshToken(token)

		assert.NoError(t, err)
		mockConfig.AssertExpectations(t)
	})

	t.Run("refresh token with access token expiration", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		secret := "test_secret"
		expiry := time.Now().Add(time.Hour) // 1 час - как у access token
		token := createTestToken(secret, expiry)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})

		err := tokenService.ValidateRefreshToken(token)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token type")
		mockConfig.AssertExpectations(t)
	})

	t.Run("expired refresh token", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		secret := "test_secret"
		expiry := time.Now().Add(-time.Hour) // Токен истек час назад
		token := createTestToken(secret, expiry)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})

		err := tokenService.ValidateRefreshToken(token)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token validation failed")
		mockConfig.AssertExpectations(t)
	})
}

func TestTokenService_GetUserFromRefreshToken(t *testing.T) {
	t.Run("successful user extraction from refresh token", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		secret := "test_secret"
		expiry := time.Now().Add(time.Hour * 24 * 8)
		token := createTestToken(secret, expiry)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})

		user, err := tokenService.GetUserFromRefreshToken(token)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "test_user_id", user.ID)
		assert.Equal(t, "test_user", user.Login)
		assert.Equal(t, "test@example.com", user.Email)

		mockConfig.AssertExpectations(t)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		invalidToken := "invalid.token.here"

		user, err := tokenService.GetUserFromRefreshToken(invalidToken)

		assert.Error(t, err)
		assert.Nil(t, user)

		mockConfig.AssertExpectations(t)
	})

	t.Run("refresh token with access token expiration", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		secret := "test_secret"
		expiry := time.Now().Add(time.Hour) // 1 час - как у access token
		token := createTestToken(secret, expiry)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})

		user, err := tokenService.GetUserFromRefreshToken(token)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "invalid token type")

		mockConfig.AssertExpectations(t)
	})
}

func TestTokenService_GetUserFromAcssToken(t *testing.T) {
	t.Run("successful user extraction from access token", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		secret := "test_secret"
		expiry := time.Now().Add(time.Hour)
		token := createTestToken(secret, expiry)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})

		user, err := tokenService.GetUserFromAcssToken(token)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "test_user_id", user.ID)
		assert.Equal(t, "test_user", user.Login)
		assert.Equal(t, "test@example.com", user.Email)

		mockConfig.AssertExpectations(t)
	})

	t.Run("invalid access token", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		invalidToken := "invalid.token.here"

		user, err := tokenService.GetUserFromAcssToken(invalidToken)

		assert.Error(t, err)
		assert.Nil(t, user)

		mockConfig.AssertExpectations(t)
	})

	t.Run("access token with refresh token expiration", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		secret := "test_secret"
		expiry := time.Now().Add(time.Hour * 24 * 8) // 8 дней - как у refresh token
		token := createTestToken(secret, expiry)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})

		user, err := tokenService.GetUserFromAcssToken(token)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "invalid token type")

		mockConfig.AssertExpectations(t)
	})

	t.Run("expired access token", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		secret := "test_secret"
		expiry := time.Now().Add(-time.Hour) // Токен истек час назад
		token := createTestToken(secret, expiry)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})

		user, err := tokenService.GetUserFromAcssToken(token)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "token validation failed")

		mockConfig.AssertExpectations(t)
	})
}

func TestTokenService_validateToken(t *testing.T) {
	t.Run("valid access token", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		secret := "test_secret"
		expiry := time.Now().Add(15 * time.Minute) // 15 минут для access token
		token := createTestToken(secret, expiry)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})

		claims, err := tokenService.validateToken(token, false)

		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, "test_user_id", claims.UserID)
		assert.Equal(t, "test_user", claims.Login)
		assert.Equal(t, "test@example.com", claims.Email)
		mockConfig.AssertExpectations(t)
	})

	t.Run("valid refresh token", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		secret := "test_secret"
		expiry := time.Now().Add(8 * 24 * time.Hour) // 8 дней для refresh token
		token := createTestToken(secret, expiry)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})

		claims, err := tokenService.validateToken(token, true)

		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, "test_user_id", claims.UserID)
		mockConfig.AssertExpectations(t)
	})

	t.Run("invalid token signature", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		// Токен подписан другим секретом
		wrongSecret := "wrong_secret"
		expiry := time.Now().Add(time.Hour)
		token := createTestToken(wrongSecret, expiry)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: "correct_secret",
		})

		claims, err := tokenService.validateToken(token, false)

		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "token validation failed")
		mockConfig.AssertExpectations(t)
	})

	t.Run("expired token", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		secret := "test_secret"
		expiry := time.Now().Add(-time.Hour) // Токен истек час назад
		token := createTestToken(secret, expiry)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})

		claims, err := tokenService.validateToken(token, false)

		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "token validation failed")
		mockConfig.AssertExpectations(t)
	})

	t.Run("malformed token", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		malformedToken := "not.a.valid.jwt.token"

		claims, err := tokenService.validateToken(malformedToken, false)

		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "token validation failed")
		mockConfig.AssertExpectations(t)
	})

	t.Run("empty token", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		claims, err := tokenService.validateToken("", false)

		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "token validation failed")
		mockConfig.AssertExpectations(t)
	})

	t.Run("access token with refresh token expiration", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		secret := "test_secret"
		expiry := time.Now().Add(8 * 24 * time.Hour) // 8 дней - как у refresh token
		token := createTestToken(secret, expiry)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})

		claims, err := tokenService.validateToken(token, false) // Проверяем как access token

		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "invalid token type")
		assert.Contains(t, err.Error(), "access token should have shorter expiration")
		mockConfig.AssertExpectations(t)
	})

	t.Run("refresh token with access token expiration", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		secret := "test_secret"
		expiry := time.Now().Add(time.Hour) // 1 час - как у access token
		token := createTestToken(secret, expiry)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})

		claims, err := tokenService.validateToken(token, true) // Проверяем как refresh token

		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "invalid token type")
		assert.Contains(t, err.Error(), "refresh token should have longer expiration")
		mockConfig.AssertExpectations(t)
	})

	t.Run("token with 23 hours expiration for access token", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		secret := "test_secret"
		expiry := time.Now().Add(23 * time.Hour)
		token := createTestToken(secret, expiry)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})

		claims, err := tokenService.validateToken(token, false)

		assert.NoError(t, err)
		assert.NotNil(t, claims)
		mockConfig.AssertExpectations(t)
	})

	t.Run("token with 25 hours expiration for refresh token", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		secret := "test_secret"
		expiry := time.Now().Add(25 * time.Hour)
		token := createTestToken(secret, expiry)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})

		claims, err := tokenService.validateToken(token, true)

		assert.NoError(t, err)
		assert.NotNil(t, claims)
		mockConfig.AssertExpectations(t)
	})

	t.Run("token with exactly 24 hours expiration for refresh token check", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		secret := "test_secret"
		expiry := time.Now().Add(24 * time.Hour) // Ровно 24 часа
		token := createTestToken(secret, expiry)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})

		// Для refresh token должно быть ошибкой (должно быть > 24 часов)
		claims, err := tokenService.validateToken(token, true)

		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "refresh token should have longer expiration")
		mockConfig.AssertExpectations(t)
	})

	t.Run("token with 23 hours expiration for access token", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		secret := "test_secret"
		expiry := time.Now().Add(23 * time.Hour) // 23 часа - допустимо для access token
		token := createTestToken(secret, expiry)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})

		claims, err := tokenService.validateToken(token, false)

		assert.NoError(t, err)
		assert.NotNil(t, claims)
		mockConfig.AssertExpectations(t)
	})

	t.Run("token with 25 hours expiration for refresh token", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		secret := "test_secret"
		expiry := time.Now().Add(25 * time.Hour) // 25 часов - допустимо для refresh token
		token := createTestToken(secret, expiry)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})

		claims, err := tokenService.validateToken(token, true)

		assert.NoError(t, err)
		assert.NotNil(t, claims)
		mockConfig.AssertExpectations(t)
	})
}

func TestTokenService_validateToken_EdgeCases(t *testing.T) {
	t.Run("token with negative expiration", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		secret := "test_secret"
		expiry := time.Now().Add(-48 * time.Hour) // Истек 2 дня назад
		token := createTestToken(secret, expiry)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})

		claims, err := tokenService.validateToken(token, false)

		assert.Error(t, err)
		assert.Nil(t, claims)
		mockConfig.AssertExpectations(t)
	})

	t.Run("token with very long expiration for refresh token", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		secret := "test_secret"
		expiry := time.Now().Add(365 * 24 * time.Hour) // 1 год
		token := createTestToken(secret, expiry)

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: secret,
		})

		claims, err := tokenService.validateToken(token, true)

		assert.NoError(t, err)
		assert.NotNil(t, claims)
		mockConfig.AssertExpectations(t)
	})

	t.Run("nil config", func(t *testing.T) {
		tokenService := &TokenService{
			config:          nil,
			userRepository:  nil,
			tokenRepository: nil,
		}

		claims, err := tokenService.validateToken("test_token", false)

		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("config with empty secret", func(t *testing.T) {
		tokenService, _, mockConfig, _ := setupTestTokenService()

		mockConfig.On("GetConfig").Return(config.ServerCfg{
			JWTSecret: "", // Пустой секрет
		})

		token := createTestToken("some_secret", time.Now().Add(time.Hour))
		claims, err := tokenService.validateToken(token, false)

		assert.Error(t, err)
		assert.Nil(t, claims)
		mockConfig.AssertExpectations(t)
	})
}

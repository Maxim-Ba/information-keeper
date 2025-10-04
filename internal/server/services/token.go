package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc/metadata"

	"github.com/Maxim-Ba/information-keeper/internal/server/dto"
	"github.com/Maxim-Ba/information-keeper/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
)

type TokenRepositoryInterface interface {
	AddToBlacklist(ctx context.Context, token string, expiry time.Time) error
	IsInBlacklist(ctx context.Context, token string) (bool, error)
}

type TokenService struct {
	config          AppConfig
	userRepository  UserRepositoryInterface
	tokenRepository TokenRepositoryInterface
}

// Remove implements TokenServiceInterface.
func (s *TokenService) Remove(ctx context.Context,token *JWTToken) error {
	if err := s.InvalidateRefreshToken(ctx,token.RefreshToken); err != nil {
		return fmt.Errorf("failed to invalidate refresh token: %w", err)
	}
	if err := s.InvalidateToken(ctx,token.AcssToken, ); err != nil {
		return fmt.Errorf("failed to invalidate token: %w", err)
	}
	return nil
}

type UserRepositoryInterface interface {
	GetUserByID(ctx context.Context,id string) (*dto.UserRepoDTO, error)
}

type CustomClaims struct {
	UserID  string `json:"user_id"`
	Login   string `json:"login"`
	Email   string `json:"email"`
	TokenID string `json:"token_id"`
	jwt.RegisteredClaims
}

func NewTokenService(
	config AppConfig,
	userRepository UserRepositoryInterface,
	tokenRepository TokenRepositoryInterface,
) *TokenService {
	return &TokenService{
		config: config, userRepository: userRepository, tokenRepository: tokenRepository}
}


func (s *TokenService) AddToBlacklist(ctx context.Context, token string, expiry time.Time) error {
	if s.tokenRepository == nil {
		return fmt.Errorf("token repository not configured")
	}

	return s.tokenRepository.AddToBlacklist(ctx , token, expiry)
}
func (s *TokenService) IsInBlacklist(ctx context.Context,token string) (bool, error) {
	if s.tokenRepository == nil {
		return false, fmt.Errorf("token repository not configured")
	}

	return s.tokenRepository.IsInBlacklist(ctx , token)
}

// Add token to blacklist
func (s *TokenService) InvalidateToken(ctx context.Context,token string,) error {
	logger.Info("TokenService InvalidateToken token:" + fmt.Sprintf("%v", token))
	// Парсим токен, чтобы получить время истечения
	claims, err := s.validateToken(token, false)
	if err != nil {
		return fmt.Errorf("failed to parse token for invalidation: %w", err)
	}

	// Добавляем в черный список через репозиторий
	if err := s.AddToBlacklist(ctx,token, claims.ExpiresAt.Time); err != nil {
		return fmt.Errorf("failed to add token to blacklist: %w", err)
	}

	return nil
}
func (s *TokenService) InvalidateRefreshToken(ctx context.Context,token string) error {
	logger.Info("TokenService InvalidateToken token:" + fmt.Sprintf("%v", token))
	// Парсим токен, чтобы получить время истечения
	claims, err := s.validateToken(token, true)
	if err != nil {
		return fmt.Errorf("failed to parse token for invalidation: %w", err)
	}

	// Добавляем в черный список через репозиторий
	if err := s.AddToBlacklist(ctx,token, claims.ExpiresAt.Time); err != nil {
		return fmt.Errorf("failed to add token to blacklist: %w", err)
	}

	return nil
}
// ValidateTokenWithBlacklist checks if token is valid and not in blacklist
func (s *TokenService) ValidateTokenWithBlacklist(ctx context.Context,token string) error {
	// Сначала проверяем в черном списке
	inBlacklist, err := s.IsInBlacklist(ctx,token)
	if err != nil {
		return fmt.Errorf("failed to check blacklist: %w", err)
	}
	if inBlacklist {
		return ErrTokenRevoked
	}

	// Затем валидируем сам токен
	return s.ValidateToken(token)
}

// ValidateRefreshTokenWithBlacklist check refresh token and not in blacklist
func (s *TokenService) ValidateRefreshTokenWithBlacklist(ctx context.Context,token string) error {
	inBlacklist, err := s.IsInBlacklist(ctx,token)
	if err != nil {
		return fmt.Errorf("failed to check blacklist: %w", err)
	}
	if inBlacklist {
		return ErrTokenRevoked
	}

	// Затем валидируем сам токен
	return s.ValidateRefreshToken(token)
}
func (s *TokenService) GenerateToken(user *dto.UserRepoDTO) (*JWTToken, error) {
	// Access token
	accessClaims := CustomClaims{
		UserID: user.ID,
		Login:  user.Login,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.config.GetConfig().JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Refresh token
	refreshClaims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Subject:   user.ID,
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.config.GetConfig().JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &JWTToken{
		AcssToken:    accessTokenString,
		RefreshToken: refreshTokenString,
	}, nil
}

func (s *TokenService) RefreshToken(ctx context.Context,refreshToken string) (*JWTToken, error) {
	inBlacklist, err := s.IsInBlacklist(ctx,refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to check blacklist: %w", err)
	}
	if inBlacklist {
		return nil, ErrTokenRevoked
	}

	// Validate refresh token
	claims, err := s.validateToken(refreshToken, true)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Добавляем использованный refresh token в черный список через репозиторий
	if err := s.AddToBlacklist(ctx,refreshToken, claims.ExpiresAt.Time); err != nil {
		return nil, fmt.Errorf("failed to blacklist used refresh token: %w", err)
	}

	user, err := s.userRepository.GetUserByID(context.TODO(),claims.Subject)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return s.GenerateToken(user)
}

func (s *TokenService) ValidateToken(token string) error {
	_, err := s.validateToken(token, false)
	return err
}

func (s *TokenService) ValidateRefreshToken(token string) error {
	_, err := s.validateToken(token, true)
	return err
}

func (s *TokenService) GetUserFromRefreshToken(token string) (*dto.UserRepoDTO, error) {
	claims, err := s.validateToken(token, true)
	if err != nil {
		return nil, err
	}

	return &dto.UserRepoDTO{
		ID:    claims.Subject,
		Login: claims.Login,
		Email: claims.Email,
	}, nil
}

func (s *TokenService) GetUserFromAcssToken(token string) (*dto.UserRepoDTO, error) {
	claims, err := s.validateToken(token, false)
	if err != nil {
		return nil, err
	}

	return &dto.UserRepoDTO{
		ID:    claims.Subject,
		Login: claims.Login,
		Email: claims.Email,
	}, nil
}

func (s *TokenService) validateToken(tokenString string, isRefresh bool) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.GetConfig().JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		// Additional validation for token type
		if isRefresh && time.Until(claims.ExpiresAt.Time) <= time.Hour*24 {
			return nil, fmt.Errorf("invalid token type: refresh token should have longer expiration")
		}
		// Access token should have shorter expiration (less than or equal to 24 hours)
		if !isRefresh && time.Until(claims.ExpiresAt.Time) > time.Hour*24 {
			return nil, fmt.Errorf("invalid token type: access token should have shorter expiration")
		}
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

func (s *TokenService)  GetAccessTokenFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return "", ErrMetaDataNotProvided
		}
		tokens := md["authorization"]
		logger.Info(fmt.Sprintf("AuthInterceptor tokens: %v", tokens))
		if len(tokens) == 0 {
			return "", ErrAuthTokenNotProvided
		}
		token := strings.TrimPrefix(tokens[0], "Bearer ")
		if err := s.ValidateTokenWithBlacklist(ctx, token); err != nil {
			logger.Error(fmt.Sprintf("AuthInterceptor failed to validate token: %v", err))
			return "", ErrInvalidToken
		}
		return token, nil
}

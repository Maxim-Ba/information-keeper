package services

import (
	"context"
	"fmt"

	config "github.com/Maxim-Ba/information-keeper/config/server"
	"github.com/Maxim-Ba/information-keeper/internal/server/dto"
	"github.com/Maxim-Ba/information-keeper/pkg/utils"
)

type UserReader interface {
	GetUserByEmail(ctx context.Context, email string) (*dto.UserRepoDTO, error)
}

type UserWriter interface {
	Update(ctx context.Context, user *dto.UserRepoDTO) (*dto.UserRepoDTO, error)
	Register(ctx context.Context, login, email, password string) (*dto.UserRepoDTO, error)
}
type PasswordManager interface {
	ChangePassword(ctx context.Context, login, oldPassword, newPassword string) error
}
type UserRepository interface {
	Login(ctx context.Context, login, password string) (*dto.UserRepoDTO, error)
	PasswordManager
	UserReader
	UserWriter
}
type TokenServiceInterface interface {
	RefreshToken(token string) (*JWTToken, error)
	GenerateToken(*dto.UserRepoDTO) (*JWTToken, error)
	Remove(token *JWTToken) error
}

type AuthService struct {
	userRepository UserRepository
	tokenService   TokenServiceInterface
	config         AppConfig
	// TODO EmailService
}

type AppConfig interface {
	GetConfig() config.ServerCfg
}

type JWTToken struct {
	AcssToken    string
	RefreshToken string
}

func New(userRepository UserRepository, tokenService TokenServiceInterface, config AppConfig) *AuthService {
	return &AuthService{userRepository: userRepository, tokenService: tokenService, config: config}
}

func (s *AuthService) Login(ctx context.Context, login, password string) (*JWTToken, error) {
	if login == "" {
		return nil, fmt.Errorf("login cannot be empty")
	}
	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	user, err := s.userRepository.Login(ctx, login, password)
	if err != nil {
		return nil, fmt.Errorf("AuthService Login: %w", err)
	}

	jwt, err := s.tokenService.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("AuthService Login GenerateToken: %w", err)
	}
	return jwt, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*JWTToken, error) {
	jwt, err := s.tokenService.RefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("AuthService RefreshToken: %w", err)
	}
	return jwt, nil
}

func (s *AuthService) Logout(ctx context.Context, token *JWTToken) error {
	if err := s.tokenService.Remove(token); err != nil {
		return fmt.Errorf("AuthService Logout: %w", err)
	}
	return nil
}

func (s *AuthService) Register(ctx context.Context, login, email, password string) (*JWTToken, error) {

	if err := s.validatePassword(password); err != nil {
		return nil, err
	}
	pm := &utils.PasswordManager{}
	passwordHash, err := pm.HashPassword(password, s.config.GetConfig().PasswordSecret)
	if err != nil {
		return nil, fmt.Errorf("AuthService Register HashPassword: %w", err)
	}
	user, err := s.userRepository.Register(ctx, login, email, passwordHash)
	if err != nil {
		// TODO standart error codes
		if err.Error() == "user already exists" {
			return nil, ErrUserAllreadyExists
		}
		return nil, fmt.Errorf("AuthService Register: %w", err)
	}
	jwt, err := s.tokenService.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("AuthService Register GenerateToken: %w", err)
	}
	s.SendEmailConfirmation(ctx, user.Email)
	return jwt, nil

}

func (s *AuthService) ChangePassword(ctx context.Context, data *dto.ChangePassword) error {

	if err := s.validatePassword(data.NewPassword); err != nil {
		return err
	}

	if err := s.userRepository.ChangePassword(ctx, data.Login, data.OldPassword, data.NewPassword); err != nil {
		return fmt.Errorf("AuthService ChangePassword: %w", err)
	}
	return nil

}
func (s *AuthService) RestorePassword(ctx context.Context, email string) error {
	_, err := s.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("AuthService RestorePassword GetUserByEmail: %w", err)
	}
	// TODO Send Email
	// TODO html with restore password
	return nil
}
func (s *AuthService) SendEmailConfirmation(ctx context.Context, email string) error {
	user, err := s.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("AuthService GetUserByEmail: %w", err)
	}
	if user == nil {
		return fmt.Errorf("AuthService GetUserByEmail: %w", err)

	}
	// TODO send email
	//TODO Change EmailConfirmed method
	user.EmailConfirmed = true
	user, err = s.userRepository.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("AuthService Update: %w", err)
	}
	return nil
}

func (s *AuthService) validatePassword(password string) error {
	if len(password) < 8 {
		// TODO error
		return fmt.Errorf("TODO error")
	}
	return nil
}

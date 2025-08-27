package services

import (
	"fmt"

	config "github.com/Maxim-Ba/information-keeper/config/server"
	"github.com/Maxim-Ba/information-keeper/internal/server/dto"
	"github.com/Maxim-Ba/information-keeper/pkg/utils"
)

type UserRepository interface {
	Register(login, email, password string) (*dto.UserRepoDTO, error)
	Login(login, password string) (*dto.UserRepoDTO, error)
	ChangePassword(login, oldPassword, newPassword string) error
	RestorePassword(email string) error
	Logout(acssToken string) error
	GetUserByEmail(email string) (*dto.UserRepoDTO, error)
	Update(user *dto.UserRepoDTO) (*dto.UserRepoDTO, error)
}
type TokenServiceInterface interface {
	RefreshToken(token string) (*JWTToken, error)
	GenerateToken(*dto.UserRepoDTO) (*JWTToken, error)
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

func (s *AuthService) Login(login, password string) (*JWTToken, error) {
	passwordHash, err := utils.HashPassword(password, s.config.GetConfig().PasswordSecret)
	if login == "" {
		return nil, fmt.Errorf("login cannot be empty")
	}
	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}
	if err != nil {
		return nil, fmt.Errorf("AuthService Login HashPassword: %w", err)
	}
	user, err := s.userRepository.Login(login, passwordHash)
	if err != nil {
		return nil, fmt.Errorf("AuthService Login: %w", err)
	}

	jwt, err := s.tokenService.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("AuthService Login GenerateToken: %w", err)
	}
	return jwt, nil
}

func (s *AuthService) RefreshToken(refreshToken string) (*JWTToken, error) {
	jwt, err := s.tokenService.RefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("AuthService RefreshToken: %w", err)
	}
	return jwt, nil
}

func (s *AuthService) Logout(acssToken string) error {
	if err := s.userRepository.Logout(acssToken); err != nil {
		return fmt.Errorf("AuthService Logout: %w", err)
	}
	return nil
}

func (s *AuthService) Register(login, email, password string) (*JWTToken, error) {

	if err := s.validatePassword(password); err != nil {
		return nil, err
	}
	passwordHash, err := utils.HashPassword(password, s.config.GetConfig().PasswordSecret)
	if err != nil {
		return nil, fmt.Errorf("AuthService Register HashPassword: %w", err)
	}
	user, err := s.userRepository.Register(login, email, passwordHash)
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
	s.SendEmailConfirmation(user.Email)
	return jwt, nil

}

func (s *AuthService) ChangePassword(login, oldPassword, newPassword, loginFromJWT string) error {
	if loginFromJWT != login {
		return ErrChangeNotYourPassword

	}
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}
	oldPasswordHash, err := utils.HashPassword(oldPassword, s.config.GetConfig().PasswordSecret)
	if err != nil {
		return fmt.Errorf("AuthService ChangePassword HashPassword: %w", err)
	}
	newPasswordHash, err := utils.HashPassword(newPassword, s.config.GetConfig().PasswordSecret)
	if err != nil {
		return fmt.Errorf("AuthService ChangePassword HashPassword: %w", err)
	}
	if err := s.userRepository.ChangePassword(login, oldPasswordHash, newPasswordHash); err != nil {
		return fmt.Errorf("AuthService ChangePassword: %w", err)
	}
	return nil

}
func (s *AuthService) RestorePassword(email string) error {
	_, err := s.userRepository.GetUserByEmail(email)
	if err != nil {
		return fmt.Errorf("AuthService RestorePassword GetUserByEmail: %w", err)
	}
	// TODO Send Email
	// TODO html with restore password
	return nil
}
func (s *AuthService) SendEmailConfirmation(email string) error {
	user, err := s.userRepository.GetUserByEmail(email)
	if err != nil {
		return fmt.Errorf("AuthService GetUserByEmail: %w", err)
	}
	if user == nil {
		return fmt.Errorf("AuthService GetUserByEmail: %w", err)

	}
	// TODO send email
	//TODO Change EmailConfirmed method
	user.EmailConfirmed = true
	user, err = s.userRepository.Update(user)
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

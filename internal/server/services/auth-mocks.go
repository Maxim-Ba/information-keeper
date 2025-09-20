package services

import (
	"context"

	config "github.com/Maxim-Ba/information-keeper/config/server"
	"github.com/Maxim-Ba/information-keeper/internal/server/dto"
	"github.com/stretchr/testify/mock"
)

var _ UserRepository = (*MockUserRepository)(nil)
var _ TokenServiceInterface = (*MockTokenService)(nil)
var _ AppConfig = (*MockAppConfig)(nil)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Register(ctx context.Context, login, email, password string) (*dto.UserRepoDTO, error) {
	args := m.Called(login, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserRepoDTO), args.Error(1)
}

func (m *MockUserRepository) Login(ctx context.Context, login, password string) (*dto.UserRepoDTO, error) {
	args := m.Called(login, password)
	return args.Get(0).(*dto.UserRepoDTO), args.Error(1)
}

func (m *MockUserRepository) ChangePassword(ctx context.Context, login, oldPassword, newPassword string) error {
	args := m.Called(login, oldPassword, newPassword)
	return args.Error(0)
}

func (m *MockUserRepository) RestorePassword(ctx context.Context, email string) error {
	args := m.Called(email)
	return args.Error(0)
}

func (m *MockUserRepository) Logout(ctx context.Context, acssToken string) error {
	args := m.Called(acssToken)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*dto.UserRepoDTO, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserRepoDTO), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *dto.UserRepoDTO) (*dto.UserRepoDTO, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserRepoDTO), args.Error(1)
}

// MockTokenService реализация TokenService для тестов
type MockTokenService struct {
	mock.Mock
}

// Remove implements TokenServiceInterface.
func (m *MockTokenService) Remove(ctx context.Context,token *JWTToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockTokenService) RefreshToken(ctx context.Context,token string) (*JWTToken, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*JWTToken), args.Error(1)
}

func (m *MockTokenService) GenerateToken(user *dto.UserRepoDTO) (*JWTToken, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*JWTToken), args.Error(1)
}

// MockAppConfig реализация AppConfig для тестов
type MockAppConfig struct {
	mock.Mock
}

func (m *MockAppConfig) GetConfig() config.ServerCfg {
	args := m.Called()
	return args.Get(0).(config.ServerCfg)
}

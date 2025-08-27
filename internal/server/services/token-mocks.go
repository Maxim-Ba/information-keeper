package services

import (
	"time"

	"github.com/Maxim-Ba/information-keeper/internal/server/dto"
	"github.com/stretchr/testify/mock"
)

type MockTokenRepository struct {
	mock.Mock
}

func (m *MockTokenRepository) AddToBlacklist(token string, expiry time.Time) error {
	args := m.Called(token, expiry)
	return args.Error(0)
}

func (m *MockTokenRepository) IsInBlacklist(token string) (bool, error) {
	args := m.Called(token)
	return args.Bool(0), args.Error(1)
}

type MockUserRepositoryInterface struct {
	mock.Mock
}

func (m *MockUserRepositoryInterface) GetUserByID(id string) (*dto.UserRepoDTO, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserRepoDTO), args.Error(1)
}

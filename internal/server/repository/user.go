package repository

import (
	"database/sql"

	"github.com/Maxim-Ba/information-keeper/internal/server/dto"
)

type UserRepository struct {
	db *sql.DB
}

// ChangePassword implements services.UserRepository.
func (u *UserRepository) ChangePassword(login string, oldPassword string, newPassword string) error {
	panic("unimplemented")
}

// GetUserByEmail implements services.UserRepository.
func (u *UserRepository) GetUserByEmail(email string) (*dto.UserRepoDTO, error) {
	panic("unimplemented")
}

// Login implements services.UserRepository.
func (u *UserRepository) Login(login string, password string) (*dto.UserRepoDTO, error) {
	panic("unimplemented")
}



// Register implements services.UserRepository.
func (u *UserRepository) Register(login string, email string, password string) (*dto.UserRepoDTO, error) {
	panic("unimplemented")
}

// RestorePassword implements services.UserRepository.
func (u *UserRepository) RestorePassword(email string) error {
	panic("unimplemented")
}

// Update implements services.UserRepository.
func (u *UserRepository) Update(user *dto.UserRepoDTO) (*dto.UserRepoDTO, error) {
	panic("unimplemented")
}

// GetUserByID implements services.UserRepositoryInterface.
func (u *UserRepository) GetUserByID(id string) (*dto.UserRepoDTO, error) {
	panic("unimplemented")
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

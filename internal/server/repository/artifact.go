package repository

import (
	"database/sql"

	"github.com/Maxim-Ba/information-keeper/internal/domain"
)

type S3Interface interface {
}
type ArtifactRepository struct {
	s3client S3Interface
	db       *sql.DB
}

func NewArtifactRepository(s3client S3Interface, db *sql.DB) *ArtifactRepository {
	return &ArtifactRepository{
		s3client: s3client,
		db:       db,
	}
}

func (r *ArtifactRepository) GetArtifacts(userID string, page int32, onpage int32) ([]domain.Artifact, error) {
	return []domain.Artifact{}, nil

}
func (r *ArtifactRepository) CreateArtifact(artifact *domain.Artifact) (*domain.Artifact, error) {
	return &domain.Artifact{}, nil
}
func (r *ArtifactRepository) UpdateArtifact(artifact *domain.Artifact) (*domain.Artifact, error) {
	return &domain.Artifact{}, nil
}
func (r *ArtifactRepository) DeleteArtifact(id string) (*domain.Artifact, error) {
	return &domain.Artifact{}, nil

}
func (r *ArtifactRepository) GetArtifactByID(id string) (*domain.Artifact, error) {
	return &domain.Artifact{}, nil

}

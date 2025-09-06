package services

import (
	"fmt"

	"github.com/Maxim-Ba/information-keeper/internal/domain"
	"github.com/Maxim-Ba/information-keeper/pkg/proto"
)

type ArtifactService struct {
	artifactRepository ArtifactRepositoryInterface
	syncManager        SyncManagerInterface
}
type ArtifactRepositoryInterface interface {
	GetArtifacts(userID string, page int32, onpage int32) ([]domain.Artifact, error)
	CreateArtifact(artifact *domain.Artifact) (*domain.Artifact, error)
	UpdateArtifact(artifact *domain.Artifact) (*domain.Artifact, error)
	DeleteArtifact(id string) (*domain.Artifact, error)
	GetArtifactByID(id string) (*domain.Artifact, error)
}
type SyncManagerInterface interface {
	Subscribe(userID string, clientID string) chan proto.SyncEvent
	Unsubscribe(userID string, clientID string)
	Broadcast(userID string, event proto.SyncEvent)
}

func NewArtifactService(
	artifactRepository ArtifactRepositoryInterface,
	syncManager SyncManagerInterface,
) *ArtifactService {
	return &ArtifactService{
		artifactRepository: artifactRepository,
		syncManager:        syncManager,
	}
}

func (s *ArtifactService) GetArtifacts(userID string, page int32, onpage int32) ([]domain.Artifact, error) {

	artifacts, err := s.artifactRepository.GetArtifacts(userID, page, onpage)

	if err != nil {
		return nil, fmt.Errorf("ArtifactService GetArtifacts: %w", err)
	}
	return artifacts, nil
}

func (s *ArtifactService) CreateArtifact(artifact *domain.Artifact) error {
	a, err := s.artifactRepository.CreateArtifact(artifact)
	if err != nil {
		return fmt.Errorf("ArtifactService CreateArtifact: %w", err)
	}

	s.syncManager.Broadcast(a.OwerID, proto.SyncEvent{
		// TODO: add event
	})
	return nil
}

func (s *ArtifactService) UpdateArtifact(userID string, artifact *domain.Artifact) error {
	artifact, err := s.artifactRepository.GetArtifactByID(artifact.ID)
	if err != nil {
		return err
	}
	if !s.isUsersArtifacts(userID, artifact) {
		return ErrForbiddenAction
	}
	a, err := s.artifactRepository.UpdateArtifact(artifact)
	if err != nil {
		return fmt.Errorf("ArtifactService UpdateArtifact: %w", err)
	}
	s.syncManager.Broadcast(a.OwerID, proto.SyncEvent{
		// TODO: add event
	})
	return nil
}
func (s *ArtifactService) DeleteArtifact(userID string, id string) error {
	artifact, err := s.artifactRepository.GetArtifactByID(id)
	if err != nil {
		return err
	}
	if !s.isUsersArtifacts(userID, artifact) {
		return ErrForbiddenAction
	}
	a, err := s.artifactRepository.DeleteArtifact(id)
	if err != nil {
		return fmt.Errorf("ArtifactService DeleteArtifact: %w", err)
	}

	s.syncManager.Broadcast(a.OwerID, proto.SyncEvent{
		// TODO: add event
	})
	return nil
}
func (s *ArtifactService) GetArtifactByID(userID string, id string) (*domain.Artifact, error) {
	if artifact, err := s.artifactRepository.GetArtifactByID(id); err != nil {
		return nil, fmt.Errorf("ArtifactService GetArtifactByID: %w", err)
	} else {
		if !s.isUsersArtifacts(userID, artifact) {
			return nil, ErrForbiddenAction
		}
		return artifact, nil
	}
}

// Get with one time password (delete after send response)
func (s *ArtifactService) GetWithOTP(userID string, id string) (*domain.Artifact, error) {
	artifact, err := s.artifactRepository.GetArtifactByID(id)
	if err != nil {
		return nil, fmt.Errorf("ArtifactService GetWithOTP GetArtifactByID: %w", err)
	}
	if !s.isUsersArtifacts(userID, artifact) {
		return nil, ErrForbiddenAction
	}
	if err := s.DeleteArtifact(userID, id); err != nil {
		return nil, fmt.Errorf("ArtifactService GetWithOTP DeleteArtifact: %w", err)
	}
	return artifact, nil
}

func (s *ArtifactService) isUsersArtifacts(userID string, artifact *domain.Artifact) bool {
	return userID == artifact.OwerID

}

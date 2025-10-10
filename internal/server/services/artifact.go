package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Maxim-Ba/information-keeper/internal/domain"
	eventidgen "github.com/Maxim-Ba/information-keeper/pkg/event-id-gen"
	"github.com/Maxim-Ba/information-keeper/pkg/logger"
	"github.com/Maxim-Ba/information-keeper/pkg/proto"
)

type ArtifactService struct {
	artifactRepository ArtifactRepositoryInterface
	syncManager        SyncManagerInterface
}
type ArtifactRepositoryInterface interface {
	GetArtifacts(ctx context.Context, userID string, page int32, onpage int32) ([]domain.Artifact, error)
	CreateArtifact(ctx context.Context, artifact *domain.Artifact) (*domain.Artifact, error)
	UpdateArtifact(ctx context.Context, artifact *domain.Artifact) (*domain.Artifact, error)
	DeleteArtifact(ctx context.Context, id string) (*domain.Artifact, error)
	GetArtifactByID(ctx context.Context, id string) (*domain.Artifact, error)
	GetPresignedURL(ctx context.Context, id string) (string, error)
	GetDownloadURL(ctx context.Context, artifactID string) (string, error)
}
type SyncManagerInterface interface {
	Subscribe(userID string, clientID string) chan *proto.SyncEvent
	Unsubscribe(userID string, clientID string)
	Broadcast(userID string, event *proto.SyncEvent)
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

func (s *ArtifactService) GetArtifacts(ctx context.Context, userID string, page int32, onpage int32) ([]domain.Artifact, error) {

	artifacts, err := s.artifactRepository.GetArtifacts(ctx, userID, page, onpage)

	if err != nil {
		return nil, fmt.Errorf("ArtifactService GetArtifacts: %w", err)
	}
	return artifacts, nil
}

func (s *ArtifactService) CreateArtifact(ctx context.Context, userID string, req *proto.CreateArtifactRequest) error {
	// Парсим данные из бинарного payload
	artifactData, err := domain.ParseArtifactData(req.Type, req.Payload)
	if err != nil {
		return fmt.Errorf("failed to parse artifact data: %w", err)
	}

	// Валидируем данные
	if err := artifactData.Validate(); err != nil {
		return fmt.Errorf("artifact data validation failed: %w", err)
	}

	// Создаем доменный артефакт
	artifact := &domain.Artifact{
		OwnerID:   userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ExpiredAt: time.Unix(req.ExpiredAt, 0),
		Type: domain.ArtifactType{
			Id:   int(req.Type),
			Name: s.getTypeName(req.Type),
		},
		MetaInfo: req.MetaInfo,
		Link:     req.Link,
		Payload:  req.Payload, // Сохраняем оригинальные бинарные данные
	}

	// Если это файл, обрабатываем отдельно
	if req.Type == proto.ArtifactTypeEnum_BINARY {
		if err := s.processBinaryArtifact(ctx, artifact, req.Payload); err != nil {
			return err
		}
	}

	a, err := s.artifactRepository.CreateArtifact(ctx, artifact)
	if err != nil {
		logger.Error(err.Error())
		return fmt.Errorf("ArtifactService CreateArtifact: %w", err)
	}

	// Отправляем событие синхронизации
	pbArtifact := domain.DomainToPBArtifact(a)
	s.syncManager.Broadcast(userID,
		&proto.SyncEvent{
			EventId:   eventidgen.GenerateEventID(),
			Timestamp: time.Now().Unix(),
			EventType: &proto.SyncEvent_ArtifactCreated{
				ArtifactCreated: &proto.ArtifactCreatedEvent{
					Artifact: pbArtifact,
				},
			},
		},
	)

	return nil
}

func (s *ArtifactService) UpdateArtifact(ctx context.Context, userID string, artifact *domain.Artifact) error {
	existingArtifact, err := s.artifactRepository.GetArtifactByID(ctx, artifact.ID)
	if err != nil {
		return err
	}

	if !s.isUsersArtifacts(userID, existingArtifact) {
		return ErrForbiddenAction
	}
	if len(artifact.Payload) > 0 {
		artifactData, err := domain.ParseArtifactData(proto.ArtifactTypeEnum(existingArtifact.Type.Id), artifact.Payload)
		if err != nil {
			return fmt.Errorf("failed to parse artifact data: %w", err)
		}

		if err := artifactData.Validate(); err != nil {
			return fmt.Errorf("artifact data validation failed: %w", err)
		}
	}
	// Обновляем только разрешенные поля
	existingArtifact.UpdatedAt = time.Now()
	existingArtifact.ExpiredAt = artifact.ExpiredAt
	existingArtifact.MetaInfo = artifact.MetaInfo
	existingArtifact.Link = artifact.Link
	if len(artifact.Payload) > 0 {
		existingArtifact.Payload = artifact.Payload
	}
	a, err := s.artifactRepository.UpdateArtifact(ctx, existingArtifact)
	if err != nil {
		return fmt.Errorf("ArtifactService UpdateArtifact: %w", err)
	}

	pbArtifact := domain.DomainToPBArtifact(a)
	s.syncManager.Broadcast(userID, &proto.SyncEvent{
		EventId:   eventidgen.GenerateEventID(),
		Timestamp: time.Now().Unix(),
		EventType: &proto.SyncEvent_ArtifactUpdated{
			ArtifactUpdated: &proto.ArtifactUpdatedEvent{
				Artifact: pbArtifact,
			},
		},
	})

	return nil
}
func (s *ArtifactService) DeleteArtifact(ctx context.Context, userID string, id string) error {
	artifact, err := s.artifactRepository.GetArtifactByID(ctx, id)
	if err != nil {
		return err
	}

	if !s.isUsersArtifacts(userID, artifact) {
		return ErrForbiddenAction
	}

	a, err := s.artifactRepository.DeleteArtifact(ctx, id)
	if err != nil {
		return fmt.Errorf("ArtifactService DeleteArtifact: %w", err)
	}

	s.syncManager.Broadcast(userID, &proto.SyncEvent{
		EventId:   eventidgen.GenerateEventID(),
		Timestamp: time.Now().Unix(),
		EventType: &proto.SyncEvent_ArtifactDeleted{
			ArtifactDeleted: &proto.ArtifactDeletedEvent{
				ArtifactId: a.ID,
			},
		},
	})

	return nil
}
func (s *ArtifactService) GetArtifactByID(ctx context.Context, userID string, id string) (*domain.Artifact, error) {
	if artifact, err := s.artifactRepository.GetArtifactByID(ctx, id); err != nil {
		return nil, fmt.Errorf("ArtifactService GetArtifactByID: %w", err)
	} else {
		if !s.isUsersArtifacts(userID, artifact) {
			return nil, ErrForbiddenAction
		}
		return artifact, nil
	}
}

// Get with one time password (delete after send response)
func (s *ArtifactService) GetWithOTP(ctx context.Context, userID string, id string) (*domain.Artifact, error) {
	artifact, err := s.artifactRepository.GetArtifactByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ArtifactService GetWithOTP GetArtifactByID: %w", err)
	}
	if !s.isUsersArtifacts(userID, artifact) {
		return nil, ErrForbiddenAction
	}
	if err := s.DeleteArtifact(ctx, userID, id); err != nil {
		return nil, fmt.Errorf("ArtifactService GetWithOTP DeleteArtifact: %w", err)
	}
	return artifact, nil
}
func (s *ArtifactService) GetPresignedURL(ctx context.Context, userID string, id string) (string, error) {
	artifact, err := s.artifactRepository.GetArtifactByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("ArtifactService GetPresignedURL: %w", err)
	}

	if !s.isUsersArtifacts(userID, artifact) {
		return "", ErrForbiddenAction
	}

	url, err := s.artifactRepository.GetPresignedURL(ctx, id)
	if err != nil {
		return "", fmt.Errorf("ArtifactService GetPresignedURL: %w", err)
	}

	return url, nil
}

func (s *ArtifactService) isUsersArtifacts(userID string, artifact *domain.Artifact) bool {
	return userID == artifact.OwnerID
}
func (s *ArtifactService) getTypeName(artifactType proto.ArtifactTypeEnum) string {
	switch artifactType {
	case proto.ArtifactTypeEnum_TEXT:
		return "text"
	case proto.ArtifactTypeEnum_LOGIN_PASSWORD:
		return "loginpassword"
	case proto.ArtifactTypeEnum_BANK_CARD:
		return "bankcard"
	case proto.ArtifactTypeEnum_BINARY:
		return "binary"
	default:
		return "text"
	}
}

func (s *ArtifactService) processBinaryArtifact(ctx context.Context, artifact *domain.Artifact, payload []byte) error {
	// Для бинарных файлов парсим данные чтобы получить метаинформацию
	var binaryData domain.BinaryData
	if err := binaryData.FromBinary(payload); err != nil {
		return fmt.Errorf("failed to parse binary data: %w", err)
	}

	// Обновляем метаинформацию если она не установлена
	if artifact.MetaInfo == "" {
		artifact.MetaInfo = binaryData.GetMetaInfo()
	}

	return nil
}
func (s *ArtifactService) GetDownloadURL(ctx context.Context, userID string, artifactID string) (string, error) {
	artifact, err := s.artifactRepository.GetArtifactByID(ctx, artifactID)
	if err != nil {
		return "", fmt.Errorf("ArtifactService GetDownloadURL: %w", err)
	}

	if !s.isUsersArtifacts(userID, artifact) {
		return "", ErrForbiddenAction
	}

	url, err := s.artifactRepository.GetDownloadURL(ctx, artifactID)
	if err != nil {
		return "", fmt.Errorf("ArtifactService GetDownloadURL: %w", err)
	}

	return url, nil
}

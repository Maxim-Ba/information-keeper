// repository/artifact.go
package repository

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Maxim-Ba/information-keeper/internal/domain"
	"github.com/Maxim-Ba/information-keeper/internal/server/s3client"
)

type ArtifactRepository struct {
	s3client s3client.S3Client
	db       *sql.DB
}

func NewArtifactRepository(s3client s3client.S3Client, db *sql.DB) *ArtifactRepository {
	return &ArtifactRepository{
		s3client: s3client,
		db:       db,
	}
}

func (r *ArtifactRepository) GetArtifacts(ctx context.Context, userID string, page int32, onpage int32) ([]domain.Artifact, error) {
	offset := (page - 1) * onpage

	query := `
        SELECT a.id, a.owner_id, a.created_at, a.updated_at, a.expired_at, 
               at.id as type_id, at.name as type_name, a.meta_info, a.link
        FROM artifacts a
        JOIN artifact_types at ON a.type_id = at.id
        WHERE a.owner_id = $1
        ORDER BY a.created_at DESC
        LIMIT $2 OFFSET $3
    `

	rows, err := r.db.QueryContext(ctx, query, userID, onpage, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get artifacts: %w", err)
	}
	defer rows.Close()

	var artifacts []domain.Artifact
	for rows.Next() {
		var artifact domain.Artifact
		var artifactType domain.ArtifactType

		err := rows.Scan(
			&artifact.ID,
			&artifact.OwerID,
			&artifact.CreatedAt,
			&artifact.UpdatedAt,
			&artifact.ExpiredAt,
			&artifactType.Id,
			&artifactType.Name,
			&artifact.MetaInfo,
			&artifact.Link,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan artifact: %w", err)
		}

		artifact.Type = artifactType
		artifacts = append(artifacts, artifact)
	}

	return artifacts, nil
}

func (r *ArtifactRepository) CreateArtifact(ctx context.Context, artifact *domain.Artifact) (*domain.Artifact, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Если это файл, загружаем в S3
	if artifact.Type.Id == 3 { // Предположим, что тип 3 = файл
		if err := r.uploadFileToS3(ctx, artifact); err != nil {
			return nil, err
		}
	}

	query := `
        INSERT INTO artifacts (id, owner_id, created_at, updated_at, expired_at, type_id, meta_info, link)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, owner_id, created_at, updated_at, expired_at, type_id, meta_info, link
    `

	var createdArtifact domain.Artifact
	var typeID int

	err = tx.QueryRowContext(ctx, query,
		artifact.ID,
		artifact.OwerID,
		time.Now(),
		time.Now(),
		artifact.ExpiredAt,
		artifact.Type.Id,
		artifact.MetaInfo,
		artifact.Link,
	).Scan(
		&createdArtifact.ID,
		&createdArtifact.OwerID,
		&createdArtifact.CreatedAt,
		&createdArtifact.UpdatedAt,
		&createdArtifact.ExpiredAt,
		&typeID,
		&createdArtifact.MetaInfo,
		&createdArtifact.Link,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create artifact: %w", err)
	}

	// Получаем информацию о типе
	typeQuery := `SELECT id, name FROM artifact_types WHERE id = $1`
	var artifactType domain.ArtifactType
	err = tx.QueryRowContext(ctx, typeQuery, typeID).Scan(&artifactType.Id, &artifactType.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get artifact type: %w", err)
	}

	createdArtifact.Type = artifactType

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &createdArtifact, nil
}

func (r *ArtifactRepository) UpdateArtifact(ctx context.Context, artifact *domain.Artifact) (*domain.Artifact, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
        UPDATE artifacts 
        SET updated_at = $1, expired_at = $2, meta_info = $3, link = $4
        WHERE id = $5
        RETURNING id, owner_id, created_at, updated_at, expired_at, type_id, meta_info, link
    `

	var updatedArtifact domain.Artifact
	var typeID int

	err = tx.QueryRowContext(ctx, query,
		time.Now(),
		artifact.ExpiredAt,
		artifact.MetaInfo,
		artifact.Link,
		artifact.ID,
	).Scan(
		&updatedArtifact.ID,
		&updatedArtifact.OwerID,
		&updatedArtifact.CreatedAt,
		&updatedArtifact.UpdatedAt,
		&updatedArtifact.ExpiredAt,
		&typeID,
		&updatedArtifact.MetaInfo,
		&updatedArtifact.Link,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrArtifactNotFound
		}
		return nil, fmt.Errorf("failed to update artifact: %w", err)
	}

	// Получаем информацию о типе
	typeQuery := `SELECT id, name FROM artifact_types WHERE id = $1`
	var artifactType domain.ArtifactType
	err = tx.QueryRowContext(ctx, typeQuery, typeID).Scan(&artifactType.Id, &artifactType.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get artifact type: %w", err)
	}

	updatedArtifact.Type = artifactType

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &updatedArtifact, nil
}

func (r *ArtifactRepository) DeleteArtifact(ctx context.Context, id string) (*domain.Artifact, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Сначала получаем артефакт для возврата и проверки типа
	artifact, err := r.GetArtifactByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Если это файл, удаляем из S3
	if artifact.Type.Id == 3 { // Тип 3 = файл
		if err := r.deleteFileFromS3(ctx, artifact.Link); err != nil {
			return nil, err
		}
	}

	query := `DELETE FROM artifacts WHERE id = $1 RETURNING id`
	var deletedID string
	err = tx.QueryRowContext(ctx, query, id).Scan(&deletedID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrArtifactNotFound
		}
		return nil, fmt.Errorf("failed to delete artifact: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return artifact, nil
}

func (r *ArtifactRepository) GetArtifactByID(ctx context.Context, id string) (*domain.Artifact, error) {
	query := `
        SELECT a.id, a.owner_id, a.created_at, a.updated_at, a.expired_at, 
               at.id as type_id, at.name as type_name, a.meta_info, a.link
        FROM artifacts a
        JOIN artifact_types at ON a.type_id = at.id
        WHERE a.id = $1
    `

	var artifact domain.Artifact
	var artifactType domain.ArtifactType

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&artifact.ID,
		&artifact.OwerID,
		&artifact.CreatedAt,
		&artifact.UpdatedAt,
		&artifact.ExpiredAt,
		&artifactType.Id,
		&artifactType.Name,
		&artifact.MetaInfo,
		&artifact.Link,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrArtifactNotFound
		}
		return nil, fmt.Errorf("failed to get artifact: %w", err)
	}

	artifact.Type = artifactType
	return &artifact, nil
}

// Вспомогательные методы для работы с S3
func (r *ArtifactRepository) uploadFileToS3(ctx context.Context, artifact *domain.Artifact) error {
	// Парсим метаинформацию для получения данных файла
	var fileMeta struct {
		FileName    string `json:"file_name"`
		Content     []byte `json:"content"`
		ContentType string `json:"content_type"`
	}

	if err := json.Unmarshal([]byte(artifact.MetaInfo), &fileMeta); err != nil {
		return fmt.Errorf("failed to parse file metadata: %w", err)
	}

	// Генерируем уникальный ключ для S3
	s3Key := fmt.Sprintf("users/%s/artifacts/%s/%s",
		artifact.OwerID, artifact.ID, fileMeta.FileName)

	// Загружаем файл в S3
	_, err := r.s3client.Upload(ctx, s3client.UploadInput{
		Key:         s3Key,
		Body:        bytes.NewReader(fileMeta.Content),
		ContentType: fileMeta.ContentType,
		ContentSize: int64(len(fileMeta.Content)),
	})

	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Обновляем ссылку на файл в S3
	artifact.Link = s3Key
	return nil
}

func (r *ArtifactRepository) deleteFileFromS3(ctx context.Context, s3Key string) error {
	if err := r.s3client.Delete(ctx, s3client.DeleteInput{
		Key: s3Key,
	}); err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}
	return nil
}

func (r *ArtifactRepository) GetPresignedURL(ctx context.Context, artifactID string) (string, error) {
	artifact, err := r.GetArtifactByID(ctx, artifactID)
	if err != nil {
		return "", err
	}

	if artifact.Type.Id != 3 { // Только для файлов
		return "", ErrInvalidInput
	}

	url, err := r.s3client.GetPresignedURL(ctx, s3client.PresignedURLInput{
		Key:     artifact.Link,
		Method:  "GET",
		Expires: 15 * time.Minute,
	})

	if err != nil {
		return "", fmt.Errorf("failed to get presigned URL: %w", err)
	}

	return url, nil
}

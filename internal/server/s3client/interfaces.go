package s3client

import (
	"context"
)

// S3Client интерфейс для работы с S3.
type S3Client interface {
	// Upload загрузка файла
	Upload(ctx context.Context, input UploadInput) (*UploadOutput, error)

	// Download скачивание файла
	Download(ctx context.Context, input DownloadInput) (*DownloadOutput, error)

	// Delete удаление файла
	Delete(ctx context.Context, input DeleteInput) error

	// List список файлов
	List(ctx context.Context, input ListInput) (*ListOutput, error)

	// Exists проверка существования файла
	Exists(ctx context.Context, input ExistsInput) (bool, error)

	// GetPresignedURL получение предварительно подписанного URL
	GetPresignedURL(ctx context.Context, input PresignedURLInput) (string, error)

	// Copy копирование файла
	Copy(ctx context.Context, input CopyInput) error

	// GetMetadata получение метаданных файла
	GetMetadata(ctx context.Context, input MetadataInput) (*MetadataOutput, error)

	// Close закрытие клиента
	Close() error
}

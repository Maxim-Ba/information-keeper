package s3client

import (
	"io"
	"time"
)

// Config конфигурация клиента
type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	Bucket          string
	UseSSL          bool
	ForcePathStyle  bool
}

// UploadInput входные данные для загрузки
type UploadInput struct {
	Key         string
	Body        io.Reader
	ContentType string
	ContentSize int64
	Metadata    map[string]string
}

// UploadOutput выходные данные загрузки
type UploadOutput struct {
	Key      string
	Location string
	ETag     string
	Size     int64
}

// DownloadInput входные данные для скачивания
type DownloadInput struct {
	Key string
}

// DownloadOutput выходные данные скачивания
type DownloadOutput struct {
	Body        io.ReadCloser
	ContentType string
	Size        int64
	ETag        string
}

// DeleteInput входные данные для удаления
type DeleteInput struct {
	Key string
}

// ListInput входные данные для списка
type ListInput struct {
	Prefix    string
	Delimiter string
	MaxKeys   int32
}

// ListOutput выходные данные списка
type ListOutput struct {
	Objects  []ObjectInfo
	Prefixes []string
}

// ObjectInfo информация об объекте
type ObjectInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	ETag         string
}

// ExistsInput входные данные для проверки существования
type ExistsInput struct {
	Key string
}

// PresignedURLInput входные данные для предварительно подписанного URL
type PresignedURLInput struct {
	Key     string
	Expires time.Duration
	Method  string // "GET", "PUT", "DELETE"
}

// CopyInput входные данные для копирования
type CopyInput struct {
	SourceKey      string
	DestinationKey string
}

// MetadataInput входные данные для метаданных
type MetadataInput struct {
	Key string
}

// MetadataOutput выходные данные метаданных
type MetadataOutput struct {
	Metadata     map[string]string
	Size         int64
	LastModified time.Time
	ContentType  string
}

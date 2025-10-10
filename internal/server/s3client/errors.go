package s3client

import (
	"errors"
	"fmt"
)

var (
	// ErrNotFound indicates that the requested object was not found.
	ErrNotFound = errors.New("object not found")
	// ErrAccessDenied indicates that access to the resource was denied.
	ErrAccessDenied = errors.New("access denied")
	// ErrBucketNotExists indicates that the specified bucket does not exist.
	ErrBucketNotExists = errors.New("bucket does not exist")
	// ErrInvalidInput indicates that the provided input is invalid.
	ErrInvalidInput = errors.New("invalid input")
	// ErrUploadFailed indicates that the upload operation failed.
	ErrUploadFailed = errors.New("upload failed")
	// ErrDownloadFailed indicates that the download operation failed.
	ErrDownloadFailed = errors.New("download failed")
)

// S3Error ошибка S3 операций.
type S3Error struct {
	Operation string
	Key       string
	Err       error
}

func (e *S3Error) Error() string {
	return fmt.Sprintf("s3 operation %s failed for key %s: %v", e.Operation, e.Key, e.Err)
}

func (e *S3Error) Unwrap() error {
	return e.Err
}

// NewS3Error creates a new S3Error with the given operation, key, and underlying error.
func NewS3Error(operation, key string, err error) error {
	return &S3Error{
		Operation: operation,
		Key:       key,
		Err:       err,
	}
}

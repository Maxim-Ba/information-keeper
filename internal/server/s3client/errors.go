package s3client

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound        = errors.New("object not found")
	ErrAccessDenied    = errors.New("access denied")
	ErrBucketNotExists = errors.New("bucket does not exist")
	ErrInvalidInput    = errors.New("invalid input")
	ErrUploadFailed    = errors.New("upload failed")
	ErrDownloadFailed  = errors.New("download failed")
)

// S3Error ошибка S3 операций
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

func NewS3Error(operation, key string, err error) error {
	return &S3Error{
		Operation: operation,
		Key:       key,
		Err:       err,
	}
}

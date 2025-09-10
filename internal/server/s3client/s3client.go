package s3client

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

// Client реализация S3 клиента
type Client struct {
    s3Client *s3.Client
    config   Config
}

func New(cfg Config) (S3Client, error) {
    var opts []func(*config.LoadOptions) error

    // Если указаны явные credentials
    if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
        creds := credentials.NewStaticCredentialsProvider(
            cfg.AccessKeyID,
            cfg.SecretAccessKey,
            "",
        )
        opts = append(opts, config.WithCredentialsProvider(creds))
    }

    if cfg.Region != "" {
        opts = append(opts, config.WithRegion(cfg.Region))
    }

    awsCfg, err := config.LoadDefaultConfig(context.Background(), opts...)
    if err != nil {
        return nil, err
    }

    // Кастомный endpoint для S3-совместимых хранилищ
    if cfg.Endpoint != "" {
        awsCfg.BaseEndpoint = aws.String(cfg.Endpoint)
    }

    s3Opts := make([]func(*s3.Options), 0)
    if cfg.ForcePathStyle {
        s3Opts = append(s3Opts, func(o *s3.Options) {
            o.UsePathStyle = true
        })
    }

    client := s3.NewFromConfig(awsCfg, s3Opts...)

    return &Client{
        s3Client: client,
        config:   cfg,
    }, nil
}

func (c *Client) Upload(ctx context.Context, input UploadInput) (*UploadOutput, error) {
    if input.Key == "" {
        return nil, ErrInvalidInput
    }

    uploadParams := &s3.PutObjectInput{
        Bucket:        aws.String(c.config.Bucket),
        Key:           aws.String(input.Key),
        Body:          input.Body,
        ContentLength: aws.Int64(input.ContentSize),
    }

    if input.ContentType != "" {
        uploadParams.ContentType = aws.String(input.ContentType)
    }

    if len(input.Metadata) > 0 {
        uploadParams.Metadata = input.Metadata
    }

    result, err := c.s3Client.PutObject(ctx, uploadParams)
    if err != nil {
        return nil, NewS3Error("upload", input.Key, err)
    }

    return &UploadOutput{
        Key:      input.Key,
        ETag:     aws.ToString(result.ETag),
        Size:     input.ContentSize,
    }, nil
}

func (c *Client) Download(ctx context.Context, input DownloadInput) (*DownloadOutput, error) {
    if input.Key == "" {
        return nil, ErrInvalidInput
    }

    result, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
        Bucket: aws.String(c.config.Bucket),
        Key:    aws.String(input.Key),
    })
    if err != nil {
        if isNotFoundError(err) {
            return nil, ErrNotFound
        }
        return nil, NewS3Error("download", input.Key, err)
    }

    return &DownloadOutput{
        Body:        result.Body,
        ContentType: aws.ToString(result.ContentType),
        Size:        aws.ToInt64(result.ContentLength),
        ETag:        aws.ToString(result.ETag),
    }, nil
}

func (c *Client) Delete(ctx context.Context, input DeleteInput) error {
    if input.Key == "" {
        return ErrInvalidInput
    }

    _, err := c.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
        Bucket: aws.String(c.config.Bucket),
        Key:    aws.String(input.Key),
    })
    if err != nil {
        return NewS3Error("delete", input.Key, err)
    }

    return nil
}

func (c *Client) List(ctx context.Context, input ListInput) (*ListOutput, error) {
    listInput := &s3.ListObjectsV2Input{
        Bucket: aws.String(c.config.Bucket),
    }

    if input.Prefix != "" {
        listInput.Prefix = aws.String(input.Prefix)
    }
    if input.Delimiter != "" {
        listInput.Delimiter = aws.String(input.Delimiter)
    }
    if input.MaxKeys > 0 {
        listInput.MaxKeys = aws.Int32(input.MaxKeys)
    }

    result, err := c.s3Client.ListObjectsV2(ctx, listInput)
    if err != nil {
        return nil, NewS3Error("list", input.Prefix, err)
    }

    output := &ListOutput{
        Objects:  make([]ObjectInfo, 0, len(result.Contents)),
        Prefixes: make([]string, 0, len(result.CommonPrefixes)),
    }

    for _, obj := range result.Contents {
        output.Objects = append(output.Objects, ObjectInfo{
            Key:          aws.ToString(obj.Key),
            Size:         aws.ToInt64(obj.Size),
            LastModified: aws.ToTime(obj.LastModified),
            ETag:         aws.ToString(obj.ETag),
        })
    }

    for _, prefix := range result.CommonPrefixes {
        output.Prefixes = append(output.Prefixes, aws.ToString(prefix.Prefix))
    }

    return output, nil
}

func (c *Client) Exists(ctx context.Context, input ExistsInput) (bool, error) {
    if input.Key == "" {
        return false, ErrInvalidInput
    }

    _, err := c.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
        Bucket: aws.String(c.config.Bucket),
        Key:    aws.String(input.Key),
    })
    if err != nil {
        if isNotFoundError(err) {
            return false, nil
        }
        return false, NewS3Error("exists", input.Key, err)
    }

    return true, nil
}

func (c *Client) GetPresignedURL(ctx context.Context, input PresignedURLInput) (string, error) {
    if input.Key == "" {
        return "", ErrInvalidInput
    }

    presignClient := s3.NewPresignClient(c.s3Client)

    // var presignInput *s3.PresignOptions
    switch input.Method {
    case "GET":
        getObjectInput := &s3.GetObjectInput{
            Bucket: aws.String(c.config.Bucket),
            Key:    aws.String(input.Key),
        }
        result, err := presignClient.PresignGetObject(ctx, getObjectInput, 
            s3.WithPresignExpires(input.Expires))
        if err != nil {
            return "", NewS3Error("presign_get", input.Key, err)
        }
        return result.URL, nil
        
    case "PUT":
        putObjectInput := &s3.PutObjectInput{
            Bucket: aws.String(c.config.Bucket),
            Key:    aws.String(input.Key),
        }
        result, err := presignClient.PresignPutObject(ctx, putObjectInput, 
            s3.WithPresignExpires(input.Expires))
        if err != nil {
            return "", NewS3Error("presign_put", input.Key, err)
        }
        return result.URL, nil
        
    default:
        return "", ErrInvalidInput
    }
}

func (c *Client) Copy(ctx context.Context, input CopyInput) error {
    if input.SourceKey == "" || input.DestinationKey == "" {
        return ErrInvalidInput
    }

    source := c.config.Bucket + "/" + input.SourceKey
    _, err := c.s3Client.CopyObject(ctx, &s3.CopyObjectInput{
        Bucket:     aws.String(c.config.Bucket),
        Key:        aws.String(input.DestinationKey),
        CopySource: aws.String(source),
    })
    if err != nil {
        return NewS3Error("copy", input.SourceKey, err)
    }

    return nil
}

func (c *Client) GetMetadata(ctx context.Context, input MetadataInput) (*MetadataOutput, error) {
    if input.Key == "" {
        return nil, ErrInvalidInput
    }

    result, err := c.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
        Bucket: aws.String(c.config.Bucket),
        Key:    aws.String(input.Key),
    })
    if err != nil {
        if isNotFoundError(err) {
            return nil, ErrNotFound
        }
        return nil, NewS3Error("metadata", input.Key, err)
    }

    metadata := make(map[string]string)
    for k, v := range result.Metadata {
        metadata[k] = v
    }

    return &MetadataOutput{
        Metadata:     metadata,
        Size:         aws.ToInt64(result.ContentLength),
        LastModified: aws.ToTime(result.LastModified),
        ContentType:  aws.ToString(result.ContentType),
    }, nil
}

func (c *Client) Close() error {
    // AWS SDK не требует явного закрытия
    return nil
}

// isNotFoundError проверяет является ли ошибка ошибкой "не найдено"
func isNotFoundError(err error) bool {
    var apiErr smithy.APIError
    if errors.As(err, &apiErr) {
        switch apiErr.ErrorCode() {
        case "NoSuchKey", "NotFound", "NoSuchBucket":
            return true
        }
    }
    return false
}

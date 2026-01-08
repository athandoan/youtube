package storage

import (
	"context"
	"net/url"
	"time"

	"github.com/athandoan/youtube/upload-service/internal/domain"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minioStorage struct {
	client *minio.Client
}

func NewMinioStorage(endpoint, accessKey, secretKey string, useSSL bool, region string) (domain.StorageService, error) {
	opts := &minio.Options{
		Creds:        credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:       useSSL,
		Region:       region,
		BucketLookup: minio.BucketLookupPath,
	}

	client, err := minio.New(endpoint, opts)
	if err != nil {
		return nil, err
	}
	return &minioStorage{client: client}, nil
}

func (s *minioStorage) PresignedPutObject(ctx context.Context, bucket, objectKey string, expiry time.Duration) (*url.URL, error) {
	return s.client.PresignedPutObject(ctx, bucket, objectKey, expiry)
}

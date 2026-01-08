package storage

import (
	"context"
	"net/url"
	"time"

	"github.com/athandoan/youtube/streaming-service/internal/domain"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minioStorage struct {
	client *minio.Client
}

func NewMinioStorage(endpoint, accessKey, secretKey string, useSSL bool, region string) (domain.StorageService, error) {
	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
		Region: region,
	}
	client, err := minio.New(endpoint, opts)
	if err != nil {
		return nil, err
	}
	return &minioStorage{client: client}, nil
}

func (s *minioStorage) PresignedGetObject(ctx context.Context, bucket, objectKey string, expiry time.Duration) (*url.URL, error) {
	reqParams := make(url.Values)
	return s.client.PresignedGetObject(ctx, bucket, objectKey, expiry, reqParams)
}

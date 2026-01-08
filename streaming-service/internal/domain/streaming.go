package domain

import (
	"context"
	"net/url"
	"time"
)

type VideoMetadata struct {
	ID         string
	BucketName string
	ObjectKey  string
}

type MetadataService interface {
	GetVideo(ctx context.Context, id string) (*VideoMetadata, error)
}

type StorageService interface {
	PresignedGetObject(ctx context.Context, bucket, objectKey string, expiry time.Duration) (*url.URL, error)
}

type StreamingUsecase interface {
	GetStreamURL(ctx context.Context, videoID string) (string, error)
}

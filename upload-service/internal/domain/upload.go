package domain

import (
	"context"
	"net/url"
	"time"
)

type StorageService interface {
	PresignedPutObject(ctx context.Context, bucket, objectKey string, expiry time.Duration) (*url.URL, error)
}

type MetadataService interface {
	CreateVideo(ctx context.Context, title, bucket, objectKey string) (string, error)
	UpdateVideoStatus(ctx context.Context, id, status string) error
}

type UploadUsecase interface {
	InitUpload(ctx context.Context, title, filename string) (string, string, error) // returns videoID, presignedURL
	CompleteUpload(ctx context.Context, videoID string) error
}

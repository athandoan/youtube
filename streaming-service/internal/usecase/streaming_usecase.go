package usecase

import (
	"context"
	"time"

	"github.com/athandoan/youtube/streaming-service/internal/domain"
)

type streamingUsecase struct {
	storage       domain.StorageService
	metadata      domain.MetadataService
	defaultBucket string
}

func NewStreamingUsecase(storage domain.StorageService, metadata domain.MetadataService, bucket string) domain.StreamingUsecase {
	return &streamingUsecase{
		storage:       storage,
		metadata:      metadata,
		defaultBucket: bucket,
	}
}

func (u *streamingUsecase) GetStreamURL(ctx context.Context, videoID string) (string, error) {
	// 1. Get Metadata
	v, err := u.metadata.GetVideo(ctx, videoID)
	if err != nil {
		return "", err
	}

	bucket := v.BucketName
	if bucket == "" {
		bucket = u.defaultBucket
	}

	// 2. Presign
	expiry := time.Hour * 1
	url, err := u.storage.PresignedGetObject(ctx, bucket, v.ObjectKey, expiry)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

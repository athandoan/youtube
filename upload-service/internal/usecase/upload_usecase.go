package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/athandoan/youtube/upload-service/internal/domain"
	"github.com/google/uuid"
)

type uploadUsecase struct {
	storage    domain.StorageService
	metadata   domain.MetadataService
	bucketName string
}

func NewUploadUsecase(storage domain.StorageService, metadata domain.MetadataService, bucketName string) domain.UploadUsecase {
	return &uploadUsecase{
		storage:    storage,
		metadata:   metadata,
		bucketName: bucketName,
	}
}

func (u *uploadUsecase) InitUpload(ctx context.Context, title, filename string) (string, string, error) {
	// Generate unique path for S3 to avoid filename collision
	fileUUID := uuid.New().String()
	objectKey := fmt.Sprintf("%s/%s", fileUUID, filename)

	// 1. Create Video in Metadata Service and get the canonical VideoID
	videoID, err := u.metadata.CreateVideo(ctx, title, u.bucketName, objectKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to create metadata: %w", err)
	}

	// 2. Generate Presigned URL
	expiry := time.Hour * 1
	url, err := u.storage.PresignedPutObject(ctx, u.bucketName, objectKey, expiry)
	if err != nil {
		return "", "", fmt.Errorf("failed to presign: %w", err)
	}

	return videoID, url.String(), nil
}

func (u *uploadUsecase) CompleteUpload(ctx context.Context, videoID string) error {
	// Call Metadata Service to update status using the canonical VideoID
	err := u.metadata.UpdateVideoStatus(ctx, videoID, "ready")
	if err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}
	return nil
}

package usecase

import (
	"context"
	"errors"
	"net/url"
	"testing"

	"github.com/athandoan/youtube/streaming-service/internal/domain"
	"github.com/athandoan/youtube/streaming-service/internal/mocks"
	"go.uber.org/mock/gomock"
)

func TestStreamingUsecase_GetStreamURL(t *testing.T) {
	tests := []struct {
		name          string
		videoID       string
		defaultBucket string
		setupMock     func(storage *mocks.MockStorageService, metadata *mocks.MockMetadataService)
		wantURL       string
		wantErr       bool
	}{
		{
			name:          "success - returns presigned URL with video's bucket",
			videoID:       "video-123",
			defaultBucket: "default-bucket",
			setupMock: func(storage *mocks.MockStorageService, metadata *mocks.MockMetadataService) {
				metadata.EXPECT().
					GetVideo(gomock.Any(), "video-123").
					Return(&domain.VideoMetadata{
						ID:         "video-123",
						BucketName: "custom-bucket",
						ObjectKey:  "uuid/video.mp4",
					}, nil)

				presignedURL, _ := url.Parse("https://s3.example.com/custom-bucket/uuid/video.mp4?signature=xxx")
				storage.EXPECT().
					PresignedGetObject(gomock.Any(), "custom-bucket", "uuid/video.mp4", gomock.Any()).
					Return(presignedURL, nil)
			},
			wantURL: "https://s3.example.com/custom-bucket/uuid/video.mp4?signature=xxx",
			wantErr: false,
		},
		{
			name:          "success - uses default bucket when video bucket is empty",
			videoID:       "video-456",
			defaultBucket: "default-bucket",
			setupMock: func(storage *mocks.MockStorageService, metadata *mocks.MockMetadataService) {
				metadata.EXPECT().
					GetVideo(gomock.Any(), "video-456").
					Return(&domain.VideoMetadata{
						ID:         "video-456",
						BucketName: "", // Empty bucket
						ObjectKey:  "uuid/video.mp4",
					}, nil)

				presignedURL, _ := url.Parse("https://s3.example.com/default-bucket/uuid/video.mp4?signature=xxx")
				storage.EXPECT().
					PresignedGetObject(gomock.Any(), "default-bucket", "uuid/video.mp4", gomock.Any()).
					Return(presignedURL, nil)
			},
			wantURL: "https://s3.example.com/default-bucket/uuid/video.mp4?signature=xxx",
			wantErr: false,
		},
		{
			name:          "error - video not found",
			videoID:       "nonexistent-id",
			defaultBucket: "default-bucket",
			setupMock: func(storage *mocks.MockStorageService, metadata *mocks.MockMetadataService) {
				metadata.EXPECT().
					GetVideo(gomock.Any(), "nonexistent-id").
					Return(nil, errors.New("video not found"))
			},
			wantErr: true,
		},
		{
			name:          "error - storage service fails to generate presigned URL",
			videoID:       "video-789",
			defaultBucket: "default-bucket",
			setupMock: func(storage *mocks.MockStorageService, metadata *mocks.MockMetadataService) {
				metadata.EXPECT().
					GetVideo(gomock.Any(), "video-789").
					Return(&domain.VideoMetadata{
						ID:         "video-789",
						BucketName: "videos",
						ObjectKey:  "uuid/video.mp4",
					}, nil)

				storage.EXPECT().
					PresignedGetObject(gomock.Any(), "videos", "uuid/video.mp4", gomock.Any()).
					Return(nil, errors.New("storage unavailable"))
			},
			wantErr: true,
		},
		{
			name:          "error - metadata service unavailable",
			videoID:       "video-123",
			defaultBucket: "default-bucket",
			setupMock: func(storage *mocks.MockStorageService, metadata *mocks.MockMetadataService) {
				metadata.EXPECT().
					GetVideo(gomock.Any(), "video-123").
					Return(nil, errors.New("connection refused"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := mocks.NewMockStorageService(ctrl)
			mockMetadata := mocks.NewMockMetadataService(ctrl)
			tt.setupMock(mockStorage, mockMetadata)

			uc := NewStreamingUsecase(mockStorage, mockMetadata, tt.defaultBucket)
			gotURL, err := uc.GetStreamURL(context.Background(), tt.videoID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetStreamURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && gotURL != tt.wantURL {
				t.Errorf("GetStreamURL() = %v, want %v", gotURL, tt.wantURL)
			}
		})
	}
}

func TestStreamingUsecase_GetStreamURL_ContextPropagation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorageService(ctrl)
	mockMetadata := mocks.NewMockMetadataService(ctrl)

	ctx := context.WithValue(context.Background(), "test-key", "test-value")

	// Verify context is propagated correctly
	mockMetadata.EXPECT().
		GetVideo(ctx, "video-123").
		Return(&domain.VideoMetadata{
			ID:         "video-123",
			BucketName: "videos",
			ObjectKey:  "test.mp4",
		}, nil)

	presignedURL, _ := url.Parse("https://example.com/presigned")
	mockStorage.EXPECT().
		PresignedGetObject(ctx, "videos", "test.mp4", gomock.Any()).
		Return(presignedURL, nil)

	uc := NewStreamingUsecase(mockStorage, mockMetadata, "default")
	_, err := uc.GetStreamURL(ctx, "video-123")

	if err != nil {
		t.Errorf("GetStreamURL() unexpected error: %v", err)
	}
}

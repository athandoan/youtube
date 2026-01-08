package usecase

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/athandoan/youtube/upload-service/internal/mocks"
	"go.uber.org/mock/gomock"
)

func TestUploadUsecase_InitUpload(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		filename  string
		setupMock func(storage *mocks.MockStorageService, metadata *mocks.MockMetadataService)
		wantErr   bool
		checkURL  bool
	}{
		{
			name:     "success - initializes upload with presigned URL",
			title:    "My Video",
			filename: "video.mp4",
			setupMock: func(storage *mocks.MockStorageService, metadata *mocks.MockMetadataService) {
				metadata.EXPECT().
					CreateVideo(gomock.Any(), "My Video", "videos", gomock.Any()).
					Return("video-123", nil)

				presignedURL, _ := url.Parse("https://s3.example.com/videos/uuid/video.mp4?signature=xxx")
				storage.EXPECT().
					PresignedPutObject(gomock.Any(), "videos", gomock.Any(), time.Hour).
					Return(presignedURL, nil)
			},
			wantErr:  false,
			checkURL: true,
		},
		{
			name:     "error - metadata service fails to create video",
			title:    "My Video",
			filename: "video.mp4",
			setupMock: func(storage *mocks.MockStorageService, metadata *mocks.MockMetadataService) {
				metadata.EXPECT().
					CreateVideo(gomock.Any(), "My Video", "videos", gomock.Any()).
					Return("", errors.New("metadata service unavailable"))
			},
			wantErr: true,
		},
		{
			name:     "error - storage service fails to generate presigned URL",
			title:    "My Video",
			filename: "video.mp4",
			setupMock: func(storage *mocks.MockStorageService, metadata *mocks.MockMetadataService) {
				metadata.EXPECT().
					CreateVideo(gomock.Any(), "My Video", "videos", gomock.Any()).
					Return("video-123", nil)

				storage.EXPECT().
					PresignedPutObject(gomock.Any(), "videos", gomock.Any(), time.Hour).
					Return(nil, errors.New("storage error"))
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

			uc := NewUploadUsecase(mockStorage, mockMetadata, "videos")
			videoID, presignedURL, err := uc.InitUpload(context.Background(), tt.title, tt.filename)

			if (err != nil) != tt.wantErr {
				t.Errorf("InitUpload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if videoID == "" {
					t.Error("InitUpload() returned empty videoID")
				}
				if tt.checkURL && presignedURL == "" {
					t.Error("InitUpload() returned empty presignedURL")
				}
			}
		})
	}
}

func TestUploadUsecase_InitUpload_ObjectKeyFormat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorageService(ctrl)
	mockMetadata := mocks.NewMockMetadataService(ctrl)

	var capturedObjectKey string

	// Capture the object key to verify format
	mockMetadata.EXPECT().
		CreateVideo(gomock.Any(), "Test Video", "test-bucket", gomock.Any()).
		DoAndReturn(func(ctx context.Context, title, bucket, objectKey string) (string, error) {
			capturedObjectKey = objectKey
			return "video-123", nil
		})

	presignedURL, _ := url.Parse("https://s3.example.com/presigned")
	mockStorage.EXPECT().
		PresignedPutObject(gomock.Any(), "test-bucket", gomock.Any(), gomock.Any()).
		Return(presignedURL, nil)

	uc := NewUploadUsecase(mockStorage, mockMetadata, "test-bucket")
	_, _, err := uc.InitUpload(context.Background(), "Test Video", "original-filename.mp4")

	if err != nil {
		t.Fatalf("InitUpload() unexpected error: %v", err)
	}

	// Verify object key format: should be {uuid}/{filename}
	// UUID is 36 characters + "/" + filename
	if len(capturedObjectKey) <= 37 {
		t.Errorf("ObjectKey too short, expected UUID/filename format, got: %s", capturedObjectKey)
	}

	// Should contain the original filename
	if capturedObjectKey[37:] != "original-filename.mp4" {
		t.Errorf("ObjectKey should end with original filename, got: %s", capturedObjectKey)
	}
}

func TestUploadUsecase_CompleteUpload(t *testing.T) {
	tests := []struct {
		name      string
		videoID   string
		setupMock func(metadata *mocks.MockMetadataService)
		wantErr   bool
	}{
		{
			name:    "success - marks video as ready",
			videoID: "video-123",
			setupMock: func(metadata *mocks.MockMetadataService) {
				metadata.EXPECT().
					UpdateVideoStatus(gomock.Any(), "video-123", "ready").
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "error - video not found",
			videoID: "nonexistent-id",
			setupMock: func(metadata *mocks.MockMetadataService) {
				metadata.EXPECT().
					UpdateVideoStatus(gomock.Any(), "nonexistent-id", "ready").
					Return(errors.New("video not found"))
			},
			wantErr: true,
		},
		{
			name:    "error - metadata service unavailable",
			videoID: "video-123",
			setupMock: func(metadata *mocks.MockMetadataService) {
				metadata.EXPECT().
					UpdateVideoStatus(gomock.Any(), "video-123", "ready").
					Return(errors.New("connection refused"))
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
			tt.setupMock(mockMetadata)

			uc := NewUploadUsecase(mockStorage, mockMetadata, "videos")
			err := uc.CompleteUpload(context.Background(), tt.videoID)

			if (err != nil) != tt.wantErr {
				t.Errorf("CompleteUpload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

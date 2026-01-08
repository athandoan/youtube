package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/athandoan/youtube/gateway-service/internal/mocks"
	"github.com/athandoan/youtube/proto/common"
	"go.uber.org/mock/gomock"
)

func TestGatewayUsecase_InitUpload(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		filename  string
		setupMock func(upload *mocks.MockUploadService)
		wantID    string
		wantURL   string
		wantErr   bool
	}{
		{
			name:     "success - initializes upload",
			title:    "My Video",
			filename: "video.mp4",
			setupMock: func(upload *mocks.MockUploadService) {
				upload.EXPECT().
					InitUpload(gomock.Any(), "My Video", "video.mp4").
					Return("video-123", "https://presigned-url.example.com", nil)
			},
			wantID:  "video-123",
			wantURL: "https://presigned-url.example.com",
			wantErr: false,
		},
		{
			name:     "error - upload service fails",
			title:    "My Video",
			filename: "video.mp4",
			setupMock: func(upload *mocks.MockUploadService) {
				upload.EXPECT().
					InitUpload(gomock.Any(), "My Video", "video.mp4").
					Return("", "", errors.New("upload service unavailable"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMetadata := mocks.NewMockMetadataService(ctrl)
			mockUpload := mocks.NewMockUploadService(ctrl)
			mockStreaming := mocks.NewMockStreamingService(ctrl)
			tt.setupMock(mockUpload)

			uc := NewGatewayUsecase(mockMetadata, mockUpload, mockStreaming)
			resp, err := uc.InitUpload(context.Background(), tt.title, tt.filename)

			if (err != nil) != tt.wantErr {
				t.Errorf("InitUpload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if resp.VideoId != tt.wantID {
					t.Errorf("InitUpload() VideoId = %v, want %v", resp.VideoId, tt.wantID)
				}
				if resp.PresignedUrl != tt.wantURL {
					t.Errorf("InitUpload() PresignedUrl = %v, want %v", resp.PresignedUrl, tt.wantURL)
				}
			}
		})
	}
}

func TestGatewayUsecase_CompleteUpload(t *testing.T) {
	tests := []struct {
		name      string
		videoID   string
		setupMock func(upload *mocks.MockUploadService)
		wantErr   bool
	}{
		{
			name:    "success - completes upload",
			videoID: "video-123",
			setupMock: func(upload *mocks.MockUploadService) {
				upload.EXPECT().
					CompleteUpload(gomock.Any(), "video-123").
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "error - upload service fails",
			videoID: "video-123",
			setupMock: func(upload *mocks.MockUploadService) {
				upload.EXPECT().
					CompleteUpload(gomock.Any(), "video-123").
					Return(errors.New("video not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMetadata := mocks.NewMockMetadataService(ctrl)
			mockUpload := mocks.NewMockUploadService(ctrl)
			mockStreaming := mocks.NewMockStreamingService(ctrl)
			tt.setupMock(mockUpload)

			uc := NewGatewayUsecase(mockMetadata, mockUpload, mockStreaming)
			resp, err := uc.CompleteUpload(context.Background(), tt.videoID)

			if (err != nil) != tt.wantErr {
				t.Errorf("CompleteUpload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && resp.Status != "success" {
				t.Errorf("CompleteUpload() Status = %v, want 'success'", resp.Status)
			}
		})
	}
}

func TestGatewayUsecase_ListVideos(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		setupMock func(metadata *mocks.MockMetadataService)
		wantCount int
		wantErr   bool
	}{
		{
			name:  "success - returns all videos",
			query: "",
			setupMock: func(metadata *mocks.MockMetadataService) {
				metadata.EXPECT().
					ListVideos(gomock.Any(), "").
					Return([]*common.Video{
						{Id: "video-1", Title: "Video 1"},
						{Id: "video-2", Title: "Video 2"},
						{Id: "video-3", Title: "Video 3"},
					}, nil)
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:  "success - returns filtered videos",
			query: "golang",
			setupMock: func(metadata *mocks.MockMetadataService) {
				metadata.EXPECT().
					ListVideos(gomock.Any(), "golang").
					Return([]*common.Video{
						{Id: "video-1", Title: "Golang Tutorial"},
					}, nil)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:  "success - returns empty list",
			query: "nonexistent",
			setupMock: func(metadata *mocks.MockMetadataService) {
				metadata.EXPECT().
					ListVideos(gomock.Any(), "nonexistent").
					Return([]*common.Video{}, nil)
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:  "error - metadata service fails",
			query: "",
			setupMock: func(metadata *mocks.MockMetadataService) {
				metadata.EXPECT().
					ListVideos(gomock.Any(), "").
					Return(nil, errors.New("metadata service unavailable"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMetadata := mocks.NewMockMetadataService(ctrl)
			mockUpload := mocks.NewMockUploadService(ctrl)
			mockStreaming := mocks.NewMockStreamingService(ctrl)
			tt.setupMock(mockMetadata)

			uc := NewGatewayUsecase(mockMetadata, mockUpload, mockStreaming)
			videos, err := uc.ListVideos(context.Background(), tt.query)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListVideos() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(videos) != tt.wantCount {
				t.Errorf("ListVideos() returned %d videos, want %d", len(videos), tt.wantCount)
			}
		})
	}
}

func TestGatewayUsecase_GetStreamURL(t *testing.T) {
	tests := []struct {
		name      string
		videoID   string
		setupMock func(streaming *mocks.MockStreamingService)
		wantURL   string
		wantErr   bool
	}{
		{
			name:    "success - returns stream URL",
			videoID: "video-123",
			setupMock: func(streaming *mocks.MockStreamingService) {
				streaming.EXPECT().
					GetStreamURL(gomock.Any(), "video-123").
					Return("https://stream.example.com/video-123?signature=xxx", nil)
			},
			wantURL: "https://stream.example.com/video-123?signature=xxx",
			wantErr: false,
		},
		{
			name:    "error - video not found",
			videoID: "nonexistent-id",
			setupMock: func(streaming *mocks.MockStreamingService) {
				streaming.EXPECT().
					GetStreamURL(gomock.Any(), "nonexistent-id").
					Return("", errors.New("video not found"))
			},
			wantErr: true,
		},
		{
			name:    "error - streaming service unavailable",
			videoID: "video-123",
			setupMock: func(streaming *mocks.MockStreamingService) {
				streaming.EXPECT().
					GetStreamURL(gomock.Any(), "video-123").
					Return("", errors.New("connection refused"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMetadata := mocks.NewMockMetadataService(ctrl)
			mockUpload := mocks.NewMockUploadService(ctrl)
			mockStreaming := mocks.NewMockStreamingService(ctrl)
			tt.setupMock(mockStreaming)

			uc := NewGatewayUsecase(mockMetadata, mockUpload, mockStreaming)
			url, err := uc.GetStreamURL(context.Background(), tt.videoID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetStreamURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && url != tt.wantURL {
				t.Errorf("GetStreamURL() = %v, want %v", url, tt.wantURL)
			}
		})
	}
}

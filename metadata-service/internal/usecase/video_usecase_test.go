package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/athandoan/youtube/metadata-service/internal/domain"
	"github.com/athandoan/youtube/metadata-service/internal/mocks"
	"go.uber.org/mock/gomock"
)

func TestVideoUsecase_Create(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		bucket    string
		objectKey string
		setupMock func(m *mocks.MockVideoRepository)
		wantErr   bool
		wantIDLen int
	}{
		{
			name:      "success - creates video with valid data",
			title:     "Test Video",
			bucket:    "videos",
			objectKey: "uuid/test.mp4",
			setupMock: func(m *mocks.MockVideoRepository) {
				m.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, v *domain.Video) error {
						if v.Title != "Test Video" {
							t.Errorf("expected title 'Test Video', got %s", v.Title)
						}
						if v.BucketName != "videos" {
							t.Errorf("expected bucket 'videos', got %s", v.BucketName)
						}
						if v.ObjectKey != "uuid/test.mp4" {
							t.Errorf("expected objectKey 'uuid/test.mp4', got %s", v.ObjectKey)
						}
						if v.Status != "pending" {
							t.Errorf("expected status 'pending', got %s", v.Status)
						}
						return nil
					})
			},
			wantErr:   false,
			wantIDLen: 36, // UUID length
		},
		{
			name:      "error - repository fails",
			title:     "Test Video",
			bucket:    "videos",
			objectKey: "uuid/test.mp4",
			setupMock: func(m *mocks.MockVideoRepository) {
				m.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockVideoRepository(ctrl)
			tt.setupMock(mockRepo)

			uc := NewVideoUsecase(mockRepo)
			id, err := uc.Create(context.Background(), tt.title, tt.bucket, tt.objectKey)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(id) != tt.wantIDLen {
				t.Errorf("Create() returned ID with length %d, want %d", len(id), tt.wantIDLen)
			}
		})
	}
}

func TestVideoUsecase_Get(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		setupMock func(m *mocks.MockVideoRepository)
		want      *domain.Video
		wantErr   bool
	}{
		{
			name: "success - returns video",
			id:   "video-123",
			setupMock: func(m *mocks.MockVideoRepository) {
				m.EXPECT().
					Get(gomock.Any(), "video-123").
					Return(&domain.Video{
						ID:         "video-123",
						Title:      "Test Video",
						BucketName: "videos",
						ObjectKey:  "uuid/test.mp4",
						Status:     "ready",
					}, nil)
			},
			want: &domain.Video{
				ID:         "video-123",
				Title:      "Test Video",
				BucketName: "videos",
				ObjectKey:  "uuid/test.mp4",
				Status:     "ready",
			},
			wantErr: false,
		},
		{
			name: "error - video not found",
			id:   "nonexistent-id",
			setupMock: func(m *mocks.MockVideoRepository) {
				m.EXPECT().
					Get(gomock.Any(), "nonexistent-id").
					Return(nil, errors.New("video not found"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockVideoRepository(ctrl)
			tt.setupMock(mockRepo)

			uc := NewVideoUsecase(mockRepo)
			got, err := uc.Get(context.Background(), tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.ID != tt.want.ID {
					t.Errorf("Get() got ID = %v, want %v", got.ID, tt.want.ID)
				}
				if got.Title != tt.want.Title {
					t.Errorf("Get() got Title = %v, want %v", got.Title, tt.want.Title)
				}
			}
		})
	}
}

func TestVideoUsecase_List(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		setupMock func(m *mocks.MockVideoRepository)
		wantCount int
		wantErr   bool
	}{
		{
			name:  "success - returns all videos when query is empty",
			query: "",
			setupMock: func(m *mocks.MockVideoRepository) {
				m.EXPECT().
					List(gomock.Any(), "").
					Return([]*domain.Video{
						{ID: "video-1", Title: "Video 1"},
						{ID: "video-2", Title: "Video 2"},
					}, nil)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:  "success - returns filtered videos",
			query: "golang",
			setupMock: func(m *mocks.MockVideoRepository) {
				m.EXPECT().
					List(gomock.Any(), "golang").
					Return([]*domain.Video{
						{ID: "video-1", Title: "Golang Tutorial"},
					}, nil)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:  "success - returns empty list",
			query: "nonexistent",
			setupMock: func(m *mocks.MockVideoRepository) {
				m.EXPECT().
					List(gomock.Any(), "nonexistent").
					Return([]*domain.Video{}, nil)
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:  "error - repository fails",
			query: "",
			setupMock: func(m *mocks.MockVideoRepository) {
				m.EXPECT().
					List(gomock.Any(), "").
					Return(nil, errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockVideoRepository(ctrl)
			tt.setupMock(mockRepo)

			uc := NewVideoUsecase(mockRepo)
			got, err := uc.List(context.Background(), tt.query)

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(got) != tt.wantCount {
				t.Errorf("List() returned %d videos, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestVideoUsecase_UpdateStatus(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		status    string
		setupMock func(m *mocks.MockVideoRepository)
		wantErr   bool
	}{
		{
			name:   "success - updates status to ready",
			id:     "video-123",
			status: "ready",
			setupMock: func(m *mocks.MockVideoRepository) {
				m.EXPECT().
					UpdateStatus(gomock.Any(), "video-123", "ready").
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "success - updates status to error",
			id:     "video-123",
			status: "error",
			setupMock: func(m *mocks.MockVideoRepository) {
				m.EXPECT().
					UpdateStatus(gomock.Any(), "video-123", "error").
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "error - video not found",
			id:     "nonexistent-id",
			status: "ready",
			setupMock: func(m *mocks.MockVideoRepository) {
				m.EXPECT().
					UpdateStatus(gomock.Any(), "nonexistent-id", "ready").
					Return(errors.New("video not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockVideoRepository(ctrl)
			tt.setupMock(mockRepo)

			uc := NewVideoUsecase(mockRepo)
			err := uc.UpdateStatus(context.Background(), tt.id, tt.status)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

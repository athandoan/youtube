package domain

//go:generate mockgen -source=gateway.go -destination=../mocks/mock_services.go -package=mocks

import (
	"context"

	"github.com/athandoan/youtube/proto/common"
	uploadpb "github.com/athandoan/youtube/proto/upload"
)

type MetadataService interface {
	ListVideos(ctx context.Context, query string) ([]*common.Video, error)
}

type UploadService interface {
	InitUpload(ctx context.Context, title, filename string) (string, string, error)
	CompleteUpload(ctx context.Context, videoID string) error
}

type StreamingService interface {
	GetStreamURL(ctx context.Context, videoID string) (string, error)
}

type GatewayUsecase interface {
	InitUpload(ctx context.Context, title, filename string) (*uploadpb.InitUploadResponse, error)
	CompleteUpload(ctx context.Context, videoID string) (*uploadpb.CompleteUploadResponse, error)
	ListVideos(ctx context.Context, query string) ([]*common.Video, error)
	GetStreamURL(ctx context.Context, videoID string) (string, error)
}

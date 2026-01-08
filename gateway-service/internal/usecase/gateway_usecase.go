package usecase

import (
	"context"

	"github.com/athandoan/youtube/gateway-service/internal/domain"
	"github.com/athandoan/youtube/proto/common"
	uploadpb "github.com/athandoan/youtube/proto/upload"
)

type gatewayUsecase struct {
	metadata  domain.MetadataService
	upload    domain.UploadService
	streaming domain.StreamingService
}

func NewGatewayUsecase(metadata domain.MetadataService, upload domain.UploadService, streaming domain.StreamingService) domain.GatewayUsecase {
	return &gatewayUsecase{metadata: metadata, upload: upload, streaming: streaming}
}

func (u *gatewayUsecase) InitUpload(ctx context.Context, title, filename string) (*uploadpb.InitUploadResponse, error) {
	id, url, err := u.upload.InitUpload(ctx, title, filename)
	if err != nil {
		return nil, err
	}
	return &uploadpb.InitUploadResponse{
		VideoId:      id,
		PresignedUrl: url,
	}, nil
}

func (u *gatewayUsecase) CompleteUpload(ctx context.Context, videoID string) (*uploadpb.CompleteUploadResponse, error) {
	err := u.upload.CompleteUpload(ctx, videoID)
	if err != nil {
		return nil, err
	}
	return &uploadpb.CompleteUploadResponse{Status: "success"}, nil
}

func (u *gatewayUsecase) ListVideos(ctx context.Context, query string) ([]*common.Video, error) {
	return u.metadata.ListVideos(ctx, query)
}

func (u *gatewayUsecase) GetStreamURL(ctx context.Context, videoID string) (string, error) {
	return u.streaming.GetStreamURL(ctx, videoID)
}

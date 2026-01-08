package grpc

import (
	"context"

	pb "github.com/athandoan/youtube/proto/upload"
	"github.com/athandoan/youtube/upload-service/internal/domain"
)

type UploadHandler struct {
	pb.UnimplementedUploadServiceServer
	Usecase domain.UploadUsecase
}

func NewUploadHandler(u domain.UploadUsecase) *UploadHandler {
	return &UploadHandler{Usecase: u}
}

func (h *UploadHandler) InitUpload(ctx context.Context, req *pb.InitUploadRequest) (*pb.InitUploadResponse, error) {
	videoID, url, err := h.Usecase.InitUpload(ctx, req.Title, req.Filename)
	if err != nil {
		return nil, err
	}
	return &pb.InitUploadResponse{
		VideoId:      videoID,
		PresignedUrl: url,
	}, nil
}

func (h *UploadHandler) CompleteUpload(ctx context.Context, req *pb.CompleteUploadRequest) (*pb.CompleteUploadResponse, error) {
	err := h.Usecase.CompleteUpload(ctx, req.VideoId)
	if err != nil {
		return nil, err
	}
	return &pb.CompleteUploadResponse{Status: "success"}, nil
}

package grpc

import (
	"context"

	"github.com/athandoan/youtube/metadata-service/internal/domain"
	"github.com/athandoan/youtube/proto/common"
	pb "github.com/athandoan/youtube/proto/metadata"
)

type MetadataHandler struct {
	pb.UnimplementedMetadataServiceServer
	Usecase domain.VideoUsecase
}

func NewMetadataHandler(u domain.VideoUsecase) *MetadataHandler {
	return &MetadataHandler{Usecase: u}
}

func (h *MetadataHandler) CreateVideo(ctx context.Context, req *pb.CreateVideoRequest) (*pb.CreateVideoResponse, error) {
	id, err := h.Usecase.Create(ctx, req.Title, req.Bucket, req.ObjectKey)
	if err != nil {
		return nil, err
	}
	return &pb.CreateVideoResponse{Id: id}, nil
}

func (h *MetadataHandler) ListVideos(ctx context.Context, req *pb.ListVideosRequest) (*pb.ListVideosResponse, error) {
	videos, err := h.Usecase.List(ctx, req.Query)
	if err != nil {
		return nil, err
	}

	var pbVideos []*common.Video
	for _, v := range videos {
		pbVideos = append(pbVideos, &common.Video{
			Id:         v.ID,
			Title:      v.Title,
			Status:     v.Status,
			CreatedAt:  v.CreatedAt.Format("2006-01-02 15:04:05"),
			BucketName: v.BucketName,
			ObjectKey:  v.ObjectKey,
		})
	}
	return &pb.ListVideosResponse{Videos: pbVideos}, nil
}

func (h *MetadataHandler) UpdateVideoStatus(ctx context.Context, req *pb.UpdateVideoStatusRequest) (*pb.UpdateVideoStatusResponse, error) {
	err := h.Usecase.UpdateStatus(ctx, req.Id, req.Status)
	if err != nil {
		return nil, err
	}
	return &pb.UpdateVideoStatusResponse{Status: "success"}, nil
}

func (h *MetadataHandler) GetVideo(ctx context.Context, req *pb.GetVideoRequest) (*common.Video, error) {
	v, err := h.Usecase.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &common.Video{
		Id:         v.ID,
		Title:      v.Title,
		Status:     v.Status,
		CreatedAt:  v.CreatedAt.Format("2006-01-02 15:04:05"),
		BucketName: v.BucketName,
		ObjectKey:  v.ObjectKey,
	}, nil
}

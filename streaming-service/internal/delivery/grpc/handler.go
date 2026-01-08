package grpc

import (
	"context"

	pb "github.com/athandoan/youtube/proto/streaming"
	"github.com/athandoan/youtube/streaming-service/internal/domain"
)

type StreamingHandler struct {
	pb.UnimplementedStreamingServiceServer
	usecase domain.StreamingUsecase
}

func NewStreamingHandler(u domain.StreamingUsecase) *StreamingHandler {
	return &StreamingHandler{usecase: u}
}

func (h *StreamingHandler) GetStreamURL(ctx context.Context, req *pb.GetStreamURLRequest) (*pb.GetStreamURLResponse, error) {
	url, err := h.usecase.GetStreamURL(ctx, req.VideoId)
	if err != nil {
		return nil, err
	}
	return &pb.GetStreamURLResponse{Url: url}, nil
}

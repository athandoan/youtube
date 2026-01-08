package rpc

import (
	"context"

	"github.com/athandoan/youtube/gateway-service/internal/domain"
	streamingpb "github.com/athandoan/youtube/proto/streaming"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type streamingClient struct {
	client streamingpb.StreamingServiceClient
	conn   *grpc.ClientConn
}

func NewStreamingClient(addr string) (domain.StreamingService, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	client := streamingpb.NewStreamingServiceClient(conn)
	return &streamingClient{client: client, conn: conn}, nil
}

func (s *streamingClient) GetStreamURL(ctx context.Context, videoID string) (string, error) {
	resp, err := s.client.GetStreamURL(ctx, &streamingpb.GetStreamURLRequest{
		VideoId: videoID,
	})
	if err != nil {
		return "", err
	}
	return resp.Url, nil
}

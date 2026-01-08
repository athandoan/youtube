package rpc

import (
	"context"

	pb "github.com/athandoan/youtube/proto/metadata"
	"github.com/athandoan/youtube/streaming-service/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type metadataClient struct {
	client pb.MetadataServiceClient
	conn   *grpc.ClientConn
}

func NewMetadataClient(addr string) (domain.MetadataService, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	client := pb.NewMetadataServiceClient(conn)
	return &metadataClient{client: client, conn: conn}, nil
}

func (m *metadataClient) GetVideo(ctx context.Context, id string) (*domain.VideoMetadata, error) {
	resp, err := m.client.GetVideo(ctx, &pb.GetVideoRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return &domain.VideoMetadata{
		ID:         resp.Id,
		BucketName: resp.BucketName,
		ObjectKey:  resp.ObjectKey,
	}, nil
}

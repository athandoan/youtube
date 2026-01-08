package rpc

import (
	"context"

	pb "github.com/athandoan/youtube/proto/metadata"
	"github.com/athandoan/youtube/upload-service/internal/domain"
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

func (m *metadataClient) CreateVideo(ctx context.Context, title, bucket, objectKey string) (string, error) {
	resp, err := m.client.CreateVideo(ctx, &pb.CreateVideoRequest{
		Title:     title,
		Bucket:    bucket,
		ObjectKey: objectKey,
	})
	if err != nil {
		return "", err
	}
	return resp.Id, nil
}

func (m *metadataClient) UpdateVideoStatus(ctx context.Context, id, status string) error {
	_, err := m.client.UpdateVideoStatus(ctx, &pb.UpdateVideoStatusRequest{
		Id:     id,
		Status: status,
	})
	return err
}

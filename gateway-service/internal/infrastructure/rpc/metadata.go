package rpc

import (
	"context"

	"github.com/athandoan/youtube/gateway-service/internal/domain"
	"github.com/athandoan/youtube/proto/common"
	metadatapb "github.com/athandoan/youtube/proto/metadata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type metadataClient struct {
	client metadatapb.MetadataServiceClient
	conn   *grpc.ClientConn
}

func NewMetadataClient(addr string) (domain.MetadataService, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	client := metadatapb.NewMetadataServiceClient(conn)
	return &metadataClient{client: client, conn: conn}, nil
}

func (m *metadataClient) ListVideos(ctx context.Context, query string) ([]*common.Video, error) {
	resp, err := m.client.ListVideos(ctx, &metadatapb.ListVideosRequest{Query: query})
	if err != nil {
		return nil, err
	}
	return resp.Videos, nil
}

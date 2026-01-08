package rpc

import (
	"context"

	"github.com/athandoan/youtube/gateway-service/internal/domain"
	uploadpb "github.com/athandoan/youtube/proto/upload"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type uploadClient struct {
	client uploadpb.UploadServiceClient
	conn   *grpc.ClientConn
}

func NewUploadClient(addr string) (domain.UploadService, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	client := uploadpb.NewUploadServiceClient(conn)
	return &uploadClient{client: client, conn: conn}, nil
}

func (u *uploadClient) InitUpload(ctx context.Context, title, filename string) (string, string, error) {
	resp, err := u.client.InitUpload(ctx, &uploadpb.InitUploadRequest{
		Title:    title,
		Filename: filename,
	})
	if err != nil {
		return "", "", err
	}
	return resp.VideoId, resp.PresignedUrl, nil
}

func (u *uploadClient) CompleteUpload(ctx context.Context, videoID string) error {
	_, err := u.client.CompleteUpload(ctx, &uploadpb.CompleteUploadRequest{
		VideoId: videoID,
	})
	return err
}

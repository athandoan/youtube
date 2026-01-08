package main

import (
	"log"
	"net"
	"os"

	pb "github.com/athandoan/youtube/proto/streaming"
	handler "github.com/athandoan/youtube/streaming-service/internal/delivery/grpc"
	"github.com/athandoan/youtube/streaming-service/internal/infrastructure/rpc"
	"github.com/athandoan/youtube/streaming-service/internal/infrastructure/storage"
	"github.com/athandoan/youtube/streaming-service/internal/usecase"
	"google.golang.org/grpc"
)

func main() {
	// 1. Init MinIO
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"
	bucketName := os.Getenv("MINIO_BUCKET")
	minioEndpoint := os.Getenv("MINIO_ENDPOINT") // Internal fallback

	// External endpoint for presigned URLs (accessible from browser)
	externalEndpoint := os.Getenv("S3_EXTERNAL_ENDPOINT")
	if externalEndpoint == "" {
		externalEndpoint = minioEndpoint
	}

	storageService, err := storage.NewMinioStorage(externalEndpoint, minioAccessKey, minioSecretKey, useSSL, "us-east-1")
	if err != nil {
		log.Fatalf("failed to create storage service: %v", err)
	}

	// 2. Init Metadata Client (gRPC)
	metaAddr := os.Getenv("METADATA_SERVICE_ADDR")
	if metaAddr == "" {
		metaAddr = "metadata-service:50051"
	}
	metadataService, err := rpc.NewMetadataClient(metaAddr)
	if err != nil {
		log.Fatalf("failed to create metadata client: %v", err)
	}

	// 3. Init Usecase
	uc := usecase.NewStreamingUsecase(storageService, metadataService, bucketName)

	// 4. Init gRPC Handler
	h := handler.NewStreamingHandler(uc)

	// 5. Start gRPC Server
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50053"
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterStreamingServiceServer(s, h)

	log.Printf("Streaming Service (gRPC) running on :%s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

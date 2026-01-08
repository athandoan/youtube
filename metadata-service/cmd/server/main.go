package main

import (
	"log"
	"net"
	"os"

	handler "github.com/athandoan/youtube/metadata-service/internal/delivery/grpc"
	"github.com/athandoan/youtube/metadata-service/internal/repository"
	"github.com/athandoan/youtube/metadata-service/internal/usecase"
	pb "github.com/athandoan/youtube/proto/metadata"
	"google.golang.org/grpc"
)

func main() {
	// 1. Init SQLite DB
	dbPath := os.Getenv("SQLITE_DB_PATH")
	if dbPath == "" {
		dbPath = "metadata.db"
	}

	repo, err := repository.NewSQLiteRepository(dbPath)
	if err != nil {
		log.Fatalf("failed to init repository: %v", err)
	}

	// 2. Init Usecase
	uc := usecase.NewVideoUsecase(repo)

	// 3. Init Handler
	h := handler.NewMetadataHandler(uc)

	// 4. Start gRPC Server
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50051"
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterMetadataServiceServer(s, h)

	log.Printf("Metadata Service (gRPC) running on :%s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

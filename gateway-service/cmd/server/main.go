package main

import (
	"log"
	"net/http"
	"os"

	handler "github.com/athandoan/youtube/gateway-service/internal/delivery/http"
	"github.com/athandoan/youtube/gateway-service/internal/infrastructure/rpc"
	"github.com/athandoan/youtube/gateway-service/internal/usecase"
)

func main() {
	// 1. Connect to Metadata Service
	metaAddr := os.Getenv("METADATA_SERVICE_ADDR")
	if metaAddr == "" {
		metaAddr = "metadata-service:50051"
	}
	metadataClient, err := rpc.NewMetadataClient(metaAddr)
	if err != nil {
		log.Fatalf("did not connect to metadata: %v", err)
	}

	// 2. Connect to Upload Service
	uploadAddr := os.Getenv("UPLOAD_SERVICE_ADDR")
	if uploadAddr == "" {
		uploadAddr = "upload-service:50052"
	}
	uploadClient, err := rpc.NewUploadClient(uploadAddr)
	if err != nil {
		log.Fatalf("did not connect to upload: %v", err)
	}

	// 3. Connect to Streaming Service
	streamAddr := os.Getenv("STREAMING_SERVICE_ADDR")
	if streamAddr == "" {
		streamAddr = "streaming-service:50053"
	}
	streamingClient, err := rpc.NewStreamingClient(streamAddr)
	if err != nil {
		log.Fatalf("did not connect to streaming: %v", err)
	}

	// 4. Init Usecase
	uc := usecase.NewGatewayUsecase(metadataClient, uploadClient, streamingClient)

	// 5. Init Handler
	h := handler.NewHandler(uc)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/upload/init", h.HandleInitUpload)
	mux.HandleFunc("/api/upload/complete", h.HandleCompleteUpload)
	mux.HandleFunc("/api/videos", h.HandleListVideos)
	mux.HandleFunc("/api/stream/videos/", h.HandleStreamVideo)

	// CORS middleware
	hMux := handler.CorsMiddleware(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Gateway Service running on :%s", port)
	if err := http.ListenAndServe(":"+port, hMux); err != nil {
		log.Fatal(err)
	}
}

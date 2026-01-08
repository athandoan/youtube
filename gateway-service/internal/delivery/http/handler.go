package http

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/athandoan/youtube/gateway-service/internal/domain"
	uploadpb "github.com/athandoan/youtube/proto/upload"
	"github.com/google/jsonapi"
)

type Handler struct {
	usecase domain.GatewayUsecase
}

func NewHandler(u domain.GatewayUsecase) *Handler {
	return &Handler{usecase: u}
}

// JSON:API Structures
type InitUploadData struct {
	ID           string `jsonapi:"primary,upload-init"`
	PresignedUrl string `jsonapi:"attr,presigned_url"`
}

type CompleteUploadData struct {
	ID     string `jsonapi:"primary,upload-status"`
	Status string `jsonapi:"attr,status"`
}

type VideoResponse struct {
	ID         string `jsonapi:"primary,video"`
	Title      string `jsonapi:"attr,title"`
	Status     string `jsonapi:"attr,status"`
	CreatedAt  string `jsonapi:"attr,created_at"`
	BucketName string `jsonapi:"attr,bucket_name"`
	ObjectKey  string `jsonapi:"attr,object_key"`
}

func writeJsonApi(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", jsonapi.MediaType)
	w.WriteHeader(http.StatusOK)

	if err := jsonapi.MarshalPayload(w, data); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeJsonApiError(w http.ResponseWriter, statusCode int, title, detail string) {
	w.Header().Set("Content-Type", jsonapi.MediaType)
	w.WriteHeader(statusCode)

	jsonApiErrors := []*jsonapi.ErrorObject{{
		Status: strconv.Itoa(statusCode),
		Title:  title,
		Detail: detail,
	}}

	if err := jsonapi.MarshalErrors(w, jsonApiErrors); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}

func (h *Handler) HandleInitUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeJsonApiError(w, http.StatusMethodNotAllowed, "Method Not Allowed", "Only POST is allowed")
		return
	}

	var req uploadpb.InitUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJsonApiError(w, http.StatusBadRequest, "Invalid Request", err.Error())
		return
	}

	resp, err := h.usecase.InitUpload(r.Context(), req.Title, req.Filename)
	if err != nil {
		writeJsonApiError(w, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	data := &InitUploadData{
		ID:           resp.VideoId,
		PresignedUrl: resp.PresignedUrl,
	}
	writeJsonApi(w, data)
}

func (h *Handler) HandleCompleteUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeJsonApiError(w, http.StatusMethodNotAllowed, "Method Not Allowed", "Only POST is allowed")
		return
	}

	var req uploadpb.CompleteUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJsonApiError(w, http.StatusBadRequest, "Invalid Request", err.Error())
		return
	}
	log.Printf("CompleteUpload request for VideoID: %s", req.VideoId)

	_, err := h.usecase.CompleteUpload(r.Context(), req.VideoId)
	if err != nil {
		writeJsonApiError(w, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	data := &CompleteUploadData{
		ID:     req.VideoId,
		Status: "success",
	}
	writeJsonApi(w, data)
}

func (h *Handler) HandleListVideos(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeJsonApiError(w, http.StatusMethodNotAllowed, "Method Not Allowed", "Only GET is allowed")
		return
	}

	query := r.URL.Query().Get("q")
	videos, err := h.usecase.ListVideos(r.Context(), query)
	if err != nil {
		writeJsonApiError(w, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	data := make([]*VideoResponse, 0)
	for _, v := range videos {
		data = append(data, &VideoResponse{
			ID:         v.Id,
			Title:      v.Title,
			Status:     v.Status,
			CreatedAt:  v.CreatedAt,
			BucketName: v.BucketName,
			ObjectKey:  v.ObjectKey,
		})
	}

	// jsonapi.MarshalPayload creates an empty data array for nil/empty slice
	// but let's be safe and ensure it is at least an empty slice if that's what we want
	// The library handles empty slices correctly as `[]`.
	writeJsonApi(w, data)
}

type StreamResponse struct {
	ID  string `jsonapi:"primary,video-stream"`
	Url string `jsonapi:"attr,url"`
}

func (h *Handler) HandleStreamVideo(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeJsonApiError(w, http.StatusMethodNotAllowed, "Method Not Allowed", "Only GET is allowed")
		return
	}

	// Extract video ID from path: /api/stream/videos/{id}
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		writeJsonApiError(w, http.StatusBadRequest, "Invalid Request", "Invalid video ID")
		return
	}
	videoID := pathParts[4] // /api/stream/videos/{id}

	url, err := h.usecase.GetStreamURL(r.Context(), videoID)
	if err != nil {
		log.Printf("Error getting stream URL: %v", err)
		writeJsonApiError(w, http.StatusNotFound, "Not Found", "Video not found")
		return
	}

	data := &StreamResponse{
		ID:  videoID,
		Url: url,
	}
	writeJsonApi(w, data)
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

package http

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/athandoan/youtube/streaming-service/internal/domain"
	"github.com/google/jsonapi"
)

type Handler struct {
	usecase domain.StreamingUsecase
}

func NewHandler(u domain.StreamingUsecase) *Handler {
	return &Handler{usecase: u}
}

type StreamResponse struct {
	ID  string `jsonapi:"primary,video-stream"`
	Url string `jsonapi:"attr,url"`
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

func (h *Handler) HandleStreamVideo(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeJsonApiError(w, http.StatusMethodNotAllowed, "Method Not Allowed", "Only GET is allowed")
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		writeJsonApiError(w, http.StatusBadRequest, "Invalid Request", "Invalid video ID")
		return
	}
	videoID := pathParts[2]

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
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

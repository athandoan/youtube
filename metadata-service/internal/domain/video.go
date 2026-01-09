package domain

//go:generate mockgen -source=video.go -destination=../mocks/mock_repository.go -package=mocks

import (
	"context"
	"time"
)

type Video struct {
	ID          string
	Title       string
	Description string
	BucketName  string
	ObjectKey   string
	Status      string
	CreatedAt   time.Time
}

type VideoRepository interface {
	Create(ctx context.Context, video *Video) error
	Get(ctx context.Context, id string) (*Video, error)
	List(ctx context.Context, query string) ([]*Video, error)
	UpdateStatus(ctx context.Context, id string, status string) error
}

type VideoUsecase interface {
	Create(ctx context.Context, title, bucket, objectKey string) (string, error)
	Get(ctx context.Context, id string) (*Video, error)
	List(ctx context.Context, query string) ([]*Video, error)
	UpdateStatus(ctx context.Context, id string, status string) error
}

package domain

//go:generate mockgen -source=video.go -destination=../mocks/mock_repository.go -package=mocks

import "time"

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
	Create(video *Video) error
	Get(id string) (*Video, error)
	List(query string) ([]*Video, error)
	UpdateStatus(id string, status string) error
}

type VideoUsecase interface {
	Create(title, bucket, objectKey string) (string, error)
	Get(id string) (*Video, error)
	List(query string) ([]*Video, error)
	UpdateStatus(id string, status string) error
}

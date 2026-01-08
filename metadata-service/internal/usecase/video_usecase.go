package usecase

import (
	"github.com/athandoan/youtube/metadata-service/internal/domain"
	"github.com/google/uuid"
)

type videoUsecase struct {
	repo domain.VideoRepository
}

func NewVideoUsecase(repo domain.VideoRepository) domain.VideoUsecase {
	return &videoUsecase{repo: repo}
}

func (u *videoUsecase) Create(title, bucket, objectKey string) (string, error) {
	id := uuid.New().String()
	video := &domain.Video{
		ID:         id,
		Title:      title,
		BucketName: bucket,
		ObjectKey:  objectKey,
		Status:     "pending",
	}
	if err := u.repo.Create(video); err != nil {
		return "", err
	}
	return id, nil
}

func (u *videoUsecase) Get(id string) (*domain.Video, error) {
	return u.repo.Get(id)
}

func (u *videoUsecase) List(query string) ([]*domain.Video, error) {
	return u.repo.List(query)
}

func (u *videoUsecase) UpdateStatus(id string, status string) error {
	return u.repo.UpdateStatus(id, status)
}

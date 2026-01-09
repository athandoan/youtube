package usecase

import (
	"context"

	"github.com/athandoan/youtube/metadata-service/internal/domain"
	"github.com/google/uuid"
)

type videoUsecase struct {
	repo domain.VideoRepository
}

func NewVideoUsecase(repo domain.VideoRepository) domain.VideoUsecase {
	return &videoUsecase{repo: repo}
}

func (u *videoUsecase) Create(ctx context.Context, title, bucket, objectKey string) (string, error) {
	id := uuid.New().String()
	video := &domain.Video{
		ID:         id,
		Title:      title,
		BucketName: bucket,
		ObjectKey:  objectKey,
		Status:     "pending",
	}
	if err := u.repo.Create(ctx, video); err != nil {
		return "", err
	}
	return id, nil
}

func (u *videoUsecase) Get(ctx context.Context, id string) (*domain.Video, error) {
	return u.repo.Get(ctx, id)
}

func (u *videoUsecase) List(ctx context.Context, query string) ([]*domain.Video, error) {
	return u.repo.List(ctx, query)
}

func (u *videoUsecase) UpdateStatus(ctx context.Context, id string, status string) error {
	return u.repo.UpdateStatus(ctx, id, status)
}

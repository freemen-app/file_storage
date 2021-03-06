package fileUseCase

import (
	"context"

	"github.com/freemen-app/file_storage/domain/dto"
)

type (
	useCase struct {
		fileRepo FileRepo
	}

	UseCase interface {
		Upload(ctx context.Context, input *dto.UploadInput) (string, error)
		Delete(ctx context.Context, input dto.DeleteInput) error
		BatchDelete(ctx context.Context, input dto.BatchDeleteInput) error
	}

	FileRepo interface {
		Upload(ctx context.Context, input *dto.UploadInput) (string, error)
		Delete(ctx context.Context, input dto.DeleteInput) error
		BatchDelete(ctx context.Context, input dto.BatchDeleteInput) error
	}
)

func New(fileRepo FileRepo) *useCase {
	return &useCase{
		fileRepo: fileRepo,
	}
}

func (u *useCase) Upload(ctx context.Context, input *dto.UploadInput) (string, error) {
	if err := input.Validate(); err != nil {
		return "", err
	}
	url, err := u.fileRepo.Upload(ctx, input)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (u *useCase) Delete(ctx context.Context, input dto.DeleteInput) error {
	if err := input.Validate(); err != nil {
		return err
	}
	err := u.fileRepo.Delete(ctx, input)
	return err
}

func (u *useCase) BatchDelete(ctx context.Context, input dto.BatchDeleteInput) error {
	if err := input.Validate(); err != nil {
		return err
	}
	err := u.fileRepo.BatchDelete(ctx, input)
	return err
}

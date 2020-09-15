package fileUseCase

import (
	"github.com/freemen-app/file_storage/domain/dto"
)

type (
	useCase struct {
		fileRepo FileRepo
	}

	FileRepo interface {
		Upload(input *dto.UploadInput) (string, error)
		Delete(input *dto.DeleteInput) error
		BatchDelete(input *dto.BatchDeleteInput) error
	}
)

func New(fileRepo FileRepo) *useCase {
	return &useCase{
		fileRepo: fileRepo,
	}
}

func (u *useCase) Upload(input *dto.UploadInput) (string, error) {
	if err := input.Validate(); err != nil {
		return "", err
	}
	url, err := u.fileRepo.Upload(input)
	return url, err
}

func (u *useCase) Delete(input *dto.DeleteInput) error {
	if err := input.Validate(); err != nil {
		return err
	}
	err := u.fileRepo.Delete(input)
	return err
}

func (u *useCase) BatchDelete(input *dto.BatchDeleteInput) error {
	if err := input.Validate(); err != nil {
		return err
	}
	err := u.fileRepo.BatchDelete(input)
	return err
}

package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/freemen-app/file_storage/domain/dto"
)

type FileUseCase struct {
	mock.Mock
}

func (u *FileUseCase) Upload(ctx context.Context, input *dto.UploadInput) (string, error) {
	args := u.Called(ctx, input)
	return args.String(0), args.Error(1)
}

func (u *FileUseCase) Delete(ctx context.Context, input dto.DeleteInput) error {
	args := u.Called(ctx, input)
	return args.Error(0)
}

func (u *FileUseCase) BatchDelete(ctx context.Context, input dto.BatchDeleteInput) error {
	args := u.Called(ctx, input)
	return args.Error(0)
}

package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/freemen-app/file_storage/domain/dto"
)

type FileRepo struct {
	mock.Mock
}

func (f *FileRepo) Upload(ctx context.Context, input *dto.UploadInput) (string, error) {
	args := f.Called(ctx, input)
	return args.String(0), args.Error(1)
}

func (f *FileRepo) Delete(ctx context.Context, input dto.DeleteInput) error {
	args := f.Called(ctx, input)
	return args.Error(0)
}

func (f *FileRepo) BatchDelete(ctx context.Context, input dto.BatchDeleteInput) error {
	args := f.Called(ctx, input)
	return args.Error(0)
}

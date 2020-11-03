package fileUseCase_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"

	"github.com/freemen-app/file_storage/domain/dto"
	"github.com/freemen-app/file_storage/infrastructure/testing/helpers"
	"github.com/freemen-app/file_storage/infrastructure/testing/mocks"
	fileUseCase "github.com/freemen-app/file_storage/usecase/file"
)

func TestNew(t *testing.T) {
	fileRepo := new(mocks.FileRepo)
	useCase := fileUseCase.New(fileRepo)
	assert.EqualValues(t, fileRepo, useCase.FileRepo())
}

func TestUseCase_Upload(t *testing.T) {
	type args struct {
		ctx   context.Context
		input *dto.UploadInput
	}
	tests := []struct {
		name      string
		args      args
		mockCalls mocks.Calls
		want      string
		wantErr   error
	}{
		{
			name: "succeed",
			args: args{
				ctx:   helpers.DefaultCtx,
				input: &dto.UploadInput{File: bytes.NewBufferString("test"), Directory: "test_dir", Filename: "test.jpg", ACL: "public-read"},
			},
			mockCalls: mocks.Calls{
				{
					Method: "Upload",
					Args: []interface{}{
						helpers.DefaultCtx,
						&dto.UploadInput{File: bytes.NewBufferString("test"), Directory: "test_dir", Filename: "test.jpg", ACL: "public-read"},
					},
					ReturnArgs: []interface{}{"https://aws.s3/test.bucket/test.jpg", nil},
				},
			},
			want: "https://aws.s3/test.bucket/test.jpg",
		},
		{
			name: "invalid input",
			args: args{
				ctx: helpers.DefaultCtx,
				input: &dto.UploadInput{
					File:     bytes.NewBufferString("test"),
					Filename: "test.jpg",
					ACL:      "invalid",
				},
			},
			wantErr: validation.Errors{},
		},
		{
			name: "error from file repo",
			args: args{
				ctx:   helpers.DefaultCtx,
				input: &dto.UploadInput{File: bytes.NewBufferString("test"), Directory: "test_dir", Filename: "test.jpg", ACL: "public-read"},
			},
			mockCalls: mocks.Calls{
				{
					Method: "Upload",
					Args: []interface{}{
						helpers.DefaultCtx,
						&dto.UploadInput{File: bytes.NewBufferString("test"), Directory: "test_dir", Filename: "test.jpg", ACL: "public-read"},
					},
					ReturnArgs: []interface{}{"", errors.New("test error")},
				},
			},
			wantErr: errors.New("test error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileRepo := new(mocks.FileRepo)
			for _, call := range tt.mockCalls {
				fileRepo.On(call.Method, call.Args...).Return(call.ReturnArgs...)
			}
			useCase := fileUseCase.New(fileRepo)
			got, gotErr := useCase.Upload(tt.args.ctx, tt.args.input)
			if _, ok := tt.wantErr.(validation.Errors); ok {
				assert.IsType(t, tt.wantErr, gotErr)
			} else {
				assert.EqualValues(t, tt.wantErr, gotErr)
			}
			assert.EqualValues(t, tt.want, got)
			fileRepo.AssertExpectations(t)
		})
	}
}

func TestUseCase_Delete(t *testing.T) {
	type args struct {
		ctx   context.Context
		input dto.DeleteInput
	}
	tests := []struct {
		name      string
		args      args
		mockCalls mocks.Calls
		wantErr   error
	}{
		{
			name: "succeed",
			args: args{
				ctx:   helpers.DefaultCtx,
				input: "https://aws.s3/test.bucket/test.jpg",
			},
			mockCalls: mocks.Calls{
				{
					Method:     "Delete",
					Args:       []interface{}{helpers.DefaultCtx, dto.DeleteInput("https://aws.s3/test.bucket/test.jpg")},
					ReturnArgs: []interface{}{nil},
				},
			},
		},
		{
			name: "invalid input",
			args: args{
				ctx:   helpers.DefaultCtx,
				input: "not url",
			},
			wantErr: validation.ErrorObject{},
		},
		{
			name: "error from file repo",
			args: args{
				ctx:   helpers.DefaultCtx,
				input: "https://aws.s3/test.bucket/test.jpg",
			},
			mockCalls: mocks.Calls{
				{
					Method:     "Delete",
					Args:       []interface{}{helpers.DefaultCtx, dto.DeleteInput("https://aws.s3/test.bucket/test.jpg")},
					ReturnArgs: []interface{}{errors.New("test error")},
				},
			},
			wantErr: errors.New("test error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileRepo := new(mocks.FileRepo)
			for _, call := range tt.mockCalls {
				fileRepo.On(call.Method, call.Args...).Return(call.ReturnArgs...)
			}
			useCase := fileUseCase.New(fileRepo)
			gotErr := useCase.Delete(tt.args.ctx, tt.args.input)
			if _, ok := tt.wantErr.(validation.ErrorObject); ok {
				assert.IsType(t, tt.wantErr, gotErr)
			} else {
				assert.EqualValues(t, tt.wantErr, gotErr)
			}
			fileRepo.AssertExpectations(t)
		})
	}
}

func TestUseCase_BatchDelete(t *testing.T) {
	type args struct {
		ctx   context.Context
		input dto.BatchDeleteInput
	}
	tests := []struct {
		name      string
		args      args
		mockCalls mocks.Calls
		wantErr   error
	}{
		{
			name: "succeed",
			args: args{
				ctx:   helpers.DefaultCtx,
				input: dto.BatchDeleteInput{"https://aws.s3/test.bucket/test.jpg", "https://aws.s3/test.bucket/test2.jpg"},
			},
			mockCalls: mocks.Calls{
				{
					Method: "BatchDelete",
					Args: []interface{}{
						helpers.DefaultCtx,
						dto.BatchDeleteInput{"https://aws.s3/test.bucket/test.jpg", "https://aws.s3/test.bucket/test2.jpg"},
					},
					ReturnArgs: []interface{}{nil},
				},
			},
		},
		{
			name: "invalid input",
			args: args{
				ctx:   helpers.DefaultCtx,
				input: dto.BatchDeleteInput{"https://aws.s3/test.bucket/test.jpg", "invalid url"},
			},
			wantErr: validation.Errors{},
		},
		{
			name: "error from file repo",
			args: args{
				ctx:   helpers.DefaultCtx,
				input: dto.BatchDeleteInput{"https://aws.s3/test.bucket/test.jpg"},
			},
			mockCalls: mocks.Calls{
				{
					Method:     "BatchDelete",
					Args:       []interface{}{helpers.DefaultCtx, dto.BatchDeleteInput{"https://aws.s3/test.bucket/test.jpg"}},
					ReturnArgs: []interface{}{errors.New("test error")},
				},
			},
			wantErr: errors.New("test error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileRepo := new(mocks.FileRepo)
			for _, call := range tt.mockCalls {
				fileRepo.On(call.Method, call.Args...).Return(call.ReturnArgs...)
			}
			useCase := fileUseCase.New(fileRepo)
			gotErr := useCase.BatchDelete(tt.args.ctx, tt.args.input)
			if _, ok := tt.wantErr.(validation.Errors); ok {
				assert.IsType(t, tt.wantErr, gotErr)
			} else {
				assert.EqualValues(t, tt.wantErr, gotErr)
			}
			fileRepo.AssertExpectations(t)
		})
	}
}

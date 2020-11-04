package fileRepo_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	fileRepo "github.com/freemen-app/file_storage/adapter/repository/file"
	"github.com/freemen-app/file_storage/config"
	"github.com/freemen-app/file_storage/domain/dto"
	customErrors "github.com/freemen-app/file_storage/domain/errors"
	awsSession "github.com/freemen-app/file_storage/infrastructure/store/aws"
	"github.com/freemen-app/file_storage/infrastructure/testing/helpers"
	"github.com/freemen-app/file_storage/infrastructure/testing/mocks"
)

type fields struct {
	Uploader     s3manageriface.UploaderAPI
	Deleter      fileRepo.Deleter
	BatchDeleter s3manageriface.BatchDelete
	bucketName   string
}

func testRepo(f *fields) *fileRepo.Repo {
	repo := &fileRepo.Repo{}
	repo.SetBucketName(f.bucketName)
	repo.SetUploader(f.Uploader)
	repo.SetDeleter(f.Deleter)
	repo.SetBatchDeleter(f.BatchDeleter)
	return repo
}

func TestNew(t *testing.T) {
	t.Run("Nil session panic", func(t *testing.T) {
		assert.Panics(t, func() {
			fileRepo.New(nil, "test.bucket")
		})
	})
	t.Run("Succeed", func(t *testing.T) {
		session := awsSession.New(config.S3Config{})
		bucket := "test.bucket"

		repo := fileRepo.New(session, bucket)
		assert.EqualValues(t, bucket, repo.BucketName())
		assert.NotNil(t, repo.Uploader())
		assert.NotNil(t, repo.Deleter())
		assert.NotNil(t, repo.BatchDeleter())
	})
}

func setupMocks(t *testing.T, f *fields, callsMap map[string]mocks.Calls) func() {
	mockObjects := reflect.ValueOf(f).Elem()
	for k, v := range callsMap {
		if field := mockObjects.FieldByName(k); !field.IsValid() {
			continue
		} else if mockObject, ok := field.Interface().(mocks.Mock); ok {
			for _, call := range v {
				mockObject.On(call.Method, call.Args...).Return(call.ReturnArgs...)
			}
		}
	}
	return func() {
		for k := range callsMap {
			if field := mockObjects.FieldByName(k); !field.IsValid() {
				continue
			} else if mockObject, ok := field.Interface().(mocks.Mock); ok {
				mockObject.AssertExpectations(t)
			}
		}
	}
}

func TestRepo_Upload(t *testing.T) {
	type args struct {
		ctx   context.Context
		input *dto.UploadInput
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		mocks   map[string]mocks.Calls
		want    string
		wantErr error
	}{
		{
			name: "succeed",
			fields: fields{
				Uploader:   new(mocks.Uploader),
				bucketName: "test.bucket",
			},
			args: args{
				ctx: helpers.DefaultCtx,
				input: &dto.UploadInput{
					File:      nil,
					Directory: "test",
					Filename:  "test.jpg",
					ACL:       "public-read",
				},
			},
			mocks: map[string]mocks.Calls{
				"Uploader": {
					{
						Method:     "UploadWithContext",
						Args:       []interface{}{helpers.DefaultCtx, mock.Anything},
						ReturnArgs: []interface{}{&s3manager.UploadOutput{Location: "https://aws.s3/test/test.jpg"}, nil},
					},
				},
			},
			want: "https://aws.s3/test/test.jpg",
		},
		{
			name: "error returned",
			fields: fields{
				Uploader:   new(mocks.Uploader),
				bucketName: "test.bucket",
			},
			args: args{
				ctx: helpers.DefaultCtx,
				input: &dto.UploadInput{
					File:      nil,
					Directory: "test",
					Filename:  "test.jpg",
					ACL:       "public-read",
				},
			},
			mocks: map[string]mocks.Calls{
				"Uploader": {
					{
						Method:     "UploadWithContext",
						Args:       []interface{}{helpers.DefaultCtx, mock.Anything},
						ReturnArgs: []interface{}{nil, context.DeadlineExceeded},
					},
				},
			},
			wantErr: context.DeadlineExceeded,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertMocks := setupMocks(t, &tt.fields, tt.mocks)
			defer assertMocks()
			repo := testRepo(&tt.fields)
			got, err := repo.Upload(tt.args.ctx, tt.args.input)
			assert.EqualValues(t, tt.wantErr, err)
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func TestRepo_Delete(t *testing.T) {
	type args struct {
		ctx   context.Context
		input dto.DeleteInput
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		mocks   map[string]mocks.Calls
		want    string
		wantErr error
	}{
		{
			name: "succeed",
			fields: fields{
				Deleter:    new(mocks.Deleter),
				bucketName: "test.bucket",
			},
			args: args{
				ctx:   helpers.DefaultCtx,
				input: "https://aws.s3/test.bucket/test.jpg",
			},
			mocks: map[string]mocks.Calls{
				"Deleter": {
					{
						Method:     "DeleteObjectWithContext",
						Args:       []interface{}{helpers.DefaultCtx, mock.Anything},
						ReturnArgs: []interface{}{new(s3.DeleteObjectOutput), nil},
					},
				},
			},
			want: "https://aws.s3/test/test.jpg",
		},
		{
			name: "invalid input",
			fields: fields{
				Deleter:    new(mocks.Deleter),
				bucketName: "test.bucket",
			},
			args: args{
				ctx:   helpers.DefaultCtx,
				input: "https://aws.s3/invalid.bucket/test.jpg",
			},
			wantErr: customErrors.InvalidURL,
		},
		{
			name: "error returned",
			fields: fields{
				Deleter:    new(mocks.Deleter),
				bucketName: "test.bucket",
			},
			args: args{
				ctx:   helpers.DefaultCtx,
				input: "https://aws.s3/test.bucket/test.jpg",
			},
			mocks: map[string]mocks.Calls{
				"Deleter": {
					{
						Method:     "DeleteObjectWithContext",
						Args:       []interface{}{helpers.DefaultCtx, mock.Anything},
						ReturnArgs: []interface{}{nil, errors.New("test error")},
					},
				},
			},
			wantErr: errors.New("test error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertMocks := setupMocks(t, &tt.fields, tt.mocks)
			defer assertMocks()
			repo := testRepo(&tt.fields)
			err := repo.Delete(tt.args.ctx, tt.args.input)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestRepo_BatchDelete(t *testing.T) {
	type args struct {
		ctx   context.Context
		input dto.BatchDeleteInput
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		mocks   map[string]mocks.Calls
		wantErr error
	}{
		{
			name: "succeed",
			fields: fields{
				BatchDeleter: new(mocks.BatchDeleter),
				bucketName:   "test.bucket",
			},
			args: args{
				ctx:   helpers.DefaultCtx,
				input: dto.BatchDeleteInput{"https://aws.s3/test.bucket/test.jpg", "https://aws.s3/test.bucket/test2.jpg"},
			},
			mocks: map[string]mocks.Calls{
				"BatchDeleter": {
					{
						Method:     "Delete",
						Args:       []interface{}{helpers.DefaultCtx, mock.Anything},
						ReturnArgs: []interface{}{nil},
					},
				},
			},
		},
		{
			name: "invalid input",
			fields: fields{
				BatchDeleter: new(mocks.BatchDeleter),
				bucketName:   "test.bucket",
			},
			args: args{
				ctx:   helpers.DefaultCtx,
				input: dto.BatchDeleteInput{"https://aws.s3/test.bucket/test.jpg", "https://aws.s3/invalid.bucket/test2.jpg"},
			},
			wantErr: customErrors.InvalidURL,
		},
		{
			name: "error returned",
			fields: fields{
				BatchDeleter: new(mocks.BatchDeleter),
				bucketName:   "test.bucket",
			},
			args: args{
				ctx:   helpers.DefaultCtx,
				input: dto.BatchDeleteInput{"https://aws.s3/test.bucket/test.jpg", "https://aws.s3/test.bucket/test2.jpg"},
			},
			mocks: map[string]mocks.Calls{
				"BatchDeleter": {
					{
						Method:     "Delete",
						Args:       []interface{}{helpers.DefaultCtx, mock.Anything},
						ReturnArgs: []interface{}{errors.New("test error")},
					},
				},
			},
			wantErr: errors.New("test error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertMocks := setupMocks(t, &tt.fields, tt.mocks)
			defer assertMocks()
			repo := testRepo(&tt.fields)
			err := repo.BatchDelete(tt.args.ctx, tt.args.input)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

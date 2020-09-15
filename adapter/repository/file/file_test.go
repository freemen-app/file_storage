package fileRepo

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"

	"github.com/freemen-app/file_storage/domain/dto"
	awsSession "github.com/freemen-app/file_storage/infrastructure/store/aws"
)

type (
	serviceWithErr struct {
		err error
	}

	batchDeleterWithErr struct {
		err error
	}

	mockedUploader struct {
		baseUrl string
	}

	uploaderWithErr struct {
		err error
	}
)

func (s *serviceWithErr) DeleteObject(*s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	return nil, s.err
}

func (d *batchDeleterWithErr) Delete(aws.Context, s3manager.BatchDeleteIterator) error {
	return d.err
}

func (m *mockedUploader) Upload(
	input *s3manager.UploadInput,
	options ...func(*s3manager.Uploader),
) (*s3manager.UploadOutput, error) {

	return &s3manager.UploadOutput{
		Location: fmt.Sprintf("%s/%s/%s", m.baseUrl, *input.Bucket, *input.Key),
	}, nil

}

func (u *uploaderWithErr) Upload(
	input *s3manager.UploadInput,
	options ...func(*s3manager.Uploader),
) (*s3manager.UploadOutput, error) {
	return nil, u.err
}

func TestNew(t *testing.T) {
	t.Run("Nil session panic", func(t *testing.T) {
		assert.Panics(t, func() {
			New(nil, "test.bucket")
		})
	})
	t.Run("Succeed", func(t *testing.T) {
		session := awsSession.New()
		bucket := "test.bucket"
		want := &repo{
			service:      s3.New(session),
			uploader:     s3manager.NewUploader(session),
			batchDeleter: s3manager.NewBatchDelete(session),
			bucketName:   bucket,
		}

		got := New(session, bucket)
		assert.IsType(t, want.service, got.service)
		assert.IsType(t, want.uploader, got.uploader)
		assert.IsType(t, want.batchDeleter, got.batchDeleter)
		assert.EqualValues(t, want.bucketName, got.bucketName)
	})
}

func TestRepo_Upload(t *testing.T) {
	baseUrl := "https://aws.com"

	type fields struct {
		service      Service
		uploader     Uploader
		batchDeleter BatchDeleter
		bucketName   string
	}
	tests := []struct {
		name    string
		fields  fields
		input   *dto.UploadInput
		want    string
		wantErr error
	}{
		{
			name: "Succeed",
			fields: fields{
				uploader:   &mockedUploader{baseUrl: baseUrl},
				bucketName: "test.bucket",
			},
			input: &dto.UploadInput{
				File:      nil,
				Directory: "test",
				Filename:  "test.jpg",
				ACL:       "public-read",
			},
			want:    fmt.Sprintf("%s/test.bucket/test/test.jpg", baseUrl),
			wantErr: nil,
		},
		{
			name: "Error",
			fields: fields{
				uploader:   &uploaderWithErr{err: errors.New("test err")},
				bucketName: "test.bucket",
			},
			input: &dto.UploadInput{
				File:      nil,
				Directory: "test",
				Filename:  "test.jpg",
				ACL:       "public-read",
			},
			want:    "",
			wantErr: errors.New("test err"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &repo{
				service:    tt.fields.service,
				uploader:   tt.fields.uploader,
				bucketName: tt.fields.bucketName,
			}
			got, err := r.Upload(tt.input)
			assert.EqualValues(t, tt.wantErr, err)
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func TestRepo_Delete(t *testing.T) {
	type fields struct {
		service      Service
		uploader     Uploader
		batchDeleter BatchDeleter
		bucketName   string
	}
	tests := []struct {
		name    string
		fields  fields
		input   dto.DeleteInput
		wantErr error
	}{
		{
			name:  "Succeed",
			input: "https://s3.com/test.bucket/test/test.jpg",
			fields: fields{
				service:    &serviceWithErr{err: nil},
				bucketName: "test.bucket",
			},
			wantErr: nil,
		},
		{
			name:  "Error while converting input",
			input: "https://s3.com/wrong.bucket/test.jpg",
			fields: fields{
				service:    &serviceWithErr{err: nil},
				bucketName: "test.bucket",
			},
			wantErr: validation.NewError("400", "url: invalid format"),
		},
		{
			name:  "Error returned from s3",
			input: "https://s3.com/test.bucket/test/test.jpg",
			fields: fields{
				service:    &serviceWithErr{err: errors.New("test error")},
				bucketName: "test.bucket",
			},
			wantErr: errors.New("test error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &repo{
				service:    tt.fields.service,
				uploader:   tt.fields.uploader,
				bucketName: tt.fields.bucketName,
			}
			err := r.Delete(&tt.input)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestRepo_BatchDelete(t *testing.T) {
	type fields struct {
		service      Service
		uploader     Uploader
		batchDeleter BatchDeleter
		bucketName   string
	}
	tests := []struct {
		name    string
		fields  fields
		input   dto.BatchDeleteInput
		wantErr error
	}{
		{
			name: "Succeed",
			input: dto.BatchDeleteInput([]dto.DeleteInput{
				"https://s3.com/test.bucket/test/test.jpg",
				"https://s3.com/test.bucket/test/test.jpg",
			}),
			fields: fields{
				batchDeleter: &batchDeleterWithErr{err: nil},
				bucketName:   "test.bucket",
			},
			wantErr: nil,
		},
		{
			name: "Error while converting input",
			input: dto.BatchDeleteInput([]dto.DeleteInput{
				"https://s3.com/test.bucket/test/test.jpg",
				"https://s3.com/wrong.bucket/test/test.jpg",
			}),
			fields: fields{
				batchDeleter: &batchDeleterWithErr{err: nil},
				bucketName:   "test.bucket",
			},
			wantErr: validation.NewError("400", "url: invalid format"),
		},
		{
			name: "Error returned from s3",
			input: dto.BatchDeleteInput([]dto.DeleteInput{
				"https://s3.com/test.bucket/test/test.jpg",
				"https://s3.com/test.bucket/test/test2.jpg",
			}),
			fields: fields{
				batchDeleter: &batchDeleterWithErr{err: errors.New("test error")},
				bucketName:   "test.bucket",
			},
			wantErr: errors.New("test error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &repo{
				batchDeleter: tt.fields.batchDeleter,
				bucketName:   tt.fields.bucketName,
			}
			err := r.BatchDelete(&tt.input)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

package fileRepo

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/freemen-app/file_storage/domain/dto"
)

type (
	Uploader interface {
		Upload(
			input *s3manager.UploadInput,
			options ...func(*s3manager.Uploader),
		) (*s3manager.UploadOutput, error)
	}

	BatchDeleter interface {
		Delete(ctx aws.Context, iter s3manager.BatchDeleteIterator) error
	}

	Service interface {
		DeleteObject(input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)
	}

	repo struct {
		service      Service
		uploader     Uploader
		batchDeleter BatchDeleter
		bucketName   string
	}
)

func New(session *session.Session, bucketName string) *repo {
	service := s3.New(session)
	uploader := s3manager.NewUploaderWithClient(service)
	batchDeleter := s3manager.NewBatchDeleteWithClient(service)
	return &repo{
		service:      service,
		uploader:     uploader,
		batchDeleter: batchDeleter,
		bucketName:   bucketName,
	}
}

func (r *repo) Upload(input *dto.UploadInput) (string, error) {
	s3Input := input.ToS3Input(r.bucketName)
	resp, err := r.uploader.Upload(s3Input)
	if err != nil {
		return "", err
	}
	return resp.Location, nil
}

func (r *repo) Delete(input *dto.DeleteInput) error {
	s3Input, err := input.ToS3Input(r.bucketName)
	if err != nil {
		return err
	}
	_, err = r.service.DeleteObject(s3Input)
	return err
}

func (r *repo) BatchDelete(input *dto.BatchDeleteInput) error {
	s3Input, err := input.ToS3Input(r.bucketName)
	if err != nil {
		return err
	}
	err = r.batchDeleter.Delete(aws.BackgroundContext(), s3Input)
	return err
}

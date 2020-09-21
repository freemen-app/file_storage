package fileRepo

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/rs/zerolog/log"

	"github.com/freemen-app/file_storage/domain/dto"
)

type (
	Service interface {
		DeleteObject(input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)
	}

	repo struct {
		service       Service
		uploader      s3manageriface.UploaderAPI
		batchUploader s3manageriface.UploadWithIterator
		batchDeleter  s3manageriface.BatchDelete
		bucketName    string
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

func (r *repo) Upload(ctx context.Context, input *dto.UploadInput) (string, error) {
	log.Info().Msgf("%s", time.Now())
	s3Input := input.ToS3Input(r.bucketName)
	log.Info().Msgf("%s", time.Now())

	resp, err := r.uploader.UploadWithContext(ctx, s3Input)

	log.Info().Msgf("%s", time.Now())
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

func (r *repo) BatchDelete(ctx context.Context, input *dto.BatchDeleteInput) error {
	s3Input, err := input.ToS3Input(r.bucketName)
	if err != nil {
		return err
	}
	err = r.batchDeleter.Delete(ctx, s3Input)
	return err
}

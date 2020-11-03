package fileRepo

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/rs/zerolog/log"

	"github.com/freemen-app/file_storage/domain/dto"
)

type (
	Deleter interface {
		DeleteObjectWithContext(ctx aws.Context, input *s3.DeleteObjectInput, opts ...request.Option) (*s3.DeleteObjectOutput, error)
	}

	repo struct {
		deleter      Deleter
		uploader     s3manageriface.UploaderAPI
		batchDeleter s3manageriface.BatchDelete
		bucketName   string
	}
)

func New(session *session.Session, bucketName string) *repo {
	service := s3.New(session)
	uploader := s3manager.NewUploaderWithClient(service)
	batchDeleter := s3manager.NewBatchDeleteWithClient(service)
	return &repo{
		deleter:      service,
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

func (r *repo) Delete(ctx context.Context, input dto.DeleteInput) error {
	if s3Input, err := input.ToS3Input(r.bucketName); err != nil {
		return err
	} else if _, err := r.deleter.DeleteObjectWithContext(ctx, s3Input); err != nil {
		return err
	}
	return nil
}

func (r *repo) BatchDelete(ctx context.Context, input dto.BatchDeleteInput) error {
	if s3Input, err := input.ToS3Input(r.bucketName); err != nil {
		return err
	} else if err := r.batchDeleter.Delete(ctx, s3Input); err != nil {
		return err
	}
	return nil
}

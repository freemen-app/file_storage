package mocks

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/stretchr/testify/mock"
)

type (
	Uploader struct {
		mock.Mock
	}

	Deleter struct {
		mock.Mock
	}

	BatchDeleter struct {
		mock.Mock
	}
)

func (u *Uploader) Upload(input *s3manager.UploadInput, f ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	args := u.Called(input)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3manager.UploadOutput), nil
}

func (u *Uploader) UploadWithContext(ctx aws.Context, input *s3manager.UploadInput, f ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	args := u.Called(ctx, input)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3manager.UploadOutput), nil
}

func (d *Deleter) DeleteObjectWithContext(ctx aws.Context, input *s3.DeleteObjectInput, opts ...request.Option) (*s3.DeleteObjectOutput, error) {
	args := d.Called(ctx, input)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.DeleteObjectOutput), nil
}

func (b *BatchDeleter) Delete(ctx aws.Context, input s3manager.BatchDeleteIterator) error {
	args := b.Called(ctx, input)
	return args.Error(0)
}

package dto

import (
	"io"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type (
	UploadInput struct {
		File      io.Reader
		Directory string
		Filename  string
		ACL       string
	}
)

func (i *UploadInput) Validate() error {
	return validation.ValidateStruct(
		i,
		validation.Field(&i.Filename, validation.Required),
		validation.Field(&i.File, validation.Required),
		validation.Field(&i.ACL, validation.In(
			"public-read",
			"public-read-write",
			"aws-exec-read",
			"authenticated-read",
			"bucket-owner-read",
			"bucket-owner-full-control",
			"log-delivery-write",
		)),
	)
}

func (i *UploadInput) ToS3Input(bucketName string) *s3manager.UploadInput {
	return &s3manager.UploadInput{
		Body:   i.File,
		Key:    aws.String(path.Join(i.Directory, i.Filename)),
		Bucket: aws.String(bucketName),
		ACL:    aws.String(i.ACL),
	}
}

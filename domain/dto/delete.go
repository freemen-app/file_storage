package dto

import (
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"

	customErrors "github.com/freemen-app/file_storage/domain/errors"
)

type (
	DeleteInput string

	BatchDeleteInput []DeleteInput
)

func (i DeleteInput) Validate() error {
	return validation.Validate(i.String(), validation.Required, is.URL)
}

func (i DeleteInput) String() string {
	return string(i)
}

func (i BatchDeleteInput) Validate() error {
	return validation.Validate([]DeleteInput(i))
}

func (i DeleteInput) ToS3Input(bucketName string) (*s3.DeleteObjectInput, error) {
	regexpString := fmt.Sprintf("https://.*/%s/(.*)$", bucketName)
	re := regexp.MustCompile(regexpString)
	groups := re.FindStringSubmatch(i.String())
	if len(groups) != 2 {
		return nil, customErrors.InvalidURL
	}
	return &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(groups[1]),
	}, nil
}

func (i BatchDeleteInput) ToS3Input(bucketName string) (*s3manager.DeleteObjectsIterator, error) {
	files := &s3manager.DeleteObjectsIterator{
		Objects: []s3manager.BatchDeleteObject{},
	}
	for _, obj := range i {
		if s3Input, err := obj.ToS3Input(bucketName); err != nil {
			return nil, err
		} else {
			files.Objects = append(files.Objects, s3manager.BatchDeleteObject{Object: s3Input})
		}
	}
	return files, nil
}

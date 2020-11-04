package awsSession

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/freemen-app/file_storage/config"
)

func New(config config.S3Config) *session.Session {
	return session.Must(session.NewSession(&aws.Config{
		Region: aws.String(config.Region),
	}))
}

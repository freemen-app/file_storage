package awsSession

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func New() *session.Session {
	return session.Must(session.NewSession(&aws.Config{
		Region: aws.String("eu-central-1"),
	}))
}

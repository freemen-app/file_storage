package dto

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/stretchr/testify/assert"
)

func TestDeleteInput_Validate(t *testing.T) {
	type fields struct {
		url string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Valid url",
			fields: fields{
				url: "https://aws.amazonaws.com/bucket/test/test.yml",
			},
			wantErr: false,
		},
		{
			name: "Valid url",
			fields: fields{
				url: "aws.amazonaws.com/bucket/test/test.yml",
			},
			wantErr: false,
		},
		{
			name: "Invalid url",
			fields: fields{
				url: "test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := DeleteInput(tt.fields.url)
			err := i.Validate()
			assert.EqualValues(t, tt.wantErr, err != nil, err)
		})
	}
}

func TestBatchDeleteInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		urls    BatchDeleteInput
		wantErr bool
	}{
		{
			name: "valid url",
			urls: BatchDeleteInput{
				"https://aws.amazonaws.com/bucket/test/test.yml",
				"https://aws.amazonaws.com/bucket/test/test2.yml",
			},
		},
		{
			name: "invalid url",
			urls: BatchDeleteInput{
				"https://aws.amazonaws.com/bucket/test/test.yml",
				"test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.urls.Validate()
			assert.EqualValues(t, tt.wantErr, err != nil, err)
		})
	}
}

func TestDeleteInput_ToS3Input(t *testing.T) {
	type fields struct {
		Url string
	}
	tests := []struct {
		name       string
		bucketName string
		fields     fields
		want       *s3.DeleteObjectInput
		wantErr    bool
	}{
		{
			name:       "Valid",
			bucketName: "test.bucket",
			fields: fields{
				Url: "https://aws.amazonaws.com/test.bucket/test/test.jpg",
			},
			want: &s3.DeleteObjectInput{
				Bucket: aws.String("test.bucket"),
				Key:    aws.String("test/test.jpg"),
			},
		},
		{
			name:       "Wrong bucket name",
			bucketName: "test.bucket",
			fields: fields{
				Url: "https://aws.amazonaws.com/wrong.bucket/test/test.jpg",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := DeleteInput(tt.fields.Url)
			got, err := i.ToS3Input(tt.bucketName)
			assert.EqualValues(t, tt.wantErr, err != nil, err)
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func TestBatchDeleteInput_ToS3Input(t *testing.T) {
	type fields struct {
		Urls []DeleteInput
	}
	tests := []struct {
		name       string
		bucketName string
		fields     fields
		want       *s3manager.DeleteObjectsIterator
		wantErr    bool
	}{
		{
			name:       "Valid",
			bucketName: "test.bucket",
			fields: fields{
				Urls: []DeleteInput{
					DeleteInput("https://aws.amazonaws.com/test.bucket/test/test.yml"),
					DeleteInput("https://aws.amazonaws.com/test.bucket/test/test2.yml"),
				},
			},
			want: &s3manager.DeleteObjectsIterator{
				Objects: []s3manager.BatchDeleteObject{
					{Object: &s3.DeleteObjectInput{
						Bucket: aws.String("test.bucket"),
						Key:    aws.String("test/test.yml"),
					}},
					{Object: &s3.DeleteObjectInput{
						Bucket: aws.String("test.bucket"),
						Key:    aws.String("test/test2.yml"),
					}},
				},
			},
		},
		{
			name:       "Wrong bucket name",
			bucketName: "test.bucket",
			fields: fields{
				Urls: []DeleteInput{
					DeleteInput("https://aws.amazonaws.com/wrong.bucket/test/test.yml"),
					DeleteInput("https://aws.amazonaws.com/test.bucket/test/test2.yml"),
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := BatchDeleteInput(tt.fields.Urls)
			got, err := i.ToS3Input(tt.bucketName)
			assert.EqualValues(t, tt.wantErr, err != nil, err)
			assert.EqualValues(t, tt.want, got)
		})
	}
}

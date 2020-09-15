package dto

import (
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/stretchr/testify/assert"
)

func TestUploadInput_ToS3Input(t *testing.T) {
	type fields struct {
		File      io.Reader
		Directory string
		Filename  string
		ACL       string
	}
	tests := []struct {
		name       string
		bucketName string
		fields     fields
		want       *s3manager.UploadInput
	}{
		{
			name:       "Valid",
			bucketName: "test.bucket",
			fields: fields{
				File:      nil,
				Directory: "test",
				Filename:  "test.jpg",
				ACL:       "public-read",
			},
			want: &s3manager.UploadInput{
				Body:   nil,
				Bucket: aws.String("test.bucket"),
				Key:    aws.String("test/test.jpg"),
				ACL:    aws.String("public-read"),
			},
		},
		{
			name:       "Empty directory",
			bucketName: "test.bucket",
			fields: fields{
				File:      nil,
				Directory: "",
				Filename:  "test.jpg",
				ACL:       "public-read",
			},
			want: &s3manager.UploadInput{
				Body:   nil,
				Bucket: aws.String("test.bucket"),
				Key:    aws.String("test.jpg"),
				ACL:    aws.String("public-read"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &UploadInput{
				File:      tt.fields.File,
				Directory: tt.fields.Directory,
				Filename:  tt.fields.Filename,
				ACL:       tt.fields.ACL,
			}
			got := i.ToS3Input(tt.bucketName)
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func TestUploadInput_Validate(t *testing.T) {
	type fields struct {
		File      io.Reader
		Directory string
		Filename  string
		ACL       string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Valid struct",
			fields: fields{
				File:      strings.NewReader("test"),
				Directory: "test",
				Filename:  "test.jpg",
				ACL:       "public-read",
			},
			wantErr: false,
		},
		{
			name: "Valid struct",
			fields: fields{
				File:     strings.NewReader("test"),
				Filename: "test.png",
				ACL:      "public-read-write",
			},
			wantErr: false,
		},
		{
			name: "Empty file",
			fields: fields{
				File:      nil,
				Directory: "test",
				Filename:  "test.jpg",
				ACL:       "test",
			},
			wantErr: true,
		},
		{
			name: "No filename",
			fields: fields{
				File:      strings.NewReader("test"),
				Directory: "test",
				ACL:       "public-read",
			},
			wantErr: true,
		},
		{
			name: "Invalid ACL",
			fields: fields{
				File:      strings.NewReader("test"),
				Directory: "test",
				Filename:  "test.jpg",
				ACL:       "test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &UploadInput{
				File:      tt.fields.File,
				Directory: tt.fields.Directory,
				Filename:  tt.fields.Filename,
				ACL:       tt.fields.ACL,
			}
			err := i.Validate()
			assert.EqualValues(t, tt.wantErr, err != nil, err)
		})
	}
}

package fileRepo

import (
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
)

type Repo = repo

func (r *repo) BucketName() string {
	return r.bucketName
}

func (r *repo) SetBucketName(bucketName string) {
	r.bucketName = bucketName
}

func (r *repo) Uploader() s3manageriface.UploaderAPI {
	return r.uploader
}

func (r *repo) SetUploader(uploader s3manageriface.UploaderAPI) {
	r.uploader = uploader
}

func (r *repo) Deleter() Deleter {
	return r.deleter
}

func (r *repo) SetDeleter(deleter Deleter) {
	r.deleter = deleter
}

func (r *repo) BatchDeleter() s3manageriface.BatchDelete {
	return r.batchDeleter
}

func (r *repo) SetBatchDeleter(batchDeleter s3manageriface.BatchDelete) {
	r.batchDeleter = batchDeleter
}

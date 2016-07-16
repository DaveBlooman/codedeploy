package storage

import (
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type customError struct {
	code    int
	event   string
	problem string
}

func (e *customError) Error() string {
	return e.problem
}

func (e *customError) Event() string {
	return e.event
}

func (e *customError) StatusCode() int {
	return e.code
}

type Storage interface {
	get(bucket, path string) (*s3.ListObjectsOutput, error)
	put(bucket, path string, fileType string, data io.ReadSeeker, size int64) (resp *s3.PutObjectOutput, err error)
}

type S3Storage struct{}

func (s *S3Storage) get(bucket, path string) (*s3.ListObjectsOutput, error) {
	svc := s3.New(session.New(config))

	params := &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(path),
	}

	return svc.ListObjects(params)

}

func (s *S3Storage) put(bucket, path string, fileType string, data io.ReadSeeker, size int64) (resp *s3.PutObjectOutput, err error) {
	svc := s3.New(session.New(config))

	params := &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(path),
		Body:          data,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(fileType),
		CacheControl:  aws.String("max-age=31536000, public"),
	}

	return svc.PutObject(params)
}

var (
	err     error
	config  *aws.Config
	storage = &S3Storage{}
)

func List(region, bucket, path string) (resp *s3.ListObjectsOutput, customErr *customError) {
	config = &aws.Config{
		Region: aws.String(region),
	}

	awsResponse, err := storage.get(bucket, path)
	return awsResponse, handleError(err)
}

func Put(region, bucket, path string, fileType string, data io.ReadSeeker, size int64, s3config *aws.Config) (resp *s3.PutObjectOutput, customErr *customError) {
	config = s3config

	awsResponse, err := storage.put(bucket, path, fileType, data, size)

	return awsResponse, handleError(err)
}

func handleError(err error) *customError {
	var scopedVariable *customError

	if err != nil {
		scopedVariable = &customError{
			500,
			"Error",
			err.Error(),
		}

		if awsErr, ok := err.(awserr.Error); ok {
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				scopedVariable = &customError{
					reqErr.StatusCode(),
					"AWS Service Error",
					awsErr.Message(),
				}
			} else {
				scopedVariable = &customError{
					500,
					"AWS Error",
					awsErr.Message(),
				}
			}
		}
	}

	return scopedVariable
}

package storage

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rijum8906/go-micro-service/packages/common/errors"
)

type S3StorageService interface {
	// Bucket operations
	CreateBucket(ctx context.Context, bucketName string) *errors.AppError
	IsBucketExists(ctx context.Context, bucketName string) (bool, *errors.AppError)

	// file operations
	UploadFile(ctx context.Context, fileName string, file io.Reader, contentType string) (string, *errors.AppError)
}

type S3Storage struct {
	client     *s3.Client
	bucketName string
	publicURL  string
}

func NewS3StorageService(ctx context.Context, endpoint, accessKey, secretKey, bucket, publicURL string) S3StorageService {
	cfg, err := config.LoadDefaultConfig(ctx,

		config.WithRegion("us-east-1"), // MinIO requires a region string, even if ignored
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		panic(err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true // Essential for MinIO/Localhost
	})

	return &S3Storage{
		client:     client,
		bucketName: bucket,
		publicURL:  publicURL,
	}
}

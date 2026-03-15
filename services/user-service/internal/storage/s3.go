// Package storage contains storage for the auth service.
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

// Bucket operations

func (s *S3Storage) CreateBucket(ctx context.Context, bucketName string) *errors.AppError {
	_, err := s.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return errors.NewAppError(500, "error creating bucket", []errors.Error{}).WithInternal(err)
	}
	return nil
}

func (s *S3Storage) IsBucketExists(ctx context.Context, bucketName string) (bool, *errors.AppError) {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return false, nil
	}
	return true, nil
}

// File operations

func (s *S3Storage) UploadFile(ctx context.Context, fileName string, file io.Reader, contentType string) (string, *errors.AppError) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(fileName),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", errors.NewAppError(500, "error uploading file", []errors.Error{}).WithInternal(err)
	}
	return s.publicURL + fileName, nil
}

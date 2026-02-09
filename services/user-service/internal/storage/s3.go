// Package storage contains storage for the auth service.
package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rijum8906/go-micro-service/packages/common/errors"
)

type S3Storage struct {
	client     *s3.Client
	bucketName string
	publicURL  string
}

func NewS3Storage(ctx context.Context, endpoint, accessKey, secretKey, bucket, publicURL string) (*S3Storage, *errors.AppError) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"), // MinIO requires a region string, even if ignored
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, errors.NewAppError(500, "error loading AWS config", []errors.Error{}).WithInternal(err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true // Essential for MinIO/Localhost
	})

	return &S3Storage{
		client:     client,
		bucketName: bucket,
		publicURL:  publicURL,
	}, nil
}

func (s *S3Storage) UploadFile(ctx context.Context, fileName string, file io.Reader, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(fileName),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", err
	}

	// Returns the link for your database
	return fmt.Sprintf("%s/%s/%s", s.publicURL, s.bucketName, fileName), nil
}

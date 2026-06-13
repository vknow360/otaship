package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Provider struct {
	s3       *s3.Client
	bucket   string
	region   string
	basePath string
}

func NewS3Provider() (*S3Provider, error) {
	accessKey := os.Getenv("S3_ACCESS_KEY")
	secretKey := os.Getenv("S3_SECRET_ACCESS_KEY")
	region := os.Getenv("S3_REGION")
	bucket := os.Getenv("S3_BUCKET_NAME")
	endpoint := os.Getenv("S3_ENDPOINT")
	basePath := os.Getenv("S3_BASE_PATH")
	if accessKey == "" || secretKey == "" || region == "" || bucket == "" {
		return nil, fmt.Errorf("missing s3 credentials")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			"",
		)),
		config.WithRegion(region),
	)

	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		}
	})

	return &S3Provider{s3: client, bucket: bucket, region: region, basePath: basePath}, nil
}

func (s *S3Provider) Name() string {
	return "s3"
}

func (s *S3Provider) Upload(
	ctx context.Context,
	key string,
	data io.Reader,
	contentType string,
	size int64,
) (string, error) {
	_, err := s.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        &s.bucket,
		Key:           &key,
		Body:          data,
		ContentType:   &contentType,
		ContentLength: &size,
	})

	if err != nil {
		return "", err
	}

	if s.basePath != "" {
		return fmt.Sprintf("%s/%s/%s", s.basePath, s.bucket, key), nil
	}

	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, key), nil
}

func (s *S3Provider) Delete(ctx context.Context, key, mimeType string) error {
	_, err := s.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *S3Provider) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.s3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})

	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *S3Provider) Ping(ctx context.Context) error {
	_, err := s.s3.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: &s.bucket,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *S3Provider) Usage(ctx context.Context) (any, error) {
	return nil, errors.New("Not implemented")
}

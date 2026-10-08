package cloud

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	myconfig "github.com/ismelen/inkomi/back/epub-worker/internal/config"
)

type CloudflareR2Storage struct {
	client *s3.Client
	bucket string
}

func NewCloudflareR2Storage(ctx context.Context) (*CloudflareR2Storage, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			myconfig.Env.R2AccessKeyId,
			myconfig.Env.R2SecretAccessKey,
			"",
		)),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", myconfig.Env.R2AccountId))
	})

	return &CloudflareR2Storage{
		client: client,
		bucket: myconfig.Env.R2BucketName,
	}, nil
}

func (s *CloudflareR2Storage) Upload(ctx context.Context, id string, filename string, body io.Reader) error {
	key := fmt.Sprintf("%s/%s", id, filename)
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   body,
	})
	return err
}

func (s *CloudflareR2Storage) Download(ctx context.Context, id string, filename string) (io.ReadCloser, error) {
	key := fmt.Sprintf("%s/%s", id, filename)
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func (s *CloudflareR2Storage) Delete(ctx context.Context, id string, filename string) error {
	key := fmt.Sprintf("%s/%s", id, filename)
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

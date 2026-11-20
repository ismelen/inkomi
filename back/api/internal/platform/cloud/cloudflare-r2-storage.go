package cloud

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

var _ ICloudStorage = (*CloudflareR2Storage)(nil)

// CloudflareR2Storage stores every file under the key "id/filename".
type CloudflareR2Storage struct {
	client *s3.Client
	bucket string
}

func NewCloudflareR2Storage(ctx context.Context, accountId, accessKeyId, secretAccessKey, bucket string) (*CloudflareR2Storage, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("load r2 config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId))
	})

	return &CloudflareR2Storage{
		client: client,
		bucket: bucket,
	}, nil
}

func objectKey(id, filename string) *string {
	return aws.String(id + "/" + filename)
}

func (s *CloudflareR2Storage) Upload(ctx context.Context, id string, filename string, body io.Reader) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    objectKey(id, filename),
		Body:   body,
	})
	return err
}

func (s *CloudflareR2Storage) Download(ctx context.Context, id string, filename string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    objectKey(id, filename),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func (s *CloudflareR2Storage) UploadFile(ctx context.Context, id string, filename string, path string) (int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return 0, fmt.Errorf("stat %s: %w", path, err)
	}

	if err := s.Upload(ctx, id, filename, file); err != nil {
		return 0, fmt.Errorf("upload %s/%s: %w", id, filename, err)
	}
	return info.Size(), nil
}

func (s *CloudflareR2Storage) DownloadFile(ctx context.Context, id string, filename string, path string) error {
	reader, err := s.Download(ctx, id, filename)
	if err != nil {
		return fmt.Errorf("download %s/%s: %w", id, filename, err)
	}
	defer reader.Close()

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}

	if _, err := io.Copy(file, reader); err != nil {
		return errors.Join(fmt.Errorf("write %s: %w", path, err), file.Close())
	}
	return file.Close()
}

func (s *CloudflareR2Storage) Size(ctx context.Context, id string, filename string) (int64, error) {
	out, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    objectKey(id, filename),
	})
	if err != nil {
		return 0, err
	}
	return aws.ToInt64(out.ContentLength), nil
}

func (s *CloudflareR2Storage) Delete(ctx context.Context, id string, filename string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    objectKey(id, filename),
	})
	return err
}

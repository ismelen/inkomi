package cloud

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type CloudflareR2Storage struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucketName    string
}

func NewCloudflareR2Storage(accountId, accessKeyId, secretAccessKey, bucketName string) (*CloudflareR2Storage, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId))
	})

	presignClient := s3.NewPresignClient(client)

	return &CloudflareR2Storage{
		client:        client,
		presignClient: presignClient,
		bucketName:    bucketName,
	}, nil
}

func (c *CloudflareR2Storage) GetUrl(id string, ext string) (string, error) {
	key := id + ext
	request, err := c.presignClient.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 15 * time.Minute
	})
	if err != nil {
		return "", err
	}

	return request.URL, nil
}

func (c *CloudflareR2Storage) Check(id string, ext string) bool {
	key := id + ext
	_, err := c.client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	})
	return err == nil
}

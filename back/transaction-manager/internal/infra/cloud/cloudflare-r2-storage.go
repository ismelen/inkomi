package cloud

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/ports"
)

const presignExpiry = 15 * time.Minute

var _ ports.ICloudStorage = (*CloudflareR2Storage)(nil)

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

func (c *CloudflareR2Storage) GetUrl(id string, filename string) (string, error) {
	key := id + "/" + filename
	request, err := c.presignClient.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = presignExpiry
	})
	if err != nil {
		return "", err
	}

	return request.URL, nil
}

func (c *CloudflareR2Storage) GetDownloadUrl(id string, filename string) (string, time.Time, error) {
	key := id + "/" + filename
	expiresAt := time.Now().Add(presignExpiry).UTC()
	request, err := c.presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket:                     aws.String(c.bucketName),
		Key:                        aws.String(key),
		ResponseContentDisposition: aws.String(attachmentDisposition(filename)),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = presignExpiry
	})
	if err != nil {
		return "", time.Time{}, err
	}

	return request.URL, expiresAt, nil
}

func (c *CloudflareR2Storage) Check(id string, filename string) bool {
	key := id + "/" + filename
	_, err := c.client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	})
	return err == nil
}

// attachmentDisposition encodes filename as RFC 5987 so non-ASCII titles survive.
func attachmentDisposition(filename string) string {
	encoded := strings.ReplaceAll(url.QueryEscape(filename), "+", "%20")
	return "attachment; filename*=UTF-8''" + encoded
}

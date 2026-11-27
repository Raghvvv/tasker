package aws

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Raghvvv/tasker/internal/server"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Client struct {
	server *server.Server
	client *s3.Client
}

func NewS3Client(s *server.Server, config *aws.Config) *S3Client {
	return &S3Client{
		server: s,
		client: s3.NewFromConfig(*config, func(o *s3.Options) {
			if s.Config.AWS.EndpointURL != "" {
				o.BaseEndpoint = aws.String(s.Config.AWS.EndpointURL)
				o.UsePathStyle = true
			}
		}),
	}
}

func (s *S3Client) UploadFiles(ctx context.Context, bucket string, fileName string, file io.Reader) (string, error) {
	fileKey := fmt.Sprintf("%s_%d", fileName, time.Now().Unix())

	var buffer bytes.Buffer
	_, err := io.Copy(&buffer, file)
	if err != nil {
		return "", fmt.Errorf("failed to read file%w", err)
	}

	_, error := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(fileKey),
		Body:        bytes.NewBuffer(buffer.Bytes()),
		ContentType: aws.String(http.DetectContentType(buffer.Bytes())),
	})
	if error != nil {
		return "", fmt.Errorf("failed to upload file to s3 %w", error)
	}

	return fileKey, nil
}

func (s *S3Client) CreatePresignedURL(ctx context.Context, bucket string, objectKey string) (string, error) {
	presignClient := s3.NewPresignClient(s.client)
	expirey := time.Minute * 60
	presignurl, error := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
	},
		s3.WithPresignExpires(expirey),
	)

	if error != nil {
		return "", fmt.Errorf("failed to get url %w", error)
	}
	return presignurl.URL, nil

}

func (s *S3Client) DeleteAttatchment(ctx context.Context, bucket string, objectKey string) error {
	_, error := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
	},
	)
	if error != nil {
		return fmt.Errorf("failed to delete object from bucket%w", error)
	}
	return nil
}

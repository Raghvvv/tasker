package aws

import (
	"context"
	"fmt"

	"github.com/Raghvvv/tasker/internal/server"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

type AWS struct {
	S3 *S3Client
}

func NewAWS(server *server.Server) (*AWS, error) {
	awsConfig := server.Config.AWS

	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(awsConfig.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			awsConfig.AccessKeyID,
			awsConfig.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &AWS{
		S3: NewS3Client(server, &cfg),
	}, nil
}

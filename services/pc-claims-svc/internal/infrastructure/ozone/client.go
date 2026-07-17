package ozone

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type OzoneClient struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucketName    string
}

// NewOzoneClient initializes the S3-compatible client for Apache Ozone
func NewOzoneClient(ctx context.Context, endpointURL string) (*OzoneClient, error) {
	// Configure the custom endpoint resolver to point to Ozone's OM/S3 Gateway
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:           endpointURL,
			SigningRegion: "us-east-1", // Dummy region required by signature v4
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config for Ozone: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true // Ozone requires path-style addressing
	})

	presignClient := s3.NewPresignClient(client)

	return &OzoneClient{
		client:        client,
		presignClient: presignClient,
		bucketName:    "vol-medhen-claims/bucket-evidence",
	}, nil
}

// GeneratePresignedUploadURL creates a time-bound URL for direct client upload
func (c *OzoneClient) GeneratePresignedUploadURL(ctx context.Context, tenantID, claimID, filename string) (string, error) {
	objectKey := fmt.Sprintf("%s/%s/evidence/%s", tenantID, claimID, filename)

	req, err := c.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 15 * time.Minute
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return req.URL, nil
}

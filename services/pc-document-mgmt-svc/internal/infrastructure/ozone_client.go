package infrastructure

import (
	"context"
	"fmt"
	"io"
	"github.com/inno-sphere/medhen-prototype/services/pc-document-mgmt-svc/internal/domain"
)

type OzoneS3Client struct {
	endpoint string
	volume   string
}

func NewOzoneS3Client(endpoint, volume string) *OzoneS3Client {
	return &OzoneS3Client{endpoint: endpoint, volume: volume}
}

func (c *OzoneS3Client) UploadStream(ctx context.Context, bucket, path string, reader io.Reader, size int64, mimeType string) (domain.StorageRef, error) {
	// Stub implementation for Apache Ozone via AWS SDK
	fmt.Printf("Uploading %d bytes to Ozone bucket %s path %s\n", size, bucket, path)

	return domain.StorageRef{
		Volume:   c.volume,
		Bucket:   bucket,
		Path:     path,
		MimeType: mimeType,
	}, nil
}

func (c *OzoneS3Client) DownloadStream(ctx context.Context, ref domain.StorageRef) (io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented")
}

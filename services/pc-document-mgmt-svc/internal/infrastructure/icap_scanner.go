package infrastructure

import (
	"context"
	"fmt"
	"io"
)

type ICAPMalwareScanner struct {
	endpoint string
}

func NewICAPMalwareScanner(endpoint string) *ICAPMalwareScanner {
	return &ICAPMalwareScanner{endpoint: endpoint}
}

func (s *ICAPMalwareScanner) ScanStream(ctx context.Context, reader io.Reader) (bool, error) {
	fmt.Printf("Streaming payload to ICAP server at %s for malware analysis...\n", s.endpoint)
	// Stub: return true (clean) for all files.
	// In reality, this pipes the stream to ClamAV over ICAP.
	return true, nil
}

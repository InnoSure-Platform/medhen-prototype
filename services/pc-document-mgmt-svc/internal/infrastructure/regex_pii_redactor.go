package infrastructure

import (
	"bytes"
	"context"
	"io"
	"regexp"
)

type RegexPIIRedactor struct {
	patterns []*regexp.Regexp
}

func NewRegexPIIRedactor() *RegexPIIRedactor {
	return &RegexPIIRedactor{
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b`), // Email
			regexp.MustCompile(`\b(?:\+?251|0)?[97]\d{8}\b`),                           // Ethiopian Phone
			regexp.MustCompile(`\b\d{4}[ -]?\d{4}[ -]?\d{4}[ -]?\d{4}\b`),              // Credit Card
		},
	}
}

type redactedReader struct {
	inner io.ReadCloser
	buf   *bytes.Buffer
}

func (r *redactedReader) Read(p []byte) (n int, err error) {
	return r.buf.Read(p)
}

func (r *redactedReader) Close() error {
	return r.inner.Close()
}

func (r *RegexPIIRedactor) RedactStream(ctx context.Context, reader io.Reader) (io.ReadCloser, error) {
	// For Tier-0, a real streaming redactor (e.g. state machine based) would be used.
	// This stub reads all into memory for regex application.
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	for _, pattern := range r.patterns {
		data = pattern.ReplaceAll(data, []byte("[REDACTED]"))
	}

	rc, ok := reader.(io.ReadCloser)
	if !ok {
		rc = io.NopCloser(reader)
	}

	return &redactedReader{
		inner: rc,
		buf:   bytes.NewBuffer(data),
	}, nil
}

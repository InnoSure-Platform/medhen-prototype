package job

import (
	"context"
	"log/slog"
	"time"

	"github.com/medhen/pc-reporting-svc/internal/application/extractor"
)

type ReportJob struct {
	JobID    string
	TenantID string
	Year     int
	Quarter  int
}

type ReportWorkerPool struct {
	jobs      chan ReportJob
	extractor *extractor.NBEMotorExtractor
}

func NewReportWorkerPool(ext *extractor.NBEMotorExtractor, workers int) *ReportWorkerPool {
	pool := &ReportWorkerPool{
		jobs:      make(chan ReportJob, 100),
		extractor: ext,
	}

	for i := 0; i < workers; i++ {
		go pool.worker()
	}

	return pool
}

func (p *ReportWorkerPool) Submit(job ReportJob) {
	p.jobs <- job
}

func (p *ReportWorkerPool) worker() {
	for job := range p.jobs {
		slog.Info("processing report job", "job_id", job.JobID, "tenant_id", job.TenantID)
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		
		// Execute Heavy Extraction
		doc, err := p.extractor.Extract(ctx, job.TenantID, job.Year, job.Quarter)
		cancel()

		if err != nil {
			slog.Error("failed to extract report", "job_id", job.JobID, "error", err)
			continue
		}

		// In a real implementation, this would upload `doc.Payload` to S3 and update a database ledger
		// with `doc.Hash` and the S3 URL.
		slog.Info("report extraction completed successfully", "job_id", job.JobID, "hash", doc.Hash, "size_bytes", len(doc.Payload))
	}
}

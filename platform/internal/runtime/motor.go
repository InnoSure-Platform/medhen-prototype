package runtime

import (
	"context"
	"os"
	"path/filepath"

	"github.com/redis/go-redis/v9"

	"github.com/InnoSure-Platform/pc-platform/internal/integration"
	"github.com/InnoSure-Platform/pc-platform/internal/pdf"
	"github.com/InnoSure-Platform/pc-platform/internal/product"
	"github.com/InnoSure-Platform/pc-platform/internal/store"
	"github.com/InnoSure-Platform/pc-platform/internal/usecase"
	"github.com/InnoSure-Platform/pc-shared-go/idempotency"
)

// BuildMotor constructs a Motor use-case with production deps from env.
func BuildMotor(ctx context.Context) *usecase.Motor {
	repo := OpenStore(ctx)
	_ = OpenKafka(ctx, repo)
	docsDir := Env("MEDHEN_DOCS_DIR", "./data/docs")
	_ = os.MkdirAll(docsDir, 0o755)
	m := &usecase.Motor{
		Repo:    repo,
		Product: product.SeedMotor(),
		Pay:     integration.NewTelebirrFromEnv(),
		SMS:     &integration.MockSMS{},
		PDF:     pdf.NewGenerator(docsDir, "/files"),
		Fayda:   integration.MockFayda{},
	}
	if url := os.Getenv("REDIS_URL"); url != "" {
		opt, err := redis.ParseURL(url)
		if err == nil {
			m.Idem = idempotency.NewStore(redis.NewClient(opt), 0)
		}
	} else {
		m.Idem = idempotency.NewMemoryStore()
	}
	return m
}

// MotorFromRepo builds a read-focused Motor use-case for audit-svc.
func MotorFromRepo(repo store.Repository) *usecase.Motor {
	return &usecase.Motor{Repo: repo, Product: product.SeedMotor()}
}

// FileServer returns the directory for generated PDFs.
func FileServer(docsDir string) string {
	if docsDir == "" {
		docsDir = Env("MEDHEN_DOCS_DIR", "./data/docs")
	}
	return filepath.Clean(docsDir)
}

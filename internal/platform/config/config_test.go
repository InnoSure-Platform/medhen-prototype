package config

import (
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	for _, k := range []string{
		"MEDHEN_ENV", "MEDHEN_HTTP_ADDR", "MEDHEN_SHUTDOWN_TIMEOUT", "MEDHEN_READ_TIMEOUT",
		"MEDHEN_WRITE_TIMEOUT", "DATABASE_URL", "MEDHEN_OUTBOX_POLL", "TELEBIRR_WEBHOOK_SECRET",
	} {
		t.Setenv(k, "")
	}
	cfg := Load()
	if cfg.Env != "dev" || cfg.HTTPAddr != ":8080" {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if cfg.OutboxPollInterval != time.Second || cfg.ReadTimeout != 15*time.Second {
		t.Fatalf("unexpected duration defaults: %+v", cfg)
	}
	if cfg.DatabaseURL != "" || cfg.TelebirrWebhookSecret != "" {
		t.Fatalf("secrets should be empty by default: %+v", cfg)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	t.Setenv("MEDHEN_ENV", "prod")
	t.Setenv("MEDHEN_HTTP_ADDR", ":9000")
	t.Setenv("MEDHEN_OUTBOX_POLL", "500ms")
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("TELEBIRR_WEBHOOK_SECRET", "s3cr3t")

	cfg := Load()
	if cfg.Env != "prod" || cfg.HTTPAddr != ":9000" {
		t.Fatalf("env not applied: %+v", cfg)
	}
	if cfg.OutboxPollInterval != 500*time.Millisecond {
		t.Fatalf("duration override failed: %v", cfg.OutboxPollInterval)
	}
	if cfg.DatabaseURL != "postgres://x" || cfg.TelebirrWebhookSecret != "s3cr3t" {
		t.Fatalf("secret env not applied: %+v", cfg)
	}
}

func TestGetdur_InvalidFallsBackToDefault(t *testing.T) {
	t.Setenv("MEDHEN_READ_TIMEOUT", "not-a-duration")
	if cfg := Load(); cfg.ReadTimeout != 15*time.Second {
		t.Fatalf("invalid duration should fall back to default, got %v", cfg.ReadTimeout)
	}
}

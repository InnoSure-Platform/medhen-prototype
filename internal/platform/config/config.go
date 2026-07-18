// Package config loads the monolith's runtime configuration from the
// environment. Every value has a safe default for local development; secrets
// (DB URLs, Keycloak, Telebirr) are added as modules are migrated in.
package config

import (
	"os"
	"time"
)

// Config is the composition-root configuration for cmd/medhen-api.
type Config struct {
	// Env is the deployment environment name ("dev", "staging", "prod").
	Env string
	// HTTPAddr is the listen address for the HTTP edge.
	HTTPAddr string
	// ShutdownTimeout bounds graceful shutdown.
	ShutdownTimeout time.Duration
	// ReadTimeout / WriteTimeout guard the HTTP server.
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// Load reads configuration from the environment.
func Load() Config {
	return Config{
		Env:             getenv("MEDHEN_ENV", "dev"),
		HTTPAddr:        getenv("MEDHEN_HTTP_ADDR", ":8080"),
		ShutdownTimeout: getdur("MEDHEN_SHUTDOWN_TIMEOUT", 15*time.Second),
		ReadTimeout:     getdur("MEDHEN_READ_TIMEOUT", 15*time.Second),
		WriteTimeout:    getdur("MEDHEN_WRITE_TIMEOUT", 30*time.Second),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getdur(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

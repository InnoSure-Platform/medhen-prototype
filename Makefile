# Medhen Makefile — single modular-monolith module.
.PHONY: build api test test-integration test-cover cover lint arch-lint vet web infra-up infra-down certs

# Build the monolith binary.
build:
	@mkdir -p bin
	go build -o bin/medhen-api ./cmd/medhen-api

# Run the monolith. DB-backed modules activate when DATABASE_URL is set;
# otherwise only the stateless modules run.
api: build
	MEDHEN_DOCS_DIR=./data/docs ./bin/medhen-api

# Unit tests (fast; testcontainers integration tests are skipped via -short).
test:
	go test -short -race ./...

# Full suite including testcontainers integration tests (requires Docker).
test-integration:
	go test -race ./...

# Full suite with cross-package coverage profile (requires Docker).
test-cover:
	go test -coverpkg=./internal/... -coverprofile=coverage.out ./...

# Enforce per-layer coverage floors on the profile from `make test-cover`.
cover: test-cover
	bash scripts/coverage-check.sh coverage.out

# Static analysis (config: .golangci.yml).
lint:
	golangci-lint run ./...

# Sealed-module boundary check (config: .go-arch-lint.yml).
arch-lint:
	go-arch-lint check --project-path .

vet:
	go vet ./...

web:
	cd web && npm run dev

# Local infra backbone (Postgres, Valkey, Kafka, Keycloak) for the monolith.
infra-up:
	cd infra && docker compose up -d

infra-down:
	cd infra && docker compose down

# Generate self-signed dev TLS certs (output gitignored under certs/).
certs:
	./scripts/gen-dev-certs.sh

# Medhen Makefile — single modular-monolith module.
.PHONY: build api test test-integration vet web infra-up infra-down certs

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

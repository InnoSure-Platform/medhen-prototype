# Medhen Makefile
.PHONY: api mesh mesh-smoke mesh-down test e2e web infra-up infra-down build-all demo-rehearse telebirr-prove

build-all:
	cd platform && go build -o ../bin/medhen-api ./cmd/medhen-api && \
	go build -o ../bin/pc-gateway ./cmd/pc-gateway && \
	go build -o ../bin/pc-party-mgmt-svc ./cmd/pc-party-mgmt-svc && \
	go build -o ../bin/pc-policy-svc ./cmd/pc-policy-svc && \
	go build -o ../bin/pc-billing-svc ./cmd/pc-billing-svc && \
	go build -o ../bin/pc-claims-svc ./cmd/pc-claims-svc && \
	go build -o ../bin/pc-audit-svc ./cmd/pc-audit-svc && \
	go build -o ../bin/pc-integration-svc ./cmd/pc-integration-svc

api: build-all
	MEDHEN_DOCS_DIR=./data/docs ./bin/medhen-api

mesh: build-all
	./scripts/mesh-up.sh

mesh-smoke:
	chmod +x scripts/*.sh
	./scripts/mesh-smoke.sh

mesh-down:
	./scripts/mesh-down.sh

demo-rehearse: mesh-smoke
	@echo "See docs/demo/DEMO-RUNBOOK.md for facilitator script"

telebirr-prove:
	./scripts/telebirr-prove.sh

test:
	cd shared/go && go test ./...
	cd platform && go test ./...

e2e:
	./scripts/demo-e2e.sh

web:
	cd web && npm run dev

infra-up:
	cd infra && docker compose up -d

infra-down:
	cd infra && docker compose down

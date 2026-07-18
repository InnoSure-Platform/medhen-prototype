# Medhen Makefile
.PHONY: api mesh mesh-smoke mesh-down test e2e web infra-up infra-down build-all demo-rehearse telebirr-prove

build-all:
	@mkdir -p bin
	@for d in services/*; do \
		if [ -d "$$d" ]; then \
			svc=$$(basename "$$d"); \
			if [ -d "$$d/cmd/server" ]; then \
				echo "Building $$svc..."; \
				(cd "$$d" && go build -o ../../bin/$$svc ./cmd/server); \
			elif [ -d "$$d/cmd/api" ]; then \
				echo "Building $$svc..."; \
				(cd "$$d" && go build -o ../../bin/$$svc ./cmd/api); \
			fi; \
		fi; \
	done

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

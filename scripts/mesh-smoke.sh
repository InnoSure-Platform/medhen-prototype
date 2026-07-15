#!/usr/bin/env bash
# Full mesh smoke: infra → mesh → Keycloak JWT → E2E on Postgres/Kafka.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

export DATABASE_URL="${DATABASE_URL:-postgres://medhen:medhen@localhost:5432/medhen?sslmode=disable}"
export PG_SCHEMA="${PG_SCHEMA:-pc_medhen}"
export KAFKA_BROKERS="${KAFKA_BROKERS:-localhost:19092}"
export REDIS_URL="${REDIS_URL:-redis://localhost:6379/0}"
export KEYCLOAK_URL="${KEYCLOAK_URL:-http://localhost:8081}"
export KEYCLOAK_REALM="${KEYCLOAK_REALM:-medhen}"
export MEDHEN_DOCS_DIR="${MEDHEN_DOCS_DIR:-$ROOT/data/docs}"
export MEDHEN_MESH=1

echo "== 1. Start infra =="
make infra-up

echo "== 2. Wait for backbone =="
wait_for() {
  local name=$1 cmd=$2
  for i in $(seq 1 60); do
    if eval "$cmd" >/dev/null 2>&1; then
      echo "  $name ready"
      return 0
    fi
    sleep 2
  done
  echo "  $name NOT ready" >&2
  return 1
}

wait_for "Postgres" "docker compose -f infra/docker-compose.yml exec -T postgres pg_isready -U medhen"
wait_for "Valkey" "docker compose -f infra/docker-compose.yml exec -T valkey valkey-cli ping | grep -q PONG"
wait_for "Keycloak" "curl -sf http://localhost:8081/realms/medhen/.well-known/openid-configuration"
wait_for "Redpanda" "curl -sf http://localhost:9644/v1/status/ready"

echo "== 3. Build + start mesh =="
make build-all
./scripts/mesh-down.sh 2>/dev/null || true
./scripts/mesh-up.sh
sleep 2
wait_for "Gateway" "curl -sf http://localhost:8080/health"

echo "== 4. Keycloak token =="
TOKEN=$("${ROOT}/scripts/keycloak-token.sh")
echo "  token length: ${#TOKEN}"

echo "== 5. E2E storyboard (Postgres + JWT) =="
export KEYCLOAK_URL
"${ROOT}/scripts/demo-e2e.sh"

echo "== 6. Verify Postgres persistence =="
COUNT=$(docker compose -f infra/docker-compose.yml exec -T postgres \
  psql -U medhen -d medhen -t -c "SET search_path TO ${PG_SCHEMA}; SELECT count(*) FROM policies WHERE status='ISSUED';" | tr -d ' ')
if [[ "${COUNT:-0}" -lt 1 ]]; then
  echo "Expected at least 1 issued policy in Postgres, got: $COUNT" >&2
  exit 1
fi
echo "  issued policies in DB: $COUNT"

echo "== 7. Verify PDF documents =="
PDF_COUNT=$(find "$MEDHEN_DOCS_DIR" -name 'EIC-MOT-*.pdf' 2>/dev/null | wc -l | tr -d ' ')
if [[ "${PDF_COUNT:-0}" -lt 1 ]]; then
  echo "Expected PDF documents in $MEDHEN_DOCS_DIR" >&2
  exit 1
fi
echo "  PDF files: $PDF_COUNT"

echo "== 8. Kafka outbox (optional) =="
OUTBOX=$(docker compose -f infra/docker-compose.yml exec -T postgres \
  psql -U medhen -d medhen -t -c "SET search_path TO ${PG_SCHEMA}; SELECT count(*) FROM outbox WHERE published_at IS NOT NULL;" 2>/dev/null | tr -d ' ' || echo "0")
echo "  published outbox rows: ${OUTBOX:-0}"

echo ""
echo "MESH SMOKE PASSED"

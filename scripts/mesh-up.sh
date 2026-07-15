#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT/platform"

if [[ -f "$ROOT/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "$ROOT/.env"
  set +a
fi

export DATABASE_URL="${DATABASE_URL:-postgres://medhen:medhen@localhost:5432/medhen?sslmode=disable}"
export PG_SCHEMA="${PG_SCHEMA:-pc_medhen}"
export KAFKA_BROKERS="${KAFKA_BROKERS:-localhost:19092}"
export REDIS_URL="${REDIS_URL:-redis://localhost:6379/0}"
export MEDHEN_DOCS_DIR="${MEDHEN_DOCS_DIR:-$ROOT/data/docs}"
export MEDHEN_MESH=1
export KEYCLOAK_URL="${KEYCLOAK_URL:-http://localhost:8081}"
export KEYCLOAK_REALM="${KEYCLOAK_REALM:-medhen}"

mkdir -p "$MEDHEN_DOCS_DIR" "$ROOT/logs"

build() {
  go build -o "$ROOT/bin/$1" "./cmd/$1"
}

for svc in pc-party-mgmt-svc pc-policy-svc pc-billing-svc pc-claims-svc pc-audit-svc pc-integration-svc pc-gateway; do
  build "$svc"
done

start_svc() {
  local name=$1
  local addr=$2
  echo "Starting $name on $addr"
  env DATABASE_URL="$DATABASE_URL" PG_SCHEMA="$PG_SCHEMA" KAFKA_BROKERS="$KAFKA_BROKERS" \
    REDIS_URL="$REDIS_URL" MEDHEN_DOCS_DIR="$MEDHEN_DOCS_DIR" KEYCLOAK_URL="$KEYCLOAK_URL" \
    KEYCLOAK_REALM="$KEYCLOAK_REALM" MEDHEN_MESH=1 MEDHEN_ADDR="$addr" \
    "$ROOT/bin/$name" > "$ROOT/logs/$name.log" 2>&1 &
  echo $! > "$ROOT/logs/$name.pid"
}

start_svc pc-party-mgmt-svc :8101
start_svc pc-policy-svc :8103
start_svc pc-billing-svc :8107
start_svc pc-claims-svc :8106
start_svc pc-audit-svc :8117
start_svc pc-integration-svc :8118
sleep 2
start_svc pc-gateway :8080

echo "Mesh up — gateway http://localhost:8080 (JWT via Keycloak when KEYCLOAK_URL set)"
echo "Logs: $ROOT/logs/"

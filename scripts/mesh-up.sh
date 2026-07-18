#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

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

mkdir -p "$MEDHEN_DOCS_DIR" "$ROOT/logs" "$ROOT/bin"

echo "Building services..."
for d in "$ROOT"/services/*; do
  if [ -d "$d" ]; then
    svc=$(basename "$d")
    echo "Building $svc..."
    if [ -d "$d/cmd/server" ]; then
      (cd "$d" && go build -o "$ROOT/bin/$svc" "./cmd/server")
    elif [ -d "$d/cmd/api" ]; then
      (cd "$d" && go build -o "$ROOT/bin/$svc" "./cmd/api")
    else
      echo "Warning: No cmd/server or cmd/api found in $d"
    fi
  fi
done

start_svc() {
  local name=$1
  local addr=$2
  if [ ! -f "$ROOT/bin/$name" ]; then
    echo "Skipping $name (not built)"
    return
  fi
  echo "Starting $name on $addr"
  env DATABASE_URL="$DATABASE_URL" PG_SCHEMA="$PG_SCHEMA" KAFKA_BROKERS="$KAFKA_BROKERS" \
    REDIS_URL="$REDIS_URL" MEDHEN_DOCS_DIR="$MEDHEN_DOCS_DIR" KEYCLOAK_URL="$KEYCLOAK_URL" \
    KEYCLOAK_REALM="$KEYCLOAK_REALM" MEDHEN_MESH=1 MEDHEN_ADDR="$addr" \
    "$ROOT/bin/$name" > "$ROOT/logs/$name.log" 2>&1 &
  echo $! > "$ROOT/logs/$name.pid"
}

port=8101
for d in "$ROOT"/services/*; do
  if [ -d "$d" ]; then
    svc=$(basename "$d")
    start_svc "$svc" ":$port"
    port=$((port + 1))
  fi
done

echo "Starting Envoy proxy for pc-gateway..."
if command -v envoy >/dev/null 2>&1; then
  envoy -c "$ROOT/infra/pc-gateway/envoy.yaml" > "$ROOT/logs/pc-gateway.log" 2>&1 &
  echo $! > "$ROOT/logs/pc-gateway.pid"
else
  echo "Warning: envoy not found in PATH. Please install Envoy or run pc-gateway using docker."
fi

echo "Mesh up — gateway Envoy on port 8080 (assumed)"
echo "Logs: $ROOT/logs/"

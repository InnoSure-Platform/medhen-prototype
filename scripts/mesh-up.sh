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

# Generate Envoy configuration dynamically
mkdir -p "$ROOT/infra/pc-gateway"
cat <<EOF > "$ROOT/infra/pc-gateway/envoy.yaml"
static_resources:
  listeners:
  - name: listener_0
    address:
      socket_address: { address: 0.0.0.0, port_value: 8080 }
    filter_chains:
    - filters:
      - name: envoy.filters.network.http_connection_manager
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
          stat_prefix: ingress_http
          route_config:
            name: local_route
            virtual_hosts:
            - name: local_service
              domains: ["*"]
              routes:
              - match: { prefix: "/auth/" }
                route: { cluster: cluster_keycloak }
EOF

port=8101
for d in "$ROOT"/services/*; do
  if [ -d "$d" ]; then
    svc=$(basename "$d")
    start_svc "$svc" ":$port"
    
    cat <<EOF >> "$ROOT/infra/pc-gateway/envoy.yaml"
              - match: { prefix: "/api/$svc/" }
                route: { cluster: cluster_$svc }
EOF
    port=$((port + 1))
  fi
done

cat <<EOF >> "$ROOT/infra/pc-gateway/envoy.yaml"
  clusters:
  - name: cluster_keycloak
    connect_timeout: 2s
    type: STRICT_DNS
    load_assignment:
      cluster_name: cluster_keycloak
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address: { address: 127.0.0.1, port_value: 8081 }
EOF

port=8101
for d in "$ROOT"/services/*; do
  if [ -d "$d" ]; then
    svc=$(basename "$d")
    cat <<EOF >> "$ROOT/infra/pc-gateway/envoy.yaml"
  - name: cluster_$svc
    connect_timeout: 2s
    type: STATIC
    load_assignment:
      cluster_name: cluster_$svc
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address: { address: 127.0.0.1, port_value: $port }
EOF
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

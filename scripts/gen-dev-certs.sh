#!/usr/bin/env bash
# Generate a self-signed TLS certificate for LOCAL DEVELOPMENT only.
# Output goes to ./certs/ which is gitignored — never commit private keys.
# Production certificates are issued by cert-manager (see infra/pc-gateway/k8s).
set -euo pipefail

CERT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/certs"
mkdir -p "$CERT_DIR"

if [[ -f "$CERT_DIR/server.key" && "${FORCE:-0}" != "1" ]]; then
  echo "certs already exist at $CERT_DIR (set FORCE=1 to regenerate)"
  exit 0
fi

openssl req -x509 -newkey rsa:2048 -nodes \
  -keyout "$CERT_DIR/server.key" \
  -out "$CERT_DIR/server.crt" \
  -days 365 \
  -subj "/CN=localhost" \
  -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"

chmod 600 "$CERT_DIR/server.key"
echo "Generated dev certs in $CERT_DIR (gitignored)."

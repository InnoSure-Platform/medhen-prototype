#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
export MEDHEN_DOCS_DIR="$ROOT/data/docs"
mkdir -p "$MEDHEN_DOCS_DIR"
cd "$ROOT"
if [[ ! -x bin/medhen-api ]]; then
  (cd platform && go build -o ../bin/medhen-api ./cmd/medhen-api)
fi
echo "Starting medhen-api on :8080"
exec ./bin/medhen-api

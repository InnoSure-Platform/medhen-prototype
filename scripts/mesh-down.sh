#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
for pidfile in "$ROOT/logs/"*.pid; do
  [[ -f "$pidfile" ]] || continue
  kill "$(cat "$pidfile")" 2>/dev/null || true
  rm -f "$pidfile"
done
echo "Mesh stopped"

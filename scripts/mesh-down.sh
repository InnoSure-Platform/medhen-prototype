#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
for pidfile in "$ROOT/logs/"*.pid; do
  [[ -f "$pidfile" ]] || continue
  kill "$(cat "$pidfile")" 2>/dev/null || true
  rm -f "$pidfile"
done

# Fallback to kill any orphaned processes that didn't have a pidfile
pkill -f "pc-(party-mgmt-svc|policy-svc|billing-svc|claims-svc|audit-svc|integration-svc|gateway)" 2>/dev/null || true

echo "Mesh stopped"

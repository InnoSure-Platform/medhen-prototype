#!/usr/bin/env bash
# Prove Telebirr integration — sandbox when credentials set, mock otherwise.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
INTEGRATION_URL="${INTEGRATION_URL:-http://localhost:8118}"

if [[ -z "${TELEBIRR_APP_ID:-}" || -z "${TELEBIRR_APP_SECRET:-}" ]]; then
  echo "TELEBIRR_APP_ID / TELEBIRR_APP_SECRET not set — running mock charge via integration-svc"
  MODE=mock
else
  echo "Telebirr sandbox credentials detected — live charge attempt"
  MODE=sandbox
fi

RESP=$(curl -sf -X POST "${INTEGRATION_URL}/internal/v1/telebirr/charge" \
  -H "Content-Type: application/json" \
  -d '{"phone":"+251911234567","amountMinor":10000,"reference":"prove-'$(uuidgen)'"}')
echo "$RESP"
RECEIPT=$(python3 -c "import sys,json; print(json.load(sys.stdin)['receiptId'])" <<<"$RESP")
echo "OK — receipt: $RECEIPT (mode: $MODE)"

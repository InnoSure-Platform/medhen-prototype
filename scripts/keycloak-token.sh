#!/usr/bin/env bash
# Obtain a Keycloak access token (resource-owner password grant — demo only).
set -euo pipefail
KC_URL="${KEYCLOAK_URL:-http://localhost:8081}"
REALM="${KEYCLOAK_REALM:-medhen}"
CLIENT="${KEYCLOAK_CLIENT:-pc-gateway}"
USER="${KEYCLOAK_USER:-demo-agent}"
PASS="${KEYCLOAK_PASS:-medhen-demo}"

RESP=$(curl -sf -X POST "${KC_URL}/realms/${REALM}/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password&client_id=${CLIENT}&username=${USER}&password=${PASS}") || {
  echo "Failed to get token from ${KC_URL}/realms/${REALM}" >&2
  exit 1
}
python3 -c "import sys,json; print(json.load(sys.stdin)['access_token'])" <<<"$RESP"

#!/usr/bin/env bash
#
# dev-bootstrap.sh — one-command local setup for the web + API.
#
# Idempotent. Safe to re-run. Assumes `make infra-up` has started the Keycloak
# container (medhen-keycloak-1). Sets the pc-web client secret in Keycloak, wires
# it (plus a NextAuth secret) into web/.env.local, and creates a demo login user.
#
# Usage:  bash scripts/dev-bootstrap.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
KC_CONTAINER="${KC_CONTAINER:-medhen-keycloak-1}"
KC_URL="${KEYCLOAK_URL:-http://localhost:8081}"
REALM="${KEYCLOAK_REALM:-medhen}"
ENV_FILE="$ROOT/web/.env.local"
DEMO_USER="${DEMO_USER:-agent}"
DEMO_PASS="${DEMO_PASS:-Passw0rd!}"

kc() { docker exec "$KC_CONTAINER" /opt/keycloak/bin/kcadm.sh "$@"; }

echo "› Checking Keycloak container ($KC_CONTAINER)…"
if ! docker ps --format '{{.Names}}' | grep -q "^${KC_CONTAINER}$"; then
  echo "  Keycloak is not running. Start infra first:  make infra-up" >&2
  exit 1
fi

echo "› Authenticating kcadm…"
kc config credentials --server "$KC_URL" --realm master --user admin --password admin >/dev/null

# Reuse an existing non-empty KEYCLOAK_SECRET from .env.local, else generate one.
existing="$(grep -E '^KEYCLOAK_SECRET=' "$ENV_FILE" 2>/dev/null | cut -d= -f2- || true)"
KC_WEB_SECRET="${existing:-$(openssl rand -hex 24)}"

echo "› Setting pc-web client secret in realm '$REALM'…"
CUUID="$(kc get clients -r "$REALM" -q clientId=pc-web --fields id --format csv --noquotes | tr -d '\r"')"
if [[ -z "$CUUID" ]]; then echo "  pc-web client not found in realm $REALM" >&2; exit 1; fi
kc update "clients/$CUUID" -r "$REALM" -s "secret=$KC_WEB_SECRET" >/dev/null

echo "› Writing web/.env.local…"
touch "$ENV_FILE"
set_env() { # key value
  if grep -qE "^$1=" "$ENV_FILE"; then
    sed -i '' "s|^$1=.*|$1=$2|" "$ENV_FILE"
  else
    printf '%s=%s\n' "$1" "$2" >> "$ENV_FILE"
  fi
}
naut="$(grep -E '^NEXTAUTH_SECRET=' "$ENV_FILE" 2>/dev/null | cut -d= -f2- || true)"
set_env KEYCLOAK_SECRET "$KC_WEB_SECRET"
set_env NEXTAUTH_SECRET "${naut:-$(openssl rand -base64 32)}"

echo "› Ensuring demo user '$DEMO_USER'…"
if [[ -z "$(kc get users -r "$REALM" -q username="$DEMO_USER" --fields id --format csv --noquotes | tr -d '\r"')" ]]; then
  kc create users -r "$REALM" -s username="$DEMO_USER" -s enabled=true -s firstName=Demo -s lastName=Agent >/dev/null
fi
kc set-password -r "$REALM" --username "$DEMO_USER" --new-password "$DEMO_PASS" >/dev/null
# Best-effort role grants (ignore if the realm role doesn't exist).
for role in staff claims admin agent; do
  kc add-roles -r "$REALM" --uusername "$DEMO_USER" --rolename "$role" >/dev/null 2>&1 || true
done

echo
echo "✓ Done."
echo "  Login:  $DEMO_USER / $DEMO_PASS"
echo "  Next:   cd web && npm run dev     # http://localhost:3000"
echo "  API:    export DATABASE_URL=postgres://medhen:medhen@localhost:5432/medhen?sslmode=disable && make api"

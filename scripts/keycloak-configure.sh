#!/usr/bin/env bash
# Post-import configuration for the local Keycloak realm.
#
# The pc-web confidential client's secret is deliberately NOT committed to the
# realm JSON. Run this after Keycloak has imported the `medhen` realm to set the
# client secret from the environment so the web app (KEYCLOAK_SECRET) and
# Keycloak agree on the same value.
#
# Required env:
#   KEYCLOAK_URL              (default http://localhost:8081)
#   KEYCLOAK_ADMIN            (default admin)
#   KEYCLOAK_ADMIN_PASSWORD   (default admin)
#   KEYCLOAK_PC_WEB_SECRET    (required — must equal the web app's KEYCLOAK_SECRET)
set -euo pipefail

KC_URL="${KEYCLOAK_URL:-http://localhost:8081}"
KC_ADMIN="${KEYCLOAK_ADMIN:-admin}"
KC_ADMIN_PW="${KEYCLOAK_ADMIN_PASSWORD:-admin}"
REALM="${KEYCLOAK_REALM:-medhen}"

: "${KEYCLOAK_PC_WEB_SECRET:?set KEYCLOAK_PC_WEB_SECRET (must match the web app KEYCLOAK_SECRET)}"

KCADM="${KCADM:-kcadm.sh}"

"$KCADM" config credentials --server "$KC_URL" --realm master \
  --user "$KC_ADMIN" --password "$KC_ADMIN_PW"

CLIENT_UUID="$("$KCADM" get clients -r "$REALM" -q clientId=pc-web --fields id --format csv --noquotes | head -n1)"
if [[ -z "$CLIENT_UUID" ]]; then
  echo "pc-web client not found in realm $REALM" >&2
  exit 1
fi

"$KCADM" update "clients/$CLIENT_UUID" -r "$REALM" \
  -s "secret=$KEYCLOAK_PC_WEB_SECRET"

echo "Set pc-web client secret from KEYCLOAK_PC_WEB_SECRET in realm $REALM."

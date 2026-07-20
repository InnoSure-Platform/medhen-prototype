#!/usr/bin/env bash
# Helper script to set production environment variables in Vercel.
# Usage: ./scripts/set-vercel-env.sh [environment] (default: production)

set -euo pipefail

ENV_TARGET="${1:-production}"
WEB_DIR="$(cd "$(dirname "$0")/../web" && pwd)"

cd "$WEB_DIR"

echo "Setting environment variables for Vercel ($ENV_TARGET)..."

# Ensure vercel CLI is installed
if ! command -v vercel &>/dev/null; then
  echo "Error: vercel CLI is not installed or not in PATH."
  echo "Run: npm install -g vercel"
  exit 1
fi

add_env() {
  local key="$1"
  local val="$2"
  echo "Adding $key..."
  printf "%s" "$val" | vercel env add "$key" "$ENV_TARGET" --force || true
}

NEXTAUTH_SECRET_VAL=$(openssl rand -base64 32)
KEYCLOAK_SECRET_VAL="supersecret-keycloak-client"
KEYCLOAK_ISSUER_VAL="${KEYCLOAK_ISSUER:-http://localhost:8081/realms/medhen}"
KEYCLOAK_CLIENT_VAL="${KEYCLOAK_CLIENT:-pc-web}"
NEXTAUTH_URL_VAL="${NEXTAUTH_URL:-http://localhost:3000}"
NEXT_PUBLIC_MEDHEN_API_VAL="${NEXT_PUBLIC_MEDHEN_API:-http://localhost:8080}"

add_env "NEXTAUTH_SECRET" "$NEXTAUTH_SECRET_VAL"
add_env "KEYCLOAK_SECRET" "$KEYCLOAK_SECRET_VAL"
add_env "KEYCLOAK_ISSUER" "$KEYCLOAK_ISSUER_VAL"
add_env "KEYCLOAK_CLIENT" "$KEYCLOAK_CLIENT_VAL"
add_env "NEXTAUTH_URL" "$NEXTAUTH_URL_VAL"
add_env "NEXT_PUBLIC_MEDHEN_API" "$NEXT_PUBLIC_MEDHEN_API_VAL"

echo "All environment variables successfully updated for Vercel ($ENV_TARGET)."

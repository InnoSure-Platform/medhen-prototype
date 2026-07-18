#!/usr/bin/env bash
# One-shot: purge previously-committed secrets from the ENTIRE git history.
#
# ⚠️  DESTRUCTIVE & IRREVERSIBLE for shared history. This rewrites every commit
#     hash and requires a coordinated force-push. All collaborators must
#     re-clone (or hard-reset) afterwards, and open PRs will need rebasing.
#
# Targets:
#   - certs/server.key, certs/server.crt   (committed TLS keypair)
#   - the literal Keycloak client secret string
#
# Prereqs: git-filter-repo (https://github.com/newren/git-filter-repo)
#   macOS:  brew install git-filter-repo
#   pip:    pipx install git-filter-repo   (or: pip install git-filter-repo)
#
# Usage:
#   1. Ensure a clean working tree and that you have rotated all secrets first.
#   2. Run this script.
#   3. Verify, then push:  git push --force-with-lease origin main
set -euo pipefail

if ! command -v git-filter-repo >/dev/null 2>&1; then
  echo "git-filter-repo not found. Install it first (see header)." >&2
  exit 1
fi

if [[ -n "$(git status --porcelain)" ]]; then
  echo "Working tree not clean. Commit or stash first." >&2
  exit 1
fi

# Safety: tag current HEAD so the pre-scrub state is recoverable locally.
BACKUP_TAG="backup/pre-secret-scrub-$(date +%Y%m%d%H%M%S)"
git tag "$BACKUP_TAG"
echo "Tagged current state as $BACKUP_TAG (local recovery point)."

# Remove the committed key/cert files from all of history.
git filter-repo --force \
  --invert-paths \
  --path certs/server.key \
  --path certs/server.crt

# Redact the known Keycloak client secret string wherever it appeared.
cat > /tmp/scrub-replacements.txt <<'EOF'
supersecret-keycloak-client==>REDACTED-ROTATED-SECRET
supersecret123==>REDACTED-ROTATED-SECRET
EOF
git filter-repo --force --replace-text /tmp/scrub-replacements.txt
rm -f /tmp/scrub-replacements.txt

echo
echo "History rewritten. Verify no matches remain:"
echo "  git log --all -p | grep -n 'supersecret' || echo clean"
echo
echo "filter-repo removes 'origin'. Re-add and force-push when ready:"
echo "  git remote add origin git@github.com:InnoSure-Platform/medhen-prototype.git"
echo "  git push --force-with-lease --all origin"
echo "  git push --force-with-lease --tags origin   # (optional)"
echo
echo "All collaborators must then re-clone or hard-reset to the rewritten history."

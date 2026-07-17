#!/usr/bin/env bash
#
# coverage-check.sh — enforce per-layer line-coverage thresholds.
#
# Reads a coverage profile produced by `go test -coverprofile`, computes
# coverage per architectural layer, and exits non-zero if any layer is below
# its floor. Designed to run in CI after `make test-cover`.
#
# Usage:
#   scripts/coverage-check.sh [coverage.out]
#
# Thresholds (line %) are intentionally layered: the domain is held to the
# highest bar; infrastructure to the lowest because some driver error branches
# are only reachable under fault injection.
set -euo pipefail

PROFILE="${1:-coverage.out}"

if [[ ! -f "$PROFILE" ]]; then
  echo "coverage-check: profile '$PROFILE' not found — run 'make test-cover' first" >&2
  exit 2
fi

# layer-path-prefix : minimum line coverage %
declare -a LAYERS=(
  "internal/domain:95"
  "internal/application:90"
  "internal/api:85"
  "internal/infrastructure:80"
)

# Build a func-level coverage report once.
FUNC_REPORT="$(go tool cover -func="$PROFILE")"

fail=0
printf '%-32s %8s %8s   %s\n' "LAYER" "COVERAGE" "FLOOR" "RESULT"
printf '%-32s %8s %8s   %s\n' "--------------------------------" "--------" "-----" "------"

for entry in "${LAYERS[@]}"; do
  prefix="${entry%%:*}"
  floor="${entry##*:}"

  # Average the per-function percentages for files under this layer prefix.
  # go tool cover -func emits fully-qualified paths
  # ("<module>/internal/domain/...:<line>: <func> <pct>%"), so match the layer
  # path as a substring rather than a line prefix.
  pct="$(awk -v p="/$prefix/" '
    index($0, p) > 0 {
      gsub(/%/, "", $NF); sum += $NF; n++
    }
    END { if (n > 0) printf "%.1f", sum / n; else print "n/a" }
  ' <<< "$FUNC_REPORT")"

  if [[ "$pct" == "n/a" ]]; then
    printf '%-32s %8s %8s   %s\n' "$prefix" "n/a" "$floor%" "SKIP (no files)"
    continue
  fi

  if awk -v a="$pct" -v b="$floor" 'BEGIN { exit !(a + 0 < b + 0) }'; then
    printf '%-32s %7s%% %7s%%   %s\n' "$prefix" "$pct" "$floor" "FAIL"
    fail=1
  else
    printf '%-32s %7s%% %7s%%   %s\n' "$prefix" "$pct" "$floor" "ok"
  fi
done

total="$(awk '/^total:/ { gsub(/%/, "", $NF); print $NF }' <<< "$FUNC_REPORT")"
echo
echo "total coverage: ${total:-unknown}%"

if [[ "$fail" -ne 0 ]]; then
  echo "coverage-check: FAILED — one or more layers below threshold" >&2
  exit 1
fi
echo "coverage-check: PASSED"

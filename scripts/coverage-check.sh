#!/usr/bin/env bash
#
# coverage-check.sh — enforce per-layer statement-coverage thresholds.
#
# Reads a coverage profile produced by `go test -coverprofile` and computes
# STATEMENT-weighted coverage per architectural layer (matching the file path),
# exiting non-zero if any layer is below its floor. Statement-weighting (rather
# than averaging per-function percentages) avoids trivial one-line functions
# skewing the result.
#
# Usage: scripts/coverage-check.sh [coverage.out]
set -euo pipefail

PROFILE="${1:-coverage.out}"

if [[ ! -f "$PROFILE" ]]; then
  echo "coverage-check: profile '$PROFILE' not found — run 'make test-cover' first" >&2
  exit 2
fi

# layer-regex : min-coverage% : label. The domain is held to the highest bar.
# Adapters/rest/cmd are reported by `go tool cover` but not gated here (their
# error branches are only reachable under fault injection).
declare -a LAYERS=(
  "/modules/[^/]+/domain/:85:domain"
  "/modules/[^/]+/app/:55:app"
  "/internal/platform/:55:platform"
)

fail=0
printf '%-16s %8s %8s   %s\n' "LAYER" "COVERAGE" "FLOOR" "RESULT"
printf '%-16s %8s %8s   %s\n' "----------------" "--------" "-----" "------"

for entry in "${LAYERS[@]}"; do
  label="${entry##*:}"
  rest="${entry%:*}"
  pat="${rest%:*}"
  floor="${rest##*:}"

  # Statement coverage for blocks whose file path matches the layer. Blocks are
  # deduped by key and their counts unioned (max) because a -coverpkg profile
  # repeats each block once per test binary that instrumented it.
  # Profile line format: <file>:<start>.<col>,<end>.<col> <numstmt> <count>
  pct="$(awk -v p="$pat" '
    NR == 1 && /^mode:/ { next }
    {
      file = $1; sub(/:[0-9].*/, "", file)
      if (file !~ p) next
      ns[$1] = $2
      if ($3 + 0 > cnt[$1]) cnt[$1] = $3
    }
    END {
      for (k in ns) { total += ns[k]; if (cnt[k] > 0) covered += ns[k] }
      if (total > 0) printf "%.1f", 100 * covered / total; else print "n/a"
    }
  ' "$PROFILE")"

  if [[ "$pct" == "n/a" ]]; then
    printf '%-16s %8s %8s   %s\n' "$label" "n/a" "$floor%" "SKIP (no files)"
    continue
  fi

  if awk -v a="$pct" -v b="$floor" 'BEGIN { exit !(a + 0 < b + 0) }'; then
    printf '%-16s %7s%% %7s%%   %s\n' "$label" "$pct" "$floor" "FAIL"
    fail=1
  else
    printf '%-16s %7s%% %7s%%   %s\n' "$label" "$pct" "$floor" "ok"
  fi
done

total="$(go tool cover -func="$PROFILE" | awk '/^total:/ { gsub(/%/, "", $NF); print $NF }')"
echo
echo "total coverage: ${total:-unknown}%"

if [[ "$fail" -ne 0 ]]; then
  echo "coverage-check: FAILED — one or more layers below threshold" >&2
  exit 1
fi
echo "coverage-check: PASSED"

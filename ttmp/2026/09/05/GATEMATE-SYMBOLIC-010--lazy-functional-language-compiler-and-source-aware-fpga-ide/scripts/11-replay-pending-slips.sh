#!/usr/bin/env bash
# These are delayed prints. Original phase-time failures remain in the diary.
set -euo pipefail
cd "$(dirname "$0")/.."
mode=${1:---list}
if [[ "$mode" != --list && "$mode" != --print ]]; then
  echo 'Usage: 11-replay-pending-slips.sh [--list|--print]' >&2
  exit 2
fi
old=../GATEMATE-SYMBOLIC-009--laboratory-4-lazy-graph-reducer-and-heap-inspector
for layout in "$old/reference/validation/report-done.yaml" reference/validation/{plan,d1-start,d1-done,d2-start,d2-done,d3-start,d3-done}.yaml; do
  test -f "$layout"
  stem=${layout%.yaml}
  if rg -l 'printed: true' "${stem}"*print.log >/dev/null 2>&1; then
    continue
  fi
  echo "Pending delayed print: $layout"
  if [[ "$mode" == --print ]]; then
    stamp=$(date -u +%Y%m%dT%H%M%SZ)
    almanach-render-service print-remote --layout "$layout" --output yaml > "${stem}-${stamp}-replay-print.log" 2>&1
  fi
done

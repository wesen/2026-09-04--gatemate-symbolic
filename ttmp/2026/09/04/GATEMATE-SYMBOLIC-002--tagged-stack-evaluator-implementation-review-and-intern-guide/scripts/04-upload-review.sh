#!/usr/bin/env bash
set -euo pipefail
ticket_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
remarquee upload bundle \
  "$ticket_dir/design-doc/01-tagged-stack-evaluator-analysis-design-and-implementation-review.md" \
  "$ticket_dir/reference/01-investigation-diary.md" \
  --name 'GATEMATE-SYMBOLIC-002 Implementation Review' \
  --remote-dir '/ai/2026/09/04/GATEMATE-SYMBOLIC-002' \
  --toc-depth 2 --non-interactive "$@"

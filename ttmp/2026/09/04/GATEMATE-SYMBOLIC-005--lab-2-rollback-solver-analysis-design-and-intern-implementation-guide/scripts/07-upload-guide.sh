#!/usr/bin/env bash
set -euo pipefail
ticket_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
python3 "$ticket_dir/scripts/06-prepare-delivery.py"
remarquee upload bundle /tmp/gatemate005-delivery/guide.md \
  --name 'GATEMATE-SYMBOLIC-005 Lab 2 Design and Intern Guide' \
  --remote-dir '/ai/2026/09/04/GATEMATE-SYMBOLIC-005' \
  --toc-depth 2 --non-interactive "$@"

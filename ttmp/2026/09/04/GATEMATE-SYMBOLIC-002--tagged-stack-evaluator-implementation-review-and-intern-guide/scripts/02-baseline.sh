#!/usr/bin/env bash
set -euo pipefail
source /home/manuel/fpga/oss-cad-suite/environment
ticket_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
repo_dir="$(git -C "$ticket_dir" rev-parse --show-toplevel)"
cd "$repo_dir/symbolic_eval"
make versions > "$ticket_dir/reference/validation/tool-versions.log" 2>&1
make test > "$ticket_dir/reference/validation/baseline-tests.log" 2>&1
python3 -m pytest sim/ --collect-only -q > "$ticket_dir/reference/validation/test-inventory.log" 2>&1
python3 "$ticket_dir/scripts/01-review-probes.py" > "$ticket_dir/reference/validation/probes.log" 2>&1

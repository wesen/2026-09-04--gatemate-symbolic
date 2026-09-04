#!/usr/bin/env bash
set -euo pipefail
ticket_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source /home/manuel/fpga/oss-cad-suite/environment
python3 "$ticket_dir/scripts/10-hardware.py" > "$ticket_dir/reference/validation/P5-hardware.log" 2>&1

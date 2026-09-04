#!/usr/bin/env bash
set -euo pipefail
source /home/manuel/fpga/oss-cad-suite/environment
ticket="$(cd "$(dirname "$0")/.." && pwd)"
python3 "$ticket/scripts/05-hardware.py" > "$ticket/reference/validation/P5-hardware.log" 2>&1

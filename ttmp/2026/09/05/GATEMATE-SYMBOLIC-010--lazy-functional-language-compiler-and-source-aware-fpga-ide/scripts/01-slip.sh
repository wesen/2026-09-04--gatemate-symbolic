#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p reference/validation
kind=${1:?}
shift
python3 /home/manuel/.pi/agent/skills/brutalist-work-slip/scripts/work_slip.py "$kind" --task GATEMATE-010 "$@"

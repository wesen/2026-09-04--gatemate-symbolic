#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
kind=${1:?plan or status}
shift
python3 /home/manuel/.pi/agent/skills/brutalist-work-slip/scripts/work_slip.py "$kind" --task GATEMATE-008 "$@"

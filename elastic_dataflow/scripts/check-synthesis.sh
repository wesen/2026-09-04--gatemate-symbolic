#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
source /home/manuel/fpga/oss-cad-suite/environment
mkdir -p build
yosys -ql build/core-synthesis.log scripts/synth-core.ys

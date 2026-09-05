#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../../../../../.."
source /home/manuel/fpga/oss-cad-suite/environment
make -C lazy_reducer bit > ttmp/2026/09/05/GATEMATE-SYMBOLIC-009--laboratory-4-lazy-graph-reducer-and-heap-inspector/reference/validation/p3-board-build.log 2>&1

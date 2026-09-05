#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../../../../../.."
elastic_dataflow/scripts/build-board.sh > ttmp/2026/09/04/GATEMATE-SYMBOLIC-008--programmable-dataflow-workbench-compiler-and-physical-debugger/reference/validation/p4-board-build.log 2>&1

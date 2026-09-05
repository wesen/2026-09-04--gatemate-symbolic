#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../../../../../.."
source /home/manuel/fpga/oss-cad-suite/environment
ticket=ttmp/2026/09/04/GATEMATE-SYMBOLIC-008--programmable-dataflow-workbench-compiler-and-physical-debugger
elastic_dataflow/scripts/test-stress.sh
GOCACHE=/tmp/gatemate008-go-cache go test ./pkg/dataflow -run TestRTLDifferential -rtl-log-dir ../../elastic_dataflow/build/stress -v > "$ticket/reference/validation/p6-differential.log" 2>&1

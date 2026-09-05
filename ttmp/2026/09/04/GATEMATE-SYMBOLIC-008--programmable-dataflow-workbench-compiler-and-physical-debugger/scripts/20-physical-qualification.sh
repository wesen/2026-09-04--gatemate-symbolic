#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../../../../../.."
source /home/manuel/fpga/oss-cad-suite/environment
ticket=ttmp/2026/09/04/GATEMATE-SYMBOLIC-008--programmable-dataflow-workbench-compiler-and-physical-debugger
if rg -q '^ERROR:' elastic_dataflow/build/nextpnr.log; then
  echo 'Refusing to program a timing-failed build' >&2
  exit 1
fi
make -C elastic_dataflow load > "$ticket/reference/validation/p6-program-board.log" 2>&1
GOCACHE=/tmp/gatemate008-go-cache go test ./pkg/dataflow -run 'TestPhysical(Qualification|Workbench)$' -count=1 -v -physical-device /dev/ttyACM0 -physical-wire-log "/home/manuel/code/wesen/2026-09-04--gatemate-symbolic/$ticket/reference/validation/p6-physical-wire.log" > "$ticket/reference/validation/p6-physical-tests.log" 2>&1

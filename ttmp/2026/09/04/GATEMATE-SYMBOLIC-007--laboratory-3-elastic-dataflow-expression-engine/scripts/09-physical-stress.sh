#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../../../../../.."
ticket="$(pwd)"/ttmp/2026/09/04/GATEMATE-SYMBOLIC-007--laboratory-3-elastic-dataflow-expression-engine
go test ./pkg/dataflow -run TestPhysicalQualification -count=1 -v -physical-device=/dev/ttyACM0 -physical-wire-log="$ticket/reference/validation/P5-physical-wire.log" > "$ticket/reference/validation/P5-physical-stress.log" 2>&1

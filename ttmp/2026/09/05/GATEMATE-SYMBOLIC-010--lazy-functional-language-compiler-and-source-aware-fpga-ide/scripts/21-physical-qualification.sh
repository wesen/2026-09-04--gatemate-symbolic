#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../../../../../.."
ticket=ttmp/2026/09/05/GATEMATE-SYMBOLIC-010--lazy-functional-language-compiler-and-source-aware-fpga-ide
GOCACHE=/tmp/gatemate009-go-cache go test ./pkg/lazylang/serial -run TestPhysicalPrograms -count=1 -v -lfl-device /dev/ttyACM0 -lfl-capture "$PWD/$ticket/reference/validation/i4-physical-uart.log"

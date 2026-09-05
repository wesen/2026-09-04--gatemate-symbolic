#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../../../../../.."
source /home/manuel/fpga/oss-cad-suite/environment
ticket=ttmp/2026/09/05/GATEMATE-SYMBOLIC-009--laboratory-4-lazy-graph-reducer-and-heap-inspector
if rg -q '^ERROR:' lazy_reducer/build/nextpnr.log;then echo 'Refusing timing-failed image' >&2;exit 1;fi
if [[ ! lazy_reducer/build/top.bit -nt lazy_reducer/build/nextpnr.log ]];then echo 'Refusing stale bitstream' >&2;exit 1;fi
make -C lazy_reducer load > "$ticket/reference/validation/p5-program-board.log" 2>&1
GOCACHE=/tmp/gatemate009-go-cache go test ./pkg/lazy -run TestPhysicalLazyQualification -count=1 -v -lazy-physical-device /dev/ttyACM0 -lazy-wire-log "$PWD/$ticket/reference/validation/p5-physical-wire.log" > "$ticket/reference/validation/p5-physical-tests.log" 2>&1

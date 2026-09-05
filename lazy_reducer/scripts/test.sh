#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p build
GOCACHE=/tmp/gatemate009-go-cache go run scripts/generate_cases.go
iverilog -g2012 -I build -s lazy_core_tb -o build/core-test ../symbolic_eval/rtl/sync_sdp_ram.sv rtl/lazy_core.sv tb/lazy_core_tb.sv
vvp build/core-test
iverilog -g2012 -s lazy_link_tb -o build/link-test ../symbolic_eval/rtl/sync_sdp_ram.sv ../symbolic_eval/rtl/uart_tx.sv ../graph_microscope/rtl/uart_rx.sv rtl/lazy_core.sv rtl/lazy_link.sv tb/lazy_link_tb.sv
vvp build/link-test

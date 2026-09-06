#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
source /home/manuel/fpga/oss-cad-suite/environment
mkdir -p build
GOCACHE=/tmp/gatemate009-go-cache go run ./scripts/generate_cases.go
iverilog -g2012 -I build -s lfl_core_tb -o build/lfl_core.vvp ../symbolic_eval/rtl/sync_sdp_ram.sv rtl/lfl_core.sv tb/lfl_core_tb.sv
vvp build/lfl_core.vvp
GOCACHE=/tmp/gatemate009-go-cache go run ./scripts/generate_link_cases.go
iverilog -g2012 -I build -s lfl_link_tb -o build/lfl_link.vvp ../symbolic_eval/rtl/sync_sdp_ram.sv ../symbolic_eval/rtl/uart_tx.sv ../graph_microscope/rtl/uart_rx.sv rtl/lfl_core.sv rtl/lfl_link.sv tb/lfl_link_tb.sv
vvp build/lfl_link.vvp

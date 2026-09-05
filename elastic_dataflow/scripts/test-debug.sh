#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p build
iverilog -g2012 -s tb_debug -o build/debug.vvp rtl/dataflow_pkg.sv ../symbolic_eval/rtl/sync_sdp_ram.sv rtl/df_fifo.sv rtl/df_unit.sv rtl/dataflow_core.sv sim/tb_debug.sv
vvp build/debug.vvp

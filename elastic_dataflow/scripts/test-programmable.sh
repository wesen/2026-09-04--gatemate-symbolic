#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p build
iverilog -g2012 -s tb_programmable -o build/programmable.vvp rtl/dataflow_pkg.sv ../symbolic_eval/rtl/sync_sdp_ram.sv rtl/df_fifo.sv rtl/df_unit.sv rtl/dataflow_core.sv sim/tb_programmable.sv
vvp build/programmable.vvp

#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p build
iverilog -g2012 -s tb_dataflow_link -o build/link.vvp rtl/dataflow_pkg.sv ../symbolic_eval/rtl/sync_sdp_ram.sv ../symbolic_eval/rtl/uart_tx.sv ../graph_microscope/rtl/uart_rx.sv rtl/df_fifo.sv rtl/df_unit.sv rtl/dataflow_core.sv rtl/dataflow_link.sv sim/tb_dataflow_link.sv
vvp build/link.vvp

#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p build/stress
for depth in 1 2 8; do
 for latency in 1 2 4 8; do
  iverilog -g2012 -s tb_dataflow_stress -Ptb_dataflow_stress.DEPTH="$depth" -Ptb_dataflow_stress.LATENCY="$latency" -o build/stress.vvp rtl/dataflow_pkg.sv ../symbolic_eval/rtl/sync_sdp_ram.sv rtl/df_fifo.sv rtl/df_unit.sv rtl/dataflow_core.sv sim/tb_dataflow_stress.sv
  vvp build/stress.vvp > "build/stress/depth-$depth-latency-$latency.log"
 done
done

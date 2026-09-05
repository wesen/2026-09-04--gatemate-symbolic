#!/usr/bin/env bash
# One-time cleanup of this task's completed tmux jobs. Keep dataflow-ide-fpga.
set -euo pipefail
for job in dataflow-board-balanced dataflow-board-build dataflow-final-checks dataflow-final-upload dataflow-physical dataflow-physical-stress dataflow-physical-stress-path dataflow-synthesis-fixed; do
 tmux kill-session -t "$job"
done

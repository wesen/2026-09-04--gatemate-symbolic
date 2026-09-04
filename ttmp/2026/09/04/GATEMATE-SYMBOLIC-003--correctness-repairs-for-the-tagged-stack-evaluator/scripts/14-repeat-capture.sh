#!/usr/bin/env bash
set -euo pipefail
ticket_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
repo_dir="$(git -C "$ticket_dir" rev-parse --show-toplevel)"
source /home/manuel/fpga/oss-cad-suite/environment
cd "$repo_dir/symbolic_eval"
stty -F /dev/ttyACM0 115200 raw -echo
timeout 8 cat /dev/ttyACM0 > "$ticket_dir/reference/validation/P5-countdown-repeat.bin" &
reader_pid=$!
sleep 0.3
openFPGALoader -b olimex_gatemateevb build/top.bit > "$ticket_dir/reference/validation/P5-countdown-repeat-load.log" 2>&1
wait "$reader_pid" || test "$?" = 124

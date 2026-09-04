#!/usr/bin/env bash
set -euo pipefail
ticket_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
repo_dir="$(git -C "$ticket_dir" rev-parse --show-toplevel)"
trap 'printf "%s\n" "$?" > "$ticket_dir/reference/validation/P5-build.exit"' EXIT
source /home/manuel/fpga/oss-cad-suite/environment
cd "$repo_dir/symbolic_eval"
git rev-parse HEAD > "$ticket_dir/reference/validation/P5-build-commit.txt"
make versions > "$ticket_dir/reference/validation/P5-versions.log" 2>&1
timeout 180 make bit PROG=fib > "$ticket_dir/reference/validation/P5-build.log" 2>&1
cp build/yosys.log "$ticket_dir/reference/validation/P5-yosys.log"
cp build/nextpnr.log "$ticket_dir/reference/validation/P5-nextpnr.log"
sha256sum build/top.bit build/prog.hex > "$ticket_dir/reference/validation/P5-image-sha256.txt"
python3 scripts/check_isa.py > "$ticket_dir/reference/validation/P5-isa.log"

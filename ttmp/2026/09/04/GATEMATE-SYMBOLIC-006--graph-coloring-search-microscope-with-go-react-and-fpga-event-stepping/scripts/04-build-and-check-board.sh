#!/usr/bin/env bash
set -euo pipefail
source /home/manuel/fpga/oss-cad-suite/environment
ticket="$(cd "$(dirname "$0")/.." && pwd)"
repo="$(git -C "$ticket" rev-parse --show-toplevel)"
out="$ticket/reference/validation"
trap 'printf "%s\n" "$?" > "$out/P3-hardware.exit"' EXIT
git -C "$repo" rev-parse HEAD > "$out/P3-build-commit.txt"
timeout 300 make -C "$repo/graph_microscope" bit > "$out/P3-build.log" 2>&1
cp "$repo/graph_microscope/build/yosys.log" "$out/P3-yosys.log"
cp "$repo/graph_microscope/build/nextpnr.log" "$out/P3-nextpnr.log"
sha256sum "$repo/graph_microscope/build/top.bit" > "$out/P3-image.sha256"
timeout 45 openFPGALoader -b olimex_gatemateevb "$repo/graph_microscope/build/top.bit" > "$out/P3-load.log" 2>&1
cd "$repo"
go test ./pkg/microscope -run TestPhysicalGraphEvents -count=1 -v -args -hardware-device /dev/ttyACM0 -hardware-out "$out" > "$out/P3-hardware.log" 2>&1

#!/usr/bin/env bash
set -euo pipefail
ticket_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
defuddle parse 'https://yosyshq.readthedocs.io/projects/yosys/en/v0.50/cmd/synth_gatemate.html' --md -o "$ticket_dir/sources/yosys-synth-gatemate.md"
defuddle parse 'https://www.olimex.com/Products/FPGA/GateMate/GateMateA1-EVB/open-source-hardware' --md -o "$ticket_dir/sources/olimex-gatematea1-evb.md"

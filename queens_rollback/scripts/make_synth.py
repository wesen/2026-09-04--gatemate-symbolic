#!/usr/bin/env python3
"""Emit an explicit, reproducible Yosys script with validated integer parameters."""
import argparse
from pathlib import Path
p=argparse.ArgumentParser()
p.add_argument('--trail',type=int,choices=[0,1],required=True)
p.add_argument('--first-only',type=int,choices=[0,1],required=True)
p.add_argument('--trail-capacity',type=int,choices=range(65),required=True)
p.add_argument('--choice-capacity',type=int,choices=range(9),required=True)
a=p.parse_args()
text='read_verilog -defer -sv rtl/queens_types_pkg.sv rtl/queens_core.sv rtl/queens_result_printer.sv rtl/queens_top.sv ../symbolic_eval/rtl/sync_sdp_ram.sv ../symbolic_eval/rtl/reset_sync.sv ../symbolic_eval/rtl/uart_tx.sv\n'
for name,value in [('USE_TRAIL',a.trail),('FIRST_ONLY',a.first_only),('TRAIL_CAPACITY',a.trail_capacity),('CHOICE_CAPACITY',a.choice_capacity)]:
    text+=f'chparam -set {name} {value} queens_top\n'
text+='synth_gatemate -top queens_top -luttree -nomx8\nstat\nwrite_json build/top.json\n'
Path('build').mkdir(exist_ok=True)
Path('build/synth.ys').write_text(text)

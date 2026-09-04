#!/usr/bin/env python3
from pathlib import Path
import subprocess
TICKET=Path(__file__).resolve().parents[1]
PROJECT=next(p for p in TICKET.parents if (p/'symbolic_eval').is_dir())/'symbolic_eval'
OUT=TICKET/'reference/validation'
base=(PROJECT/'sim/tb_top.sv').read_text()
for mode in ['production','mapped']:
    bench=base.replace('.DEEP_DEPTH(30)', '.DEEP_DEPTH(512)')
    if mode=='mapped':
        start=bench.index('  top #(');end=bench.index('    .clk_10m',start)
        bench=bench[:start]+'  top dut (\n'+bench[end:]
        start=bench.index('    for (i = 0; i < 1024;');end=bench.index('    void\x27($value$plusargs("restart_cycle',start)
        bench=bench[:start]+bench[end:]
    path=TICKET/f'scripts/tb_{mode}.sv';path.write_text(bench)
    if mode=='production':
        files=['rtl/symbolic_types_pkg.sv','rtl/sync_sdp_ram.sv','rtl/stack_core_bram.sv','rtl/program_rom.sv','rtl/rv_reg.sv','rtl/reset_sync.sv','rtl/uart_tx.sv','rtl/top.sv']
    else:
        files=['build/mapped.v','/home/manuel/fpga/oss-cad-suite/share/yosys/gatemate/cells_sim.v']
    with (OUT/f'P5-{mode}-compile.log').open('w') as log:
        subprocess.run(['iverilog','-g2012','-s','tb_top','-o',f'build/{mode}.vvp',*files,'sim/CC_USR_RSTN.sv',str(path)],cwd=PROJECT,stdout=log,stderr=subprocess.STDOUT,check=True)
    with (OUT/f'P5-{mode}-simulation.log').open('w') as log:
        subprocess.run(['vvp',f'build/{mode}.vvp','+rom=build/countdown.hex'],cwd=PROJECT,stdout=log,stderr=subprocess.STDOUT,check=True,timeout=90)
    output=(OUT/f'P5-{mode}-simulation.log').read_text()
    raw=bytes(int(line.split()[1],16) for line in output.splitlines() if line.startswith('UARTBYTE '))
    print(mode,repr(raw),flush=True)

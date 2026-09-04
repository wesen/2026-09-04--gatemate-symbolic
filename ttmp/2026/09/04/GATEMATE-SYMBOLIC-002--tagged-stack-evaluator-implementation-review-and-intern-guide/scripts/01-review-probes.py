#!/usr/bin/env python3
"""Reproduce review findings without modifying production sources; run in CAD environment."""
from pathlib import Path
import difflib
import json
import subprocess
import sys

TICKET = Path(__file__).resolve().parents[1]
ROOT = next(p for p in TICKET.parents if (p / 'symbolic_eval').is_dir())
PROJECT = ROOT / 'symbolic_eval'
OUT = TICKET / 'reference' / 'validation'
OUT.mkdir(parents=True, exist_ok=True)
sys.path[:0] = [str(PROJECT / 'tools')]
from asm20 import assemble
from opcodes import encode
from stack_model import run_program

CASES = {
    'empty-dup': 'DUP\nHALT\n',
    'emit-jmp': 'PUSH_S15 7\nEMIT\nJMP end\nend: HALT\n',
    'emit-swap': 'PUSH_S15 1\nPUSH_S15 2\nPUSH_S15 3\nEMIT\nSWAP\nEMIT\nEMIT\nHALT\n',
    'deep-fault-tags': 'PUSH_TRUE\nPUSH_S15 2\nPUSH_S15 3\nDROP\nWORD 0xF8000\n',
    'call-ret-timing': 'CALL sub\nHALT\nsub: RET\n',
}
programs = {name: assemble(src)[0] for name, src in CASES.items()}
programs['pc-rom-boundary'] = [encode(0x0B, 1023)] + [0] * 1022 + [encode(0, 7)]
results = []
for core, depth_param in [('stack_core', 'STACK_DEPTH'), ('stack_core_bram', 'DEEP_DEPTH')]:
    bench_name = f'tb_{core}'
    bench = (PROJECT / 'sim' / f'{bench_name}.sv').read_text()
    # Separate cycle observation exposes return depth between EXECUTE and COMMIT.
    bench = bench.replace('endmodule', '''always @(posedge clk) begin
      #2;
      if (rst_n && cycle < 35)
        $display("CYCLE %0d PC %0d RDEPTH %0d TRACE %0d", cycle, pc_o, rdepth_o, trace_valid);
    end
endmodule''')
    bench_path = TICKET / 'scripts' / f'probe-{bench_name}.sv'
    bench_path.write_text(bench)
    vvp = OUT / f'{bench_name}.vvp'
    rtl = [PROJECT / 'rtl' / 'symbolic_types_pkg.sv']
    if core.endswith('bram'):
        rtl.append(PROJECT / 'rtl' / 'sync_sdp_ram.sv')
    rtl.append(PROJECT / 'rtl' / f'{core}.sv')
    compiled = subprocess.run(['iverilog', '-g2012', '-s', bench_name, '-o', str(vvp),
        *map(str, rtl), str(bench_path), f'-P{bench_name}.{depth_param}={30 if core.endswith("bram") else 32}'],
        capture_output=True, text=True, check=True)
    (OUT / f'{core}-compile.log').write_text(compiled.stdout + compiled.stderr)
    for name, words in programs.items():
        hex_path = OUT / f'{name}.hex'
        hex_path.write_text(''.join(f'{w:05x}\n' for w in words + [0] * (1024 - len(words))))
        result = subprocess.run(['vvp', str(vvp), f'+rom={hex_path}', '+max_cycles=10000'],
                                capture_output=True, text=True, check=True, timeout=60)
        (OUT / f'{core}-{name}.log').write_text(result.stdout + result.stderr)
        model = run_program(words, total_depth=32)
        expected = [r.line() for r in model.trace] + [model.final_line()]
        actual = [line for line in result.stdout.splitlines() if line.startswith(('TRACE', 'FINAL'))]
        diff = '\n'.join(difflib.unified_diff(expected, actual, 'model', core, lineterm=''))
        (OUT / f'{core}-{name}.diff').write_text(diff + '\n')
        results.append({'case': name, 'core': core, 'match': expected == actual,
                        'model_final': expected[-1], 'rtl_final': actual[-1],
                        'bench_pass': 'TB_PASS' in result.stdout})
        print(json.dumps(results[-1]))
    vvp.unlink()
for name, source, depth in [
    ('malformed-line', 'PUSH_S15 1 2\nHALT\n', 1024),
    ('word-capacity', 'WORD 0\nWORD 0\nWORD 0\n', 2),
]:
    try:
        words, _, listing = assemble(source, rom_depth=depth)
        result = {'case': name, 'accepted': True, 'words': words, 'rom_depth': depth, 'listing': listing}
    except ValueError as exc:
        result = {'case': name, 'accepted': False, 'error': str(exc)}
    results.append(result)
    print(json.dumps(result))
(OUT / 'probe-results.json').write_text(json.dumps(results, indent=2) + '\n')

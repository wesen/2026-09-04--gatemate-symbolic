#!/usr/bin/env python3
from pathlib import Path
ROOT=next(p for p in Path(__file__).resolve().parents if (p/'symbolic_eval').is_dir())/'symbolic_eval'
p=ROOT/'sim/test_bram.py';s=p.read_text();start=s.index('PUSH_OPS =');end=s.index('@pytest.mark.parametrize("seed", range(24))',start)
s=s[:start]+'from program_generation import gen_program\n\n\n'+s[end:]
s=s.replace('gen_program(rng, length=120, total_depth=total_depth)', 'gen_program(rng, length=120, total_depth=total_depth,\n                               illegal_p=0 if seed % 2 == 0 else 0.06)')
s=s.replace('gen_program(rng, length=120, total_depth=8)', 'gen_program(rng, length=120, total_depth=8,\n                           illegal_p=0 if seed % 2 == 0 else 0.06)')
p.write_text(s)

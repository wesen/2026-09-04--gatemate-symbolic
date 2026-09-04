#!/usr/bin/env python3
"""Print and archive the overall plan or a phase boundary slip."""
import argparse
from pathlib import Path
import subprocess
import sys
TICKET = Path(__file__).resolve().parents[1]
PHASES = {
    'P1': ('Instruction and assembler guards', 'P2 commitment and fault context'),
    'P2': ('Atomic retirement and fault context', 'P3 ROM boundary semantics'),
    'P3': ('ROM boundary and return addresses', 'P4 full state verification'),
    'P4': ('Full state and generator coverage', 'P5 build and hardware validation'),
    'P5': ('Build docs and hardware validation', 'review completed repairs'),
}
p = argparse.ArgumentParser()
p.add_argument('phase', choices=['PLAN', *PHASES])
p.add_argument('state', choices=['start', 'done'])
p.add_argument('--commit')
p.add_argument('--fact', action='append', default=[])
a = p.parse_args()
mode = 'status' if a.state == 'done' else 'plan'
command = [sys.executable, '/home/manuel/.pi/agent/skills/brutalist-work-slip/scripts/work_slip.py',
           mode, '--task', 'GATEMATE-003', '--label', a.phase + ' ' + a.state.upper(),
           '--out', str(TICKET / 'reference/slips' / f'{a.phase}-{a.state}.yaml')]
if a.phase == 'PLAN':
    command += ['--title', 'Tagged evaluator correctness repairs', '--next', 'P1 instruction and assembler guards']
    for phase, (title, _) in PHASES.items():
        command += ['--phase', phase + ' ' + title]
else:
    title, next_phase = PHASES[a.phase]
    command += ['--title', title, '--next', next_phase if a.state == 'done' else 'execute ' + a.phase]
    for item in ({
        'P1': ['empty DUP and post EMIT regressions', 'reject malformed and oversized assembly'],
        'P2': ['commit return depth atomically', 'fetch true deep operand fault tags'],
        'P3': ['preserve one past ROM addresses', 'match model fetch faults and CALL RET'],
        'P4': ['compare all live stack values', 'legal generation and coverage checks'],
        'P5': ['run tests synth route and pack', 'update docs and capture board output'],
    }[a.phase]):
        command += ['--did' if a.state == 'done' else '--phase', item]
if a.commit:
    # No GitHub remote assumption: archive the local commit as a fact.
    command += ['--fact', 'COMMIT=' + a.commit]
for fact in a.fact:
    command += ['--fact', fact]
r = subprocess.run(command, text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
(TICKET / 'reference/slips' / f'{a.phase}-{a.state}.log').write_text(r.stdout)
print(r.stdout, end='')
sys.exit(r.returncode)

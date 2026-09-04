#!/usr/bin/env python3
"""Replay and preserve exact architectural examples for the CPU article."""
from pathlib import Path
import json
import sys

ticket = Path(__file__).resolve().parents[1]
project = next(p for p in ticket.parents if (p / 'symbolic_eval').is_dir()) / 'symbolic_eval'
sys.path.insert(0, str(project / 'tools'))
from asm20 import assemble
from stack_model import Machine, Value40

def snapshot(machine):
    return {'pc': machine.pc, 'stack': [f'{v.word():010x}' for v in machine.stack],
            'rstack': machine.rstack.copy(), 'depth': machine.depth}

results = {}
def capture(name, source, **kwargs):
    words, symbols, _ = assemble(source, kwargs.get('rom_depth', 1024))
    m = Machine(program=words, **kwargs)
    rows = []
    max_depth = max_returns = 0
    while not m.halted and m.fault is None:
        assert len(rows) < 10000
        before = snapshot(m)
        rec = m.step()
        rows.append({'before': before, 'trace': rec.line(), 'after': snapshot(m)})
        max_depth = max(max_depth, m.depth)
        max_returns = max(max_returns, len(m.rstack))
    result = {'source': source, 'symbols': symbols, 'words': [f'{w:05x}' for w in words],
              'rows': rows, 'final': m.final_line(), 'events': len(rows),
              'max_data_depth': max_depth, 'max_return_depth': max_returns,
              'uart': ''.join(f'T{v.tag:X}:{v.payload:08X}\r\n' for v in m.output)}
    results[name] = result
    return m

for name in ['arith', 'sq', 'countdown', 'typefault', 'fib']:
    m = capture(name, (project / f'programs/{name}.asm').read_text(), total_depth=514)
    if name != 'typefault':
        assert m.halted and m.fault is None
capture('fib3', (project/'programs/fib.asm').read_text().replace('PUSH_S15 10', 'PUSH_S15 3'), total_depth=514)
capture('overflow', 'PUSH_S15 16383\nDUP\nMUL\nDUP\nADD\nDUP\nADD\nDUP\nADD\nDUP\nADD\nHALT')
capture('noncanonical', 'JZ 1\nHALT', stack=[Value40(1, 2)])
capture('flag_eq', 'EQ\nHALT', stack=[Value40(0, 7, 0), Value40(0, 7, 1)])
capture('rom_escape', 'JMP 6\nRET\nHALT\nHALT\nHALT\nHALT\nCALL 1', rom_depth=7)
assert results['overflow']['final'].split()[2] == 'ARITH_OVERFLOW'
assert results['rom_escape']['rows'][-1]['trace'].split()[4] == 'FETCH'
assert results['arith']['uart'] == 'T1:00000001\r\n'
assert results['sq']['uart'] == 'T0:00000019\r\n'
assert results['fib']['uart'] == 'T0:00000037\r\n'
out = ticket / 'reference/validation/report-examples.json'
out.write_text(json.dumps(results, indent=2) + '\n')
for name, r in results.items():
    print(name, json.dumps({k: r[k] for k in ['symbols','events','max_data_depth','max_return_depth','final','uart']}))
    if name in ['arith','sq','overflow','rom_escape']:
        for row in r['rows']:
            print(row['trace'], row['after']['stack'], row['after']['rstack'])

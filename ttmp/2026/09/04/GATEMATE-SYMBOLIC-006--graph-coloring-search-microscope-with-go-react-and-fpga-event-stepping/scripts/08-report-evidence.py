#!/usr/bin/env python3
"""Decode archived physical UART records independently for the technical report.

This script never opens the device. Reset runs are retained separately; tables
and totals refer to the first complete run in each archived capture.
"""
from collections import Counter
import hashlib
import itertools
import json
from pathlib import Path

T = Path(__file__).resolve().parents[1]
OUT = T / 'reference/validation'
NAMES = ['', 'CREATE', 'UPDATE', 'WRITE', 'PROPAGATED', 'CONTRADICTION',
         'RESTORE', 'RESTORED', 'POP', 'OUTPUT', 'COMPLETE', 'FAULT']
report = {}
for name in ['triangle', 'unsatisfiable', 'path', 'first', 'root']:
    path = OUT / f'hardware-{name}-wire.log'
    lines = path.read_text().splitlines()
    load = next(line[3:] for line in lines if line.startswith('> L'))
    config = bytes.fromhex(load)
    assert len(config) == 12 and __import__('functools').reduce(int.__xor__, config) == 0
    n, k, cut = config[:3]
    edges = [(u, v) for u in range(n) for v in range(u + 1, n) if config[3 + u] & (1 << v)]
    records = []
    for line in lines:
        if not line.startswith('< E'):
            continue
        p = bytes.fromhex(line[3:])
        assert len(p) == 34 and __import__('functools').reduce(int.__xor__, p) == 0
        seq, kind = int.from_bytes(p[:4]), p[4]
        assert seq == len(records) + 1
        cp, trail = int.from_bytes(p[25:30]), int.from_bytes(p[30:33])
        e = dict(sequence=seq, name=NAMES[kind], domains=list(reversed(p[5:13]))[:n],
                 propagated=f'{p[13]:02X}', choiceTop=p[14], trailTop=p[15], base=p[16],
                 count=int.from_bytes(p[17:21]), result=int.from_bytes(p[21:24]), fault=p[24])
        if kind in [1, 2]:
            e['choice'] = dict(vertex=(cp >> 37) & 7, remaining=(cp >> 29) & 255,
                               mark=(cp >> 22) & 127, propagated=(cp >> 14) & 255)
        if kind in [3, 6]:
            e['trail'] = dict(vertex=(trail >> 17) & 7, old=(trail >> 9) & 255, level=(trail >> 4) & 31)
        if kind == 9:
            e['coloring'] = [(e['result'] >> (3 * v)) & 7 for v in range(n)]
        records.append(e)
        if kind in [10, 11]:
            break
    oracle = [list(a) for a in itertools.product(range(k), repeat=n) if all(a[u] != a[v] for u, v in edges)]
    outputs = [e['coloring'] for e in records if e['name'] == 'OUTPUT']
    expected = oracle[:1] if cut else oracle
    assert sorted(outputs) == sorted(expected), (name, outputs, expected)
    assert records[-1]['name'] == 'COMPLETE' and records[-1]['count'] == len(expected)
    report[name] = dict(vertices=n, colors=k, firstOnly=bool(cut), edges=edges,
                        source_sha256=hashlib.sha256(path.read_bytes()).hexdigest(),
                        events=len(records), solutions=len(outputs), outputs=outputs,
                        histogram=dict(Counter(e['name'] for e in records)), records=records)
(OUT / 'report-decoded-traces.json').write_text(json.dumps(report, indent=2) + '\n')
tables = []
for name, limit in [('triangle', 24), ('unsatisfiable', 14), ('root', 3), ('first', 12)]:
    tables += [f'## {name}', '', '| Seq | Event | Domains | P | C | T | Base | Count |',
               '|---:|---|---|---|---:|---:|---:|---:|']
    for e in report[name]['records'][:limit]:
        masks = ' '.join(f'{v:02X}' for v in e['domains'])
        tables.append(f"| {e['sequence']} | {e['name']} | {masks} | {e['propagated']} | {e['choiceTop']} | {e['trailTop']} | {e['base']} | {e['count']} |")
    tables.append('')
(OUT / '08-report-trace-tables.md').write_text('''---
Title: Decoded Physical Graph Trace Tables
Ticket: GATEMATE-SYMBOLIC-006
Status: complete
Topics: [fpga, gatemate]
DocType: reference
Intent: long-term
Summary: Generated report evidence from archived physical UART frames.
---

''' + '\n'.join(tables))
print(json.dumps({name: {key: value[key] for key in ['events', 'solutions', 'outputs', 'histogram']} for name, value in report.items()}, indent=2))

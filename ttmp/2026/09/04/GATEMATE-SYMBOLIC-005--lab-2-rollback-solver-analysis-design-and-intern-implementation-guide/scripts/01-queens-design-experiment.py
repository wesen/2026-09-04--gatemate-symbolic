#!/usr/bin/env python3
"""Design experiment, not production Lab 2: compare snapshot/trail semantics."""
from collections import Counter
from pathlib import Path
import hashlib
import json
import random

OUT = Path(__file__).resolve().parents[1] / 'reference/validation'
OUT.mkdir(exist_ok=True)

def independent():
    results = []
    def visit(rows):
        c = len(rows)
        if c == 8:
            results.append(tuple(rows)); return
        for r in range(8):
            if all(r != x and abs(r-x) != c-i for i,x in enumerate(rows)):
                visit(rows + [r])
    visit([])
    return results

def one(mask):
    return mask != 0 and mask & (mask-1) == 0

def attack(c0, r0, c):
    rows = (r0, r0+c-c0, r0-c+c0)
    return sum(1 << r for r in set(rows) if 0 <= r < 8)

class Solver:
    def __init__(self, mode, first=False, capacity=64, choices=8):
        self.mode, self.first, self.capacity, self.choice_capacity = mode,first,capacity,choices
        self.domains, self.propagated, self.trail = [255]*8,0,[]
        self.level = 0
        self.results, self.events = [],[]
        self.counters = Counter()
        self.rng = random.Random(42)

    def write(self, c, mask):
        old = self.domains[c]
        if old == mask: return
        assert mask & ~old == 0
        if self.mode == 'trail':
            if len(self.trail) == self.capacity: raise RuntimeError('TRAIL_FULL')
            self.trail.append((c,old))
            self.counters['trail_pushes'] += 1
            self.counters['max_trail'] = max(self.counters['max_trail'],len(self.trail))
        self.domains[c] = mask
        self.propagated &= ~(1 << c)
        self.counters['domain_writes'] += 1
        self.events.append(('write',c,old,mask,tuple(self.domains),self.propagated))

    def propagate(self):
        while True:
            if 0 in self.domains:
                self.counters['contradictions'] += 1; return False
            sources = [c for c in range(8) if one(self.domains[c]) and not self.propagated & (1<<c)]
            if not sources: return True
            c0 = sources[0]
            r0 = self.domains[c0].bit_length()-1
            for c in range(8):
                self.counters['propagation_visits'] += 1
                if c == c0: continue
                self.write(c,self.domains[c] & (~attack(c0,r0,c) & 255))
                if self.domains[c] == 0:
                    self.counters['contradictions'] += 1; return False
            self.propagated |= 1 << c0

    def run(self):
        if not self.propagate(): return False
        if all(one(mask) for mask in self.domains):
            frozen = (self.domains.copy(),self.propagated,self.trail.copy(),len(self.results))
            for _ in range(self.rng.randrange(8)):
                self.counters['result_stalls'] += 1
                assert frozen == (self.domains,self.propagated,self.trail,len(self.results))
            rows = tuple(mask.bit_length()-1 for mask in self.domains)
            self.results.append(rows)
            self.events.append(('solution',rows))
            return self.first
        if self.level == self.choice_capacity: raise RuntimeError('CHOICE_FULL')
        if self.mode == 'trail' and len(self.trail) == self.capacity: raise RuntimeError('TRAIL_FULL')
        c = next(c for c,mask in enumerate(self.domains) if not one(mask))
        saved,prop,mark = self.domains.copy(),self.propagated,len(self.trail)
        self.counters['choice_creations'] += 1
        if self.mode == 'snapshot': self.counters['snapshot_payload_bits'] += 72
        self.level += 1
        self.counters['max_choices'] = max(self.counters['max_choices'],self.level)
        for row in range(8):
            if not saved[c] & (1<<row): continue
            self.counters['alternatives'] += 1
            self.write(c,1<<row)
            if self.run(): return True
            if self.mode == 'trail':
                while len(self.trail) > mark:
                    col,old = self.trail.pop()
                    self.domains[col] = old
                    self.counters['trail_pops'] += 1
            else:
                self.domains = saved.copy()
                self.counters['snapshot_restored_payload_bits'] += 72
            self.propagated = prop
            assert self.domains == saved
            self.counters['restoration_checks'] += 1
            self.events.append(('restored',tuple(saved),prop))
        self.level -= 1
        return False

oracle = independent()
assert len(oracle) == len(set(oracle)) == 92
assert oracle[0] == (0,4,7,5,2,6,1,3)
summary = {}
runs = {}
for mode,first in [('snapshot',False),('trail',False),('trail',True)]:
    solver = Solver(mode,first)
    solver.run()
    assert solver.results == (oracle[:1] if first else oracle)
    key = mode + ('_first' if first else '')
    runs[key] = solver
    summary[key] = dict(solver.counters)
    summary[key]['solutions'] = len(solver.results)
    summary[key]['remaining_trail'] = len(solver.trail)
    summary[key]['ordered_solution_sha256'] = hashlib.sha256(json.dumps(solver.results).encode()).hexdigest()
assert runs['snapshot'].events == runs['trail'].events
assert runs['trail'].counters['max_trail'] <= 64
for capacity,choices,expected in [(1,8,'TRAIL_FULL'),(64,0,'CHOICE_FULL')]:
    solver = Solver('trail',capacity=capacity,choices=choices)
    try:
        solver.run()
        raise AssertionError('expected capacity fault')
    except RuntimeError as error:
        assert str(error) == expected
        if expected == 'TRAIL_FULL':
            assert solver.domains == [1]+[255]*7
        else:
            assert solver.domains == [255]*8 and not solver.trail
summary['checks'] = ['independent ordered 92-solution oracle','snapshot/trail semantic events equal',
                     'restoration snapshots equal','seeded abstract output stalls preserve state',
                     'FIRST_ONLY produces oracle prefix','precise tiny-capacity domain checks']
summary['first_board'] = list(oracle[0])
summary['first_solution24_hex'] = f'{sum(r << (3*c) for c,r in enumerate(oracle[0])):06x}'
(OUT/'design-experiment.json').write_text(json.dumps(summary,indent=2)+'\n')
(OUT/'solutions.json').write_text(json.dumps(oracle,indent=2)+'\n')
print(json.dumps(summary,indent=2))

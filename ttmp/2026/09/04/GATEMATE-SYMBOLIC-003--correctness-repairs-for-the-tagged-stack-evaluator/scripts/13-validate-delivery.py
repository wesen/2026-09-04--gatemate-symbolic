#!/usr/bin/env python3
"""Validate final phase evidence without rerunning or reprinting it."""
from pathlib import Path
import json
TICKET=Path(__file__).resolve().parents[1]
OUT=TICKET/'reference/validation'
assert '197 passed' in (OUT/'P4-final.log').read_text()
assert (OUT/'P5-build.exit').read_text().strip()=='0'
assert (OUT/'P5-hardware.exit').read_text().strip()=='0'
board=json.loads((OUT/'P5-hardware.json').read_text())
assert [r['program'] for r in board]==['fib','arith','typefault','countdown']
assert all(r['match'] for r in board)
for phase in ['PLAN','P1','P2','P3','P4','P5']:
    for state in (['start'] if phase=='PLAN' else ['start','done']):
        receipt=(TICKET/f'reference/slips/{phase}-{state}.log').read_text()
        assert 'printed: true' in receipt,(phase,state)
coverage=json.loads((OUT/'P4-coverage.json').read_text())
assert len(coverage['faults'])==9
assert coverage['branches']['taken'] and coverage['branches']['not_taken']
print('197 tests; 9 fault kinds; both branch outcomes; four board captures; 11 printed slips verified')

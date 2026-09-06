#!/usr/bin/env python3
"""Check retained completion receipts without rerunning qualified workloads."""
from pathlib import Path
import json
root = Path(__file__).resolve().parents[1]
checks = {}
checks['tasks_complete'] = '- [ ]' not in (root/'tasks.md').read_text()
checks['ticket_closed'] = 'Status: complete' in (root/'index.md').read_text()
for phase in range(1,7):
    for edge in ['start','done']:
        checks[f'i{phase}_{edge}_printed'] = 'printed: yes' in (root/f'reference/validation/i{phase}-{edge}-print.log').read_text()
checks['plan_printed'] = 'printed: yes' in (root/'reference/validation/implementation-plan-print.log').read_text()
checks['physical_programs_pass'] = '--- PASS: TestPhysicalPrograms' in (root/'reference/validation/i4-physical-tests.log').read_text()
checks['routing_pass'] = '21.26 MHz (PASS at 10.00 MHz)' in (root/'reference/validation/i4-build.log').read_text()
checks['eleven_screenshots'] = len(list((root/'reference/screenshots').glob('*.png')))==11
checks['guide_uploaded'] = 'OK: uploaded GATEMATE 010 Implemented Language and IDE Guide.pdf' in (root/'reference/validation/implementation-remarkable-upload.log').read_text()
checks['docmgr_clean'] = 'All checks passed' in (root/'reference/validation/final-docmgr-doctor.log').read_text()
(root/'reference/validation/completion-check.json').write_text(json.dumps(checks,indent=2)+'\n')
assert all(checks.values()), checks
print(f'{len(checks)} completion checks passed')

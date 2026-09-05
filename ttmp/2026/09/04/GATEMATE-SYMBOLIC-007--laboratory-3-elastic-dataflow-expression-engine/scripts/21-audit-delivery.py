#!/usr/bin/env python3
"""Audit final ticket deliverables without rerunning already-passed code tests."""
from pathlib import Path
import re,json
root=Path(__file__).resolve().parents[1]
images=sorted((root/'reference/screenshots').glob('*.png'))
slips=sorted((root/'reference/slips').glob('*.log'))
assert len(images)==10,len(images)
assert len(slips)==16,len(slips)
for p in slips:assert 'printed: true' in p.read_text(),p
assert '- [ ]' not in (root/'tasks.md').read_text()
assert 'Status: complete' in (root/'index.md').read_text()
assert 'OK: uploaded' in (root/'reference/validation/P7-upload.log').read_text()
checks=(root/'reference/validation/P7-check-summary.log').read_text().splitlines()
assert len(checks)==10 and all(s.startswith('PASS ') for s in checks)
missing=[]
for p in [root/'reference/03-implemented-laboratory-handoff-and-screenshot-atlas.md',root/'index.md']:
 for raw in re.findall(r'\]\(([^)]+)\)',p.read_text()):
  target=raw.split('#',1)[0]
  if target and '://' not in target and not (p.parent/target).exists():missing.append((str(p),target))
assert not missing,missing
print(json.dumps({'ticket':'complete','screenshots':len(images),'printedReceipts':len(slips),'checks':checks,'localLinks':'valid','remarkable':'uploaded'},indent=2))

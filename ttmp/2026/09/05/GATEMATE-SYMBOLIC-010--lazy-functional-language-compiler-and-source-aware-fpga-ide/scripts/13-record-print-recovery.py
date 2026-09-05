#!/usr/bin/env python3
from pathlib import Path
import json
root=Path(__file__).resolve().parents[1]
old=root.parent/'GATEMATE-SYMBOLIC-009--laboratory-4-lazy-graph-reducer-and-heap-inspector'
receipts={}
for directory,name in [(old,'report-done')]+[(root,n) for n in ['plan','d1-start','d1-done','d2-start','d2-done','d3-start','d3-done']]:
 matches=[p for p in (directory/'reference/validation').glob(name+'*print.log') if 'printed: true' in p.read_text()]
 assert matches, f'No successful print receipt: {name}'
 receipts[name]=[str(p.relative_to(root.parent)) for p in matches]
(root/'reference/validation/print-recovery.json').write_text(json.dumps(receipts,indent=2)+'\n')
p=root/'scripts/10-audit.py';s=p.read_text()
s=s.replace("log=root/f'reference/validation/{name}-print.log'\n slips[name]='printed' if log.exists() and 'printed: true' in log.read_text() else 'pending'", "logs=(root/'reference/validation').glob(name+'*print.log')\n slips[name]='printed' if any('printed: true' in p.read_text() for p in logs) else 'pending'")
p.write_text(s)
for p in [root/'index.md',root/'sources/index-body.md']:
 s=p.read_text().replace('I1–I6 and the explicit pending-print task remain open.','I1–I6 remain open. All eight delayed report/design slips printed successfully after the user repaired the service.')
 s=s.replace('Eight report/design slips remain pending; list or replay them with scripts/11-replay-pending-slips.sh.','The eight delayed report/design slips are now printed; successful timestamped receipts are indexed in reference/validation/print-recovery.json.')
 p.write_text(s)
print(json.dumps({'confirmed_prints':len(receipts)},indent=2))

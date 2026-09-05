#!/usr/bin/env python3
from pathlib import Path
root=Path(__file__).resolve().parents[1]
p=root/'index.md';s=p.read_text();s=s.replace('The target reMarkable folder is /ai/2026/09/05/GATEMATE-SYMBOLIC-010.','The 23-page guide was uploaded successfully as GATEMATE 010 Lazy Functional Language Design.pdf to /ai/2026/09/05/GATEMATE-SYMBOLIC-010. The separate source-evidence document remains in the ticket. D1–D3 are complete; I1–I6 and the explicit pending-print task remain open.')
s=s.replace('Only logs with printed: true establish physical printing.','Only logs with printed: true establish physical printing. Eight report/design slips remain pending; list or replay them with scripts/11-replay-pending-slips.sh. See reference/validation/remarkable-upload.log and delivery-audit.json for delivery evidence.')
p.write_text(s)
(root/'sources/index-body.md').write_text(s.split('---',2)[2].lstrip())
old=root.parent/'GATEMATE-SYMBOLIC-009--laboratory-4-lazy-graph-reducer-and-heap-inspector/index.md'
s=old.read_text()
if '## Published project report' not in s:
 s+='\n## Published project report\n\nThe [textbook-style reducer report](reference/03-project-report-inside-the-physically-qualified-lazy-graph-reducer.md) explains the implemented machine with five physical screenshots. It was published to go-go-parc in commit 23de3f4. The lazy-language follow-up design is in GATEMATE-SYMBOLIC-010.\n'
old.write_text(s)

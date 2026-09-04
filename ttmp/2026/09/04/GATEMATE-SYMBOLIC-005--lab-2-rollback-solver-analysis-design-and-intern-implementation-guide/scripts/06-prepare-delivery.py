#!/usr/bin/env python3
"""Validate the recovered guide and prepare portable image paths for PDF export."""
from pathlib import Path
import json
import re
import yaml

ticket = Path(__file__).resolve().parents[1]
source = ticket / 'design-doc/01-lab-2-snapshot-to-trail-refactoring-and-intern-implementation-guide.md'
text = source.read_text()
metadata = yaml.safe_load(text.split('---',2)[1])
assert metadata['Ticket'] == 'GATEMATE-SYMBOLIC-005'
assert len(text.split()) > 7000
assert '<!--' not in text
assert len(re.findall(r'^```',text,re.M)) % 2 == 0
results = json.loads((ticket/'reference/validation/design-experiment.json').read_text())
assert results['snapshot']['solutions'] == results['trail']['solutions'] == 92
assert results['trail']['max_trail'] == 32
assert results['first_solution24_hex'] == '672be0'
images = re.findall(r'!\[[^]]*\]\(([^)]+)\)',text)
assert len(images) == 2
for image in images:
    path = (source.parent/image).resolve()
    assert path.is_file()
    text = text.replace(f']({image})', f']({path})')
delivery = Path('/tmp/gatemate005-delivery')
delivery.mkdir(exist_ok=True)
text = ('---\ntitle: "Lab 2 Snapshot to Trail Design and Intern Guide"\n---\n'
        + text.split('---',2)[2])
(delivery/'guide.md').write_text(text)
report = {'words':len(text.split()), 'images':len(images), 'metadata':'valid',
          'experiment':'92 ordered solutions; snapshot/trail agreement',
          'draft_recovered':True, 'delivery_source':str(delivery/'guide.md')}
(ticket/'reference/validation/document-check.json').write_text(json.dumps(report,indent=2)+'\n')
print(json.dumps(report,indent=2))

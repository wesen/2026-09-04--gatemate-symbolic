#!/usr/bin/env python3
"""Check article metadata, quoted execution records, examples, and local links."""
from pathlib import Path
import contextlib
import io
import json
import re
import sys
import yaml

ticket = Path(__file__).resolve().parents[1]
project = next(p for p in ticket.parents if (p / 'symbolic_eval').is_dir()) / 'symbolic_eval'
article = Path(sys.argv[1])
text = article.read_text()
metadata = yaml.safe_load(text.split('---', 2)[1])
assert metadata['type'] == 'article'
assert metadata['source_revision'] == '9acc6cc737a1566b75edf427f57c6c357dfb87ce'
assert len(text.split()) >= 7000
fences = [line for line in text.splitlines() if line.startswith('```')]
assert len(fences) % 2 == 0
examples = json.loads((ticket / 'reference/validation/report-examples.json').read_text())
actual = {r['trace'] for e in examples.values() for r in e['rows']}
actual.update(e['final'] for e in examples.values())
quoted = [line for line in text.splitlines() if line.startswith(('TRACE ', 'FINAL '))]
assert all(line in actual for line in quoted), set(quoted) - actual
for example, events, depth, returns in [('fib',1681,12,10), ('fib3',47,5,3), ('countdown',55,3,1)]:
    e = examples[example]
    assert (e['events'], e['max_data_depth'], e['max_return_depth']) == (events,depth,returns)
for path in set(re.findall(r'`((?:rtl|tools|sim)/[\w/.]+\.(?:sv|py))(?::\d+)?`', text)):
    assert (project / path).is_file(), path
vault = Path('/home/manuel/code/wesen/go-go-golems/go-go-parc')
names = {p.stem for p in (vault / 'Projects').rglob('*.md')}
for link in re.findall(r'\[\[([^]|]+)(?:\|[^]]+)?\]\]', text):
    assert link in names, link
sys.path.insert(0, str(project / 'tools'))
namespace = {}
with contextlib.redirect_stdout(io.StringIO()):
    exec(re.search(r'```python\n(.*?)\n```', text, re.S).group(1), namespace)
assert namespace['machine'].halted
assert namespace['machine'].stack[0].payload == 12
print(json.dumps({'words':len(text.split()), 'quoted_records_verified':len(quoted),
                  'mermaid_diagrams':text.count('```mermaid'), 'python_example':'passed',
                  'metadata':'valid', 'vault_links':'resolved'}, indent=2))

#!/usr/bin/env python3
"""Recover the authored report from an unapplied patch in this session archive.

The initial patch combined delete/add of one path, which apply_patch rejected.
This one-time recovery writes only the still-empty docmgr report template.
"""
import argparse
import json
from pathlib import Path

p = argparse.ArgumentParser()
p.add_argument('transcript', type=Path)
a = p.parse_args()
ticket = Path(__file__).resolve().parents[1]
target = ticket / 'reference/02-inside-the-graph-coloring-search-microscope-technical-project-report.md'
assert '<!-- What is the purpose' in target.read_text(), 'Refusing to replace a completed report'
for line in a.transcript.open():
    record = json.loads(line).get('payload', {})
    source = record.get('input', '')
    prefix = 'text(await tools.apply_patch('
    if not source.startswith(prefix) or 'The graph-coloring laboratory implements a complete, configurable search system' not in source:
        continue
    patch, _ = json.JSONDecoder().raw_decode(source[len(prefix):])
    header = '*** Add File: ' + str(target) + '\n'
    body = patch.split(header, 1)[1].split('*** End Patch', 1)[0]
    lines = body.splitlines()
    assert all(line.startswith('+') for line in lines)
    article = '\n'.join(line[1:] for line in lines) + '\n'
    assert len(article.split()) > 5000
    target.write_text(article)
    print(f'Recovered {len(article.split())} words into {target}')
    break
else:
    raise SystemExit('Draft patch was not found')

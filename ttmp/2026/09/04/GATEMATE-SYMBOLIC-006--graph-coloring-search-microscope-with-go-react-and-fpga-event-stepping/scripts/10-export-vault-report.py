#!/usr/bin/env python3
"""Validate and export the full article plus local screenshot assets append-only."""
import argparse
import hashlib
import json
from pathlib import Path
import re
import shutil
import yaml

p = argparse.ArgumentParser()
p.add_argument('--destination-directory', type=Path, default=Path('/tmp/gatemate-graph-report'))
a = p.parse_args()
ticket = Path(__file__).resolve().parents[1]
source = ticket / 'reference/02-inside-the-graph-coloring-search-microscope-technical-project-report.md'
raw = source.read_text()
metadata, body = raw.split('---', 2)[1:]
assert yaml.safe_load(metadata)['Ticket'] == 'GATEMATE-SYMBOLIC-006'
assert len(body.split()) > 7000
assert len(re.findall(r'^```', body, re.M)) % 2 == 0
assert body.count('```mermaid') == 3
assert '<!--' not in body
trace = json.loads((ticket / 'reference/validation/report-decoded-traces.json').read_text())
assert [(trace[n]['events'], trace[n]['solutions']) for n in ['triangle', 'unsatisfiable', 'path', 'first', 'root']] == [(86, 6), (26, 0), (32, 2), (12, 1), (3, 0)]
for event in trace['triangle']['records'][:20]:
    masks = ' '.join(f'{v:02X}' for v in event['domains'])
    row = f"| {event['sequence']} | {event['name']} | `{masks}` | `{event['propagated']}` | {event['choiceTop']} | {event['trailTop']} | {event['count']} |"
    assert row in body, row
assets = {
    'P6-desktop.png': 'gatemate-graph-microscope-desktop.png',
    'report-rollback.png': 'gatemate-graph-microscope-rollback.png',
    'P6-mobile-history.png': 'gatemate-graph-microscope-mobile-history.png',
}
filename = 'ARTICLE - GateMate Symbolic - Inside the Graph Coloring Search Microscope.md'
dest = a.destination_directory / filename
assert not dest.exists(), f'Refusing to replace existing article: {dest}'
for old, new in assets.items():
    image = ticket / 'reference/validation' / old
    assert image.is_file() and image.read_bytes().startswith(b'\x89PNG\r\n\x1a\n')
    target = a.destination_directory / '_assets' / new
    assert not target.exists(), f'Refusing to replace existing asset: {target}'
    body = body.replace(f'(validation/{old})', f'(_assets/{new})')
assert len(re.findall(r'!\[.*?\]\(_assets/.*?\)', body)) == 3
front = dict(title='Inside the Graph Coloring Search Microscope: Reversible FPGA Search from Constraints to a Browser',
             aliases=['GateMate Graph Coloring Technical Report', 'Lab 3 Search Microscope'],
             tags=['article', 'fpga', 'gatemate', 'constraint-solving', 'golang', 'react'],
             status='complete', type='article', created='2026-09-04',
             repo='/home/manuel/code/wesen/2026-09-04--gatemate-symbolic',
             source_revision='9e7868721c3e80208b3534a33ac7aa98f4ec4407',
             related_tickets=['GATEMATE-SYMBOLIC-006'])
body += '\n## Related vault notes\n\n- [[ARTICLE - GateMate Symbolic - Inside an FPGA Rollback Solver]] explains the preceding queens laboratory and its recovery machinery.\n- [[ARTICLE - GateMate Symbolic - Inside a Tagged Stack CPU]] describes the first laboratory and its processor architecture.\n- [[ARTICLE - Playbook - GateMate Board Evidence Workflow]] documents the evidence workflow for physical board experiments.\n'
dest.parent.mkdir(parents=True, exist_ok=True)
(dest.parent / '_assets').mkdir(exist_ok=True)
for old, new in assets.items():
    shutil.copyfile(ticket / 'reference/validation' / old, dest.parent / '_assets' / new)
dest.write_text('---\n' + yaml.safe_dump(front, sort_keys=False, allow_unicode=True) + '---\n' + body)
receipt = dict(article=str(dest), body_words=len(body.split()), mermaid_diagrams=3,
               screenshots=[str(dest.parent / '_assets' / v) for v in assets.values()],
               source_sha256=hashlib.sha256(source.read_bytes()).hexdigest(),
               article_sha256=hashlib.sha256(dest.read_bytes()).hexdigest(),
               triangle_trace_rows_verified=20)
print(json.dumps(receipt, indent=2))

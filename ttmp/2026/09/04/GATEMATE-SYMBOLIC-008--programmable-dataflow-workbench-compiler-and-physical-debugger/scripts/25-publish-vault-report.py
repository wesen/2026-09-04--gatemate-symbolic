#!/usr/bin/env python3
"""Validate and append the illustrated report to the go-go-parc vault."""
import argparse
import hashlib
import json
import re
from pathlib import Path

import yaml

parser = argparse.ArgumentParser()
parser.add_argument('--publish', action='store_true')
args = parser.parse_args()
ticket = Path(__file__).resolve().parents[1]
repo = ticket.parents[4]
source = ticket / 'various/programmable-dataflow-vault-report.md'
text = source.read_text()
meta = yaml.safe_load(text.split('---', 2)[1])
assert meta['type'] == 'article' and meta['status'] == 'complete'
assert text.count('\n```') % 2 == 0
assert len(text.split()) > 5000
assets = {
    'gatemate-workbench-fpga-breakpoint.png': '05-fpga-breakpoint.png',
    'gatemate-workbench-fpga-result-trace.png': '06-fpga-result-trace.png',
    'gatemate-workbench-fpga-history.png': '07-fpga-history.png',
    'gatemate-workbench-model-breakpoint.png': '01-model-breakpoint.png',
    'gatemate-workbench-model-result.png': '02-model-result.png',
    'gatemate-workbench-model-mobile.png': '04-model-mobile.png',
}
figures = re.findall(r'!\[[^]]*\]\(_assets/([^)]*)\)', text)
assert set(figures) == set(assets), figures
for relative in re.findall(r'https://github.com/wesen/2026-09-04--gatemate-symbolic/blob/85969b4/([^)]*)', text):
    assert (repo / relative).is_file(), relative
evidence = json.loads((ticket / 'reference/validation/p6-browser-fpga.json').read_text())
assert evidence['halted']['cycle'] == 6
assert evidence['result']['value'] == 30
debug = evidence['completed']['snapshot']['debug']
assert len(debug['events']) == 21 and debug['dropped'] == 8
assert any(e['cycle'] == 32 and e['kind'] == 3 and e['token']['value'] == 30 for e in debug['events'])
vault = Path('/home/manuel/code/wesen/go-go-golems/go-go-parc')
folder = vault / 'Projects/2026/09/05'
destination = folder / 'ARTICLE - GateMate Symbolic - Programmable Dataflow Workbench.md'
hashes = {}
copies = [(source, destination)]
for name, original in assets.items():
    origin = ticket / 'reference/screenshots' / original
    data = origin.read_bytes()
    assert data.startswith(b'\x89PNG\r\n\x1a\n'), origin
    hashes[name] = hashlib.sha256(data).hexdigest()
    copies.append((origin, folder / '_assets' / name))
for origin, target in copies:
    if target.exists():
        assert target.read_bytes() == origin.read_bytes(), f'Would overwrite different historical content: {target}'
if args.publish:
    for origin, target in copies:
        target.parent.mkdir(parents=True, exist_ok=True)
        if not target.exists():
            with target.open('xb') as out:
                out.write(origin.read_bytes())
        assert target.read_bytes() == origin.read_bytes()
print(json.dumps({'published': args.publish, 'words': len(text.split()), 'figures': hashes,
                  'note': str(destination), 'source_revision': meta['source_revision']}, indent=2))

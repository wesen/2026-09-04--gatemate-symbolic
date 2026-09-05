#!/usr/bin/env python3
"""Validate and append the dataflow article and immutable screenshot copies."""
import argparse
import hashlib
import json
from pathlib import Path
import re
import shutil
import yaml

TICKET = Path(__file__).resolve().parents[1]
REPO = TICKET.parents[4]
VAULT = Path('/home/manuel/code/wesen/go-go-golems/go-go-parc')
REPORT = TICKET / 'reference/04-inside-an-elastic-dataflow-engine.md'
DEST = VAULT / 'Projects/2026/09/04'
NAME = 'ARTICLE - GateMate Symbolic - Inside an Elastic Dataflow Engine.md'
IMAGES = {
    'gatemate-dataflow-book.png': '06-fpga-book-complete.png',
    'gatemate-dataflow-held-output.png': '07-fpga-held-output-guard.png',
    'gatemate-dataflow-in-flight.png': '08-fpga-in-flight.png',
    'gatemate-dataflow-cancellation.png': '09-fpga-cancellation-complete.png',
    'gatemate-dataflow-history.png': '10-fpga-history.png',
}

def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--publish', action='store_true')
    args = parser.parse_args()
    source = REPORT.read_text()
    front = yaml.safe_load(source.split('---', 2)[1])
    assert front['type'] == 'article' and front['source_revision'].startswith('9953407')
    assert len(re.findall(r'^```', source, re.M)) % 2 == 0
    article = source
    hashes = {}
    for name, original in IMAGES.items():
        original_path = TICKET / 'reference/screenshots' / original
        assert original_path.read_bytes().startswith(b'\x89PNG\r\n\x1a\n')
        assert f'(screenshots/{original})' in source
        article = article.replace(f'(screenshots/{original})', f'(_assets/{name})')
        hashes[name] = digest(original_path)
    for link in re.findall(r'\[\[([^]|]+)', article):
        assert list(VAULT.rglob(link + '.md')), link
    # Check explicit repository file references, including grouped paths where possible.
    paths = re.findall(r'`((?:pkg|internal|web|elastic_dataflow)/[^` ,]+\.(?:go|sv|tsx|sh))`', source)
    for path in paths:
        assert (REPO / path).is_file(), path
    if args.publish:
        target = DEST / NAME
        assert not target.exists(), 'append-only: destination article already exists'
        assets = DEST / '_assets'
        assets.mkdir(parents=True, exist_ok=True)
        for name, original in IMAGES.items():
            target_image = assets / name
            assert not target_image.exists() or digest(target_image) == hashes[name]
            if not target_image.exists():
                shutil.copy2(TICKET / 'reference/screenshots' / original, target_image)
        target.write_text(article)
        assert target.read_text() == article
        for name in IMAGES:
            assert digest(assets / name) == hashes[name]
    result = {'article': str(DEST / NAME), 'words': len(source.split()),
              'screenshots': hashes, 'mermaid_diagrams': source.count('```mermaid'),
              'checked_source_paths': len(paths), 'published': args.publish}
    print(json.dumps(result, indent=2))

if __name__ == '__main__':
    main()

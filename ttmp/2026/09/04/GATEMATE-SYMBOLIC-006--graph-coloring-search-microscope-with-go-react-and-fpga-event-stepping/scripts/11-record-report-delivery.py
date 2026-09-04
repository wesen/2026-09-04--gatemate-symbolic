#!/usr/bin/env python3
"""Record verified vault delivery without modifying the vault or device."""
import hashlib
import json
from pathlib import Path
import subprocess

T = Path(__file__).resolve().parents[1]
vault = Path('/home/manuel/code/wesen/go-go-golems/go-go-parc')
receipt = json.loads((T / 'reference/validation/report-vault-export.json').read_text())
head = subprocess.check_output(['git', '-C', str(vault), 'rev-parse', 'HEAD'], text=True).strip()
remote = subprocess.check_output(['git', '-C', str(vault), 'rev-parse', 'origin/main'], text=True).strip()
assert head == remote
article = Path(receipt['article'])
assert hashlib.sha256(article.read_bytes()).hexdigest() == receipt['article_sha256']
mapping = {'P6-desktop.png': 'gatemate-graph-microscope-desktop.png',
           'P6-mobile-history.png': 'gatemate-graph-microscope-mobile-history.png',
           'report-rollback.png': 'gatemate-graph-microscope-rollback.png'}
for old, new in mapping.items():
    assert (T / 'reference/validation' / old).read_bytes() == (article.parent / '_assets' / new).read_bytes()
receipt.update(vault_commit=head, origin_main=remote, assets_byte_identical=True,
               markdown_parsed_with='pandoc',
               obsidian_reading_view='Not verified: installed Obsidian launcher reports outdated installer without CLI support.',
               new_screenshot=dict(engine='serial', graph='triangle with three colors', generation=7,
                                   latest_event=86, total_colorings=6, inspected_event=13,
                                   domains=[1, 6, 6], propagated=255, choice_top=2, trail_top=3))
(T / 'reference/validation/report-vault-delivery.json').write_text(json.dumps(receipt, indent=2) + '\n')
print(json.dumps(receipt, indent=2))

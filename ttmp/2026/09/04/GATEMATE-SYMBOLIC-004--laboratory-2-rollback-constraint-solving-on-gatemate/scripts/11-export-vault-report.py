#!/usr/bin/env python3
"""Prepare and validate an append-only Obsidian report from the ticket article."""
from pathlib import Path
import argparse
import re
import yaml

p = argparse.ArgumentParser()
p.add_argument('--destination',type=Path,default=Path('/tmp/queens-vault-report.md'))
a = p.parse_args()
ticket = Path(__file__).resolve().parents[1]
source = ticket/'reference/02-inside-the-eight-queens-rollback-engine-technical-project-report.md'
body = source.read_text().split('---',2)[2].strip()
assert len(re.findall(r'^```',body,re.M)) % 2 == 0
assert len(body.split()) > 5000
assert '<!--' not in body
assert all(term in body for term in ['4000080280100401','672BE0','74,523','106,480','RAM_HALF'])
metadata = dict(title='Inside an FPGA Rollback Solver: Eight Queens, Reversible State, and Synchronous Memory',
    aliases=['GateMate Eight-Queens Technical Deep Dive','Lab 2 Rollback Solver Architecture'],
    tags=['article','fpga','gatemate','constraint-solving','computer-architecture','systemverilog'],
    status='complete',type='article',created='2026-09-04',
    repo='/home/manuel/code/wesen/2026-09-04--gatemate-symbolic',
    source_revision='4345ee4a9eb3460d6e94646d77e08dffa91489a0',
    related_tickets=['GATEMATE-SYMBOLIC-004','GATEMATE-SYMBOLIC-005'])
related = '\n\n## Related vault notes\n\n- [[ARTICLE - GateMate Symbolic - Inside a Tagged Stack CPU]] explains the preceding laboratory and its distinct processor architecture.\n- [[ARTICLE - Playbook - GateMate Board Evidence Workflow]] describes the physical evidence workflow.\n'
text = '---\n'+yaml.safe_dump(metadata,sort_keys=False,allow_unicode=True)+'---\n\n'+body+related
assert not a.destination.exists(), f'Refusing to replace existing note: {a.destination}'
a.destination.parent.mkdir(parents=True,exist_ok=True)
a.destination.write_text(text)
print(f'Wrote {len(body.split())} body words to {a.destination}')

#!/usr/bin/env python3
"""Recover this turn's authored guide from a rejected multi-operation patch.

Reads only the explicitly named local session and extracts our exact guide patch;
does not print transcript content or rewrite any file except the generated guide.
"""
import json
from pathlib import Path
import sys

ticket = Path(__file__).resolve().parents[1]
target = ticket / 'design-doc/01-lab-2-snapshot-to-trail-refactoring-and-intern-implementation-guide.md'
marker = '# Laboratory 2: from board snapshots to a mutation trail'
def strings(value):
    if isinstance(value,str): yield value
    elif isinstance(value,dict):
        for child in value.values(): yield from strings(child)
    elif isinstance(value,list):
        for child in value: yield from strings(child)
patches = []
for line in Path(sys.argv[1]).read_text().splitlines():
    for candidate in strings(json.loads(line)):
        if marker not in candidate or '*** Delete File:' not in candidate:
            continue
        prefix = 'tools.apply_patch('
        if prefix in candidate:
            argument = candidate.split(prefix,1)[1]
            if not argument.startswith('"'): continue
            patch,_ = json.JSONDecoder().raw_decode(argument)
            if isinstance(patch,str) and marker in patch: patches.append(patch)
assert patches, 'Authored patch not found; no files changed'
patch = patches[-1]
body = patch.split(f'*** Add File: {target}\n',1)[1].split('*** End Patch',1)[0]
lines = body.splitlines()
assert all(line.startswith('+') for line in lines)
text = '\n'.join(line[1:] for line in lines)+'\n'
assert text.startswith('---\nTitle:') and marker in text
assert len(text.split()) > 6000
target.write_text(text)
print(f'Recovered authored guide: {len(text.split())} words')

#!/usr/bin/env python3
"""Append a strict-format diary step from a retained JSON entry."""
import json
import sys
from pathlib import Path
root = Path(__file__).resolve().parents[1]
data = json.loads(Path(sys.argv[1]).read_text())
path = root / 'reference/01-implementation-diary.md'
text = path.read_text()
if '## Step ' not in text:
    text = text.split('---', 2)[0] + '---' + text.split('---', 2)[1] + '---\n\n# Implementation diary\n\n## Goal\n\nRecord the programmable workbench design, implementation, validation, physical evidence, and delivery.\n'
number = text.count('\n## Step ') + 1
prompt = 'See Step 1.'
if number == 1:
    prompt = 'Exact request as JSON:\n\n```json\n' + json.dumps(json.loads((root/'sources/user-request.json').read_text())['prompt']) + '\n```'
text += f'\n## Step {number}: {data["title"]}\n\n{data["prose"]}\n\n### Prompt Context\n\n**User prompt (verbatim):** {prompt}\n\n**Assistant interpretation:** Build the programmable graph/compiler/debugger workbench with an intern guide, physical evidence, and phased delivery.\n\n**Inferred user intent:** Make the existing FPGA execution engine programmable and inspectable while preserving a reviewable account of its correctness.\n'
if data.get('commit'):
    text += '\n**Commit (code):** ' + data['commit'] + '\n'
for key, heading in [('did','What I did'),('why','Why'),('worked','What worked'),('failed',"What didn't work"),('learned','What I learned'),('tricky','What was tricky to build'),('review','What warrants a second pair of eyes'),('future','What should be done in the future'),('validate','Code review instructions'),('details','Technical details')]:
    text += '\n### ' + heading + '\n\n' + '\n'.join('- ' + x for x in data.get(key,['N/A'])) + '\n'
path.write_text(text)

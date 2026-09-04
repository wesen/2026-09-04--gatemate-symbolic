#!/usr/bin/env python3
"""Append a fully structured chronological diary step."""
import argparse
from pathlib import Path
p=argparse.ArgumentParser()
for name in ['title','summary','did','why','worked','failed','learned','tricky','review','future','validate','details','commit']:
    p.add_argument('--'+name, required=name not in ['commit'])
a=p.parse_args()
f=Path(__file__).resolve().parents[1]/'reference/01-implementation-diary.md'
existing=f.read_text()
if '# Diary\n' not in existing:
    front=existing.split('---',2)[1]
    existing='---'+front+'---\n\n# Diary\n\n## Goal\n\nImplement and validate the review findings, preserving commands, decisions, commits, and printed phase receipts.\n'
step=existing.count('### Prompt Context')+1
prompt=('Create a new ticket, add design doc on how to fix things, then build them. commit at appropriate intervals and keep a detailed diary as you work (using the diary format from the skill)\n\nPrint out a brutalist work slip with the plan / different phases for the ticket. then before stsarting a phase, plrint a split about the phase, and print one when the phase is done.')
text=f'\n## Step {step}: {a.title}\n\n{a.summary}\n\n### Prompt Context\n\n**User prompt (verbatim):**\n\n'+(('> '+prompt.replace('\n','\n> ')) if step==1 else '(see Step 1)')+'\n\n**Assistant interpretation:** Implement the repair design in phases with reproducible validation, focused commits, detailed diary entries, and printed phase boundaries.\n\n**Inferred user intent:** Turn the documented review findings into a reliable working evaluator with visible progress and reviewable evidence.\n'
if a.commit: text+=f'\n**Commit (code/work):** {a.commit}\n'
for label,field in [('What I did','did'),('Why','why'),('What worked','worked'),("What didn't work",'failed'),('What I learned','learned'),('What was tricky to build','tricky'),('What warrants a second pair of eyes','review'),('What should be done in the future','future'),('Code review instructions','validate'),('Technical details','details')]:
    text+=f'\n### {label}\n\n{getattr(a,field)}\n'
f.write_text(existing+text)

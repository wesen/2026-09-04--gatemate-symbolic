#!/usr/bin/env python3
"""Append a strict implementation diary step with the current prompt context."""
import argparse
from pathlib import Path

p=argparse.ArgumentParser()
for field in ['title','summary','did','why','worked','failed','learned','tricky','review','future','validate','details']:
    p.add_argument('--'+field,required=True)
p.add_argument('--commit')
a=p.parse_args()
path=Path(__file__).resolve().parents[1]/'reference/01-diary.md'
old=path.read_text()
step=old.count('### Prompt Context')+1
prompt='Now implement it, commit at appropriate intervals and keep a detailed diary as you work (using the diary format from the skill)\n\n. Print out a brutalist work slip with the plan / different phases for the ticket. then before stsarting a phase, plrint a split about the phase, and print one when the phase is done.'
context='\n'.join(('> '+line).rstrip() for line in prompt.splitlines()) if step==2 else '(see Step 2)'
text=f'\n## Step {step}: {a.title}\n\n{a.summary}\n\n### Prompt Context\n\n**User prompt (verbatim):**\n\n{context}\n\n**Assistant interpretation:** Implement the Lab 2 guide in tested phases, with committed milestones and printed boundaries.\n\n**Inferred user intent:** Deliver a working and measurable rollback solver with reproducible evidence and a detailed implementation record.\n'
if a.commit: text+=f'\n**Commit (code/work):** {a.commit}\n'
for label,field in [('What I did','did'),('Why','why'),('What worked','worked'),("What didn't work",'failed'),('What I learned','learned'),('What was tricky to build','tricky'),('What warrants a second pair of eyes','review'),('What should be done in the future','future'),('Code review instructions','validate'),('Technical details','details')]:
    text+=f'\n### {label}\n\n{getattr(a,field)}\n'
path.write_text(old+text)

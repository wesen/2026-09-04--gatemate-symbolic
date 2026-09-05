#!/usr/bin/env python3
"""Task slips and strict chronological diary records for the dataflow laboratory."""
import argparse
import json
from pathlib import Path
import subprocess
import sys

T = Path(__file__).resolve().parents[1]
PHASES = {'P1':'intern guide and reMarkable delivery', 'P2':'typed semantic and transaction models',
          'P3':'elastic RTL scheduling and cancellation', 'P4':'UART host control and board execution',
          'P5':'stress tests measurements and handoff'}
p=argparse.ArgumentParser();s=p.add_subparsers(dest='command',required=True)
x=s.add_parser('slip');x.add_argument('phase',choices=['PLAN',*PHASES]);x.add_argument('state',choices=['start','done']);x.add_argument('--commit');x.add_argument('--fact',action='append',default=[])
x=s.add_parser('diary');x.add_argument('entry',type=Path)
s.add_parser('tidy')
a=p.parse_args()
if a.command=='tidy':
    for name in ['index.md','tasks.md','changelog.md']:
        f=T/name;f.write_text(f.read_text().rstrip()+'\n')
elif a.command=='slip':
    target=T/'reference/slips'/f'{a.phase}-{a.state}'
    cmd=[sys.executable,'/home/manuel/.pi/agent/skills/brutalist-work-slip/scripts/work_slip.py',
         'status' if a.state=='done' else 'plan','--task','GATEMATE-007','--label',f'{a.phase} {a.state.upper()}',
         '--title','Elastic dataflow expression engine','--out',str(target.with_suffix('.yaml'))]
    if a.phase=='PLAN':
        for phase,title in PHASES.items():cmd+=['--phase',phase+' '+title]
    else:cmd+=['--did' if a.state=='done' else '--phase',PHASES[a.phase]]
    cmd+=['--next','review completed lab' if a.phase=='P5' and a.state=='done' else 'continue validated tasks']
    if a.commit:
        commit=subprocess.check_output(['git','rev-parse',a.commit],text=True).strip()
        cmd+=['--commit',commit,'--repo','wesen/2026-09-04--gatemate-symbolic']
    for fact in a.fact:cmd+=['--fact',fact]
    r=subprocess.run(cmd,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,text=True)
    target.with_suffix('.log').write_text(r.stdout);print(r.stdout);sys.exit(r.returncode)
else:
    e=json.loads(a.entry.read_text());path=T/'reference/01-implementation-diary.md';old=path.read_text()
    n=old.count('### Prompt Context')+1
    if n==1:old='---'+old.split('---',2)[1]+'---\n\n# Implementation diary\n\n## Goal\n\nExplain every design, implementation, validation, and physical delivery step in the dataflow laboratory.\n'
    prompt=e.get('prompt','(see Step 1)')
    body=f'\n## Step {n}: {e["title"]}\n\n{e["summary"]}\n\n### Prompt Context\n\n**User prompt (verbatim, JSON encoded):**\n\n```json\n{json.dumps(prompt)}\n```\n\n**Assistant interpretation:** Design, publish, implement, and validate the book\'s elastic dataflow laboratory with task commits, detailed diary, and physical work slips.\n\n**Inferred user intent:** Build a working concurrent symbolic machine and preserve enough explanation and evidence for an intern to understand and maintain it.\n'
    if e.get('commit'):body+=f'\n**Commit:** {e["commit"]}\n'
    for label,key in [('What I did','did'),('Why','why'),('What worked','worked'),("What didn't work",'failed'),('What I learned','learned'),('What was tricky to build','tricky'),('What warrants a second pair of eyes','review'),('What should be done in the future','future'),('Code review instructions','validate'),('Technical details','details')]:body+=f'\n### {label}\n\n{e[key]}\n'
    path.write_text(old+body)

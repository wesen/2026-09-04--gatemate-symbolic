#!/usr/bin/env python3
"""Print phase slips and append strict diary steps for the graph microscope."""
import argparse
import json
from pathlib import Path
import subprocess
import sys

T = Path(__file__).resolve().parents[1]
PHASES = {'P1':'Design guide and reMarkable delivery','P2':'Go model and serial event contract',
          'P3':'Graph RTL and physical UART stepping','P4':'Go session service and HTTP API',
          'P5':'React search microscope interface','P6':'End to end validation and handoff'}
p=argparse.ArgumentParser()
sub=p.add_subparsers(dest='command',required=True)
s=sub.add_parser('slip');s.add_argument('phase',choices=['PLAN',*PHASES]);s.add_argument('state',choices=['start','done']);s.add_argument('--commit');s.add_argument('--fact',action='append',default=[])
d=sub.add_parser('diary')
for field in ['title','summary','did','why','worked','failed','learned','tricky','review','future','validate','details']:
    d.add_argument('--'+field,required=True)
d.add_argument('--commit')
a=p.parse_args()
if a.command=='slip':
    out=T/'reference/slips';out.mkdir(exist_ok=True)
    cmd=[sys.executable,'/home/manuel/.pi/agent/skills/brutalist-work-slip/scripts/work_slip.py','status' if a.state=='done' else 'plan',
         '--task','GATEMATE-006','--label',f'{a.phase} {a.state.upper()}','--title',PHASES.get(a.phase,'Graph coloring search microscope'),
         '--out',str(out/f'{a.phase}-{a.state}.yaml')]
    if a.phase=='PLAN':
        for phase,title in PHASES.items():cmd+=['--phase',f'{phase} {title}']
    else:cmd+=['--did' if a.state=='done' else '--phase',PHASES[a.phase]]
    cmd+=['--next','review completed microscope' if a.phase=='P6' and a.state=='done' else 'continue validated phases']
    if a.commit and a.state=='done':
        commit=subprocess.check_output(['git','rev-parse',a.commit],text=True).strip()
        cmd+=['--commit',commit,'--repo','wesen/2026-09-04--gatemate-symbolic']
    for fact in a.fact:cmd+=['--fact',fact]
    result=subprocess.run(cmd,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,text=True)
    (out/f'{a.phase}-{a.state}.log').write_text(result.stdout);print(result.stdout);sys.exit(result.returncode)
path=T/'reference/01-implementation-diary.md'
old=path.read_text()
if '### Prompt Context' not in old:old=old.split('---',2)[0]+'---'+old.split('---',2)[1]+'---\n\n# Implementation Diary\n\n## Goal\n\nRecord design, implementation, validation, failures, and physical delivery of the graph-coloring search microscope.\n'
n=old.count('### Prompt Context')+1
prompt='Ok, create anew docmgr ticket for that, Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable.\n\nThen implement it, using golang + typescript/react for the frontend.\n\ncommit at appropriate intervals and keep a detailed diary as you work (using the diary format from the skill).\n\nPrint out a brutalist work slip with the plan / different phases for the ticket. then before stsarting a phase, plrint a split about the phase, and print one when the phase is done.'
context='```json\n'+json.dumps(prompt)+'\n```' if n==1 else '(see Step 1)'
entry=f'\n## Step {n}: {a.title}\n\n{a.summary}\n\n### Prompt Context\n\n**User prompt (verbatim, JSON encoded):**\n\n{context}\n\n**Assistant interpretation:** Design, publish, and implement the graph-coloring FPGA microscope with Go and React, committed milestones, a detailed diary, and physical phase slips.\n\n**Inferred user intent:** Deliver a working next laboratory with enough explanation and evidence for an intern to maintain it.\n'
if a.commit:entry+=f'\n**Commit:** {a.commit}\n'
for label,key in [('What I did','did'),('Why','why'),('What worked','worked'),("What didn't work",'failed'),('What I learned','learned'),('What was tricky to build','tricky'),('What warrants a second pair of eyes','review'),('What should be done in the future','future'),('Code review instructions','validate'),('Technical details','details')]:entry+=f'\n### {label}\n\n{getattr(a,key)}\n'
path.write_text(old+entry)

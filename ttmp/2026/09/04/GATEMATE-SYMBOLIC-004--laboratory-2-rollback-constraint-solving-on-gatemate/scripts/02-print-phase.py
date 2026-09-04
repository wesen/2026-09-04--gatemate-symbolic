#!/usr/bin/env python3
"""Print and retain the requested Lab 2 plan and phase boundary slips."""
import argparse
from pathlib import Path
import subprocess
import sys

ticket = Path(__file__).resolve().parents[1]
phases = {'P1':'Executable model and independent oracle',
          'P2':'Full snapshot RTL baseline',
          'P3':'Trail RTL and reverse restoration',
          'P4':'Faults reset stalls and cut verification',
          'P5':'Board integration and measured results'}
p = argparse.ArgumentParser()
p.add_argument('phase',choices=['PLAN',*phases])
p.add_argument('state',choices=['start','done'])
p.add_argument('--commit')
p.add_argument('--fact',action='append',default=[])
a=p.parse_args()
out=ticket/'reference/slips'
out.mkdir(exist_ok=True)
cmd=[sys.executable,'/home/manuel/.pi/agent/skills/brutalist-work-slip/scripts/work_slip.py',
     'status' if a.state=='done' else 'plan','--task','GATEMATE-004',
     '--label',f'{a.phase} {a.state.upper()}', '--title',
     'Eight queens rollback solver' if a.phase=='PLAN' else phases[a.phase],
     '--out',str(out/f'{a.phase}-{a.state}.yaml')]
if a.phase=='PLAN':
    for phase,title in phases.items(): cmd += ['--phase',f'{phase} {title}']
    cmd += ['--next','P1 executable model']
else:
    cmd += ['--did' if a.state=='done' else '--phase',phases[a.phase]]
    cmd += ['--next','review completed solver' if a.phase=='P5' and a.state=='done' else 'continue tested implementation']
if a.commit and a.state=='done':
    commit=subprocess.check_output(['git','rev-parse',a.commit],text=True).strip()
    cmd += ['--commit',commit,'--repo','wesen/2026-09-04--gatemate-symbolic']
for fact in a.fact: cmd += ['--fact',fact]
result=subprocess.run(cmd,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
(out/f'{a.phase}-{a.state}.log').write_text(result.stdout)
print(result.stdout)
sys.exit(result.returncode)

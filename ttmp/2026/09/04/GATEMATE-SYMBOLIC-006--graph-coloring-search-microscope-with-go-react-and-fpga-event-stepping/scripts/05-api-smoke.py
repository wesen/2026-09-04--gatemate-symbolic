#!/usr/bin/env python3
"""Exercise the Go HTTP API against its explicitly reported execution source."""
from pathlib import Path
import argparse
import json
import time
import urllib.request
import urllib.error

p=argparse.ArgumentParser();p.add_argument('--base',default='http://127.0.0.1:8086');p.add_argument('--engine',default='serial');a=p.parse_args()
T=Path(__file__).resolve().parents[1]
def request(path,data=None):
    payload=None if data is None else json.dumps(data).encode()
    req=urllib.request.Request(a.base+path,data=payload,headers={'Content-Type':'application/json'})
    with urllib.request.urlopen(req,timeout=5) as response:return json.load(response)
state=request('/api/state');assert state['engine']==a.engine,state
if state['running']:request('/api/control',{'action':'pause'})
try:
    request('/api/graph',{'vertices':3,'colors':3,'edges':[[0,1,2]]})
    raise AssertionError('Malformed edge accepted')
except urllib.error.HTTPError as error:assert error.code==400
results=[]
for name,graph,want in [
    ('triangle',dict(vertices=3,colors=3,edges=[[0,1],[1,2],[0,2]],firstOnly=False),6),
    ('unsatisfiable',dict(vertices=3,colors=2,edges=[[0,1],[1,2],[0,2]],firstOnly=False),0),
    ('path',dict(vertices=4,colors=2,edges=[[0,1],[1,2],[2,3]],firstOnly=False),2),
    ('first',dict(vertices=3,colors=3,edges=[[0,1],[1,2],[0,2]],firstOnly=True),1),
]:
    state=request('/api/graph',graph);assert state['latest']['sequence']==0
    state=request('/api/control',{'action':'step'});assert state['latest']['sequence']==1
    first=request(f"/api/events/1?generation={state['generation']}");assert first['sequence']==1
    request('/api/control',{'action':'run'})
    deadline=time.monotonic()+10
    while time.monotonic()<deadline:
        state=request('/api/state')
        assert not state['error'],state
        if state['latest']['kind']==10:break
        time.sleep(.04)
    assert state['latest']['kind']==10 and state['latest']['count']==want,state
    results.append(dict(graph=name,engine=state['engine'],events=state['latest']['sequence'],solutions=state['latest']['count'],generation=state['generation']))
(T/'reference/validation'/f'api-{a.engine}-smoke.json').write_text(json.dumps(results,indent=2)+'\n')
print(json.dumps(results,indent=2))

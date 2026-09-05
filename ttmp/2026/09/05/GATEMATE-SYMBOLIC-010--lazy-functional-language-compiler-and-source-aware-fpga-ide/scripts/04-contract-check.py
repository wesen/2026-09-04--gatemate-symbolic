#!/usr/bin/env python3
"""Check proposed bit layouts and serialize a hand-derived example; not a runtime."""
from pathlib import Path
import json,hashlib
root=Path(__file__).resolve().parents[1]
layouts={
 'object': [('tag',8),('flags',8),('a',16),('b',16),('payload',32)],
 'code':[('op',8),('flags',8),('a',16),('b',16),('c',16),('immediate',32),('span',16),('reserved',16)],
 'frame':[('kind',8),('op',8),('a',16),('b',16),('c',16),('d',16),('savedRef',16),('span',16),('reserved',16)],
 'trace':[('cycle',32),('address',16),('span',16),('kind',8),('flags',8),('reserved',16),('old',80),('new',80)],
 'capability':[('signature',32),('version',8),('heapCapacity',16),('stackCapacity',16),('codeCapacity',16),('traceCapacity',16),('reserved',24)]}
widths={'object':80,'code':128,'frame':128,'trace':256,'capability':128}
def pack(layout_name,**fields):
 value=0
 assert set(fields)<=dict(layouts[layout_name]).keys()
 for name,width in layouts[layout_name]:
  v=fields.get(name,0);assert 0<=v<1<<width,(name,v)
  value=(value<<width)|v
 return value
for kind,fields in layouts.items():
 assert sum(w for _,w in fields)==widths[kind]
 for name,width in fields:
  v=pack(kind,**{name:(1<<width)-1})
  assert v.bit_count()==width

def request(op,payload=b''):
 checksum=0
 for b in payload:checksum^=b
 return op+(payload.hex()+f'{checksum:02x}' if payload else '')+'\n'
# Inline version of the shared milestone, manually lowered in postorder.
code=[
 dict(op=1),dict(op=0,a=14),dict(op=6,a=0,b=1,immediate=2),dict(op=2,a=2),
 dict(op=0,a=13),dict(op=3,a=3,b=4),dict(op=1),dict(op=1),dict(op=6,a=6,b=7),
 dict(op=1),dict(op=1),dict(op=6,a=9,b=10),dict(op=6,a=8,b=11),dict(op=4,a=5,b=12)]
objects=[dict(tag=13,payload=n) for n in range(1,11)]
objects += [dict(tag=1,payload=0),dict(tag=1,payload=1),dict(tag=2),dict(tag=0,payload=21),dict(tag=0,payload=2),dict(tag=5,a=13,b=16),dict(tag=8,a=15,b=65535)]
packed_code=[f'{pack("code",**c,span=i+1):032x}' for i,c in enumerate(code)]
packed_objects=[f'{pack("object",**o):020x}' for o in objects]
assert len(code)==14 and len(objects)==17
for c in code:
 if c['op'] in [2,3,4,6]:
  assert c['a']<len(code)
  if c['op']!=2:assert c['b']<len(code)
 if c['op']==0:assert c['a']<15
fixture={'status':'hand-derived design fixture, not compiler output','source':'def main : Int = let x : Int = (fun (n : Int) -> n * 2) 21 in (x + x) + (x + x);','root':15,'constantEnd':15,'code':packed_code,'objects':packed_objects,'expected':{'result':168,'muls':1,'adds':3,'claims':3,'updates':3,'runtimeAllocations':9,'resultRef':25},'provenance_note':'Spans 1..14 are symbolic fixture IDs, not source byte ranges.'}
fixture_bytes=json.dumps(fixture,indent=2)+'\n'
(root/'sources/shared-design-fixture.json').write_text(fixture_bytes)
profile={'heap':{'count':2048,'width':80},'code':{'count':2048,'width':128},'stack':{'count':512,'width':128},'provenance':{'count':2048,'width':16},'trace':{'count':64,'width':256}}
bits=sum(v['count']*v['width'] for v in profile.values());assert bits==540672
begin=b''.join(n.to_bytes(2,'big') for n in [14,17,15,15])
force=(15).to_bytes(2,'big')
report={'checked':'field widths, bounds, golden framing and illustrative static references only','widths':widths,'logicalMemoryBits':bits,'profile':profile,'golden':{'begin':request('B',begin),'force':request('F',force),'heapRoot':request('H',(15).to_bytes(2,'big')+bytes.fromhex(packed_objects[15])),'code0':packed_code[0]},'fixtureSha256':hashlib.sha256(fixture_bytes.encode()).hexdigest()}
(root/'sources/format-schema.json').write_text(json.dumps({'layouts':layouts,'widths':widths,'profile':profile},indent=2)+'\n')
print(json.dumps(report,indent=2))

#!/usr/bin/env python3
"""Archive reproducible qualification evidence and summarize physical counters."""
from pathlib import Path
import gzip,json
root=Path(__file__).resolve().parents[1]
validation=root/'reference/validation'
for source,target in [(Path('elastic_dataflow/build/nextpnr.log'),'P4-route-final.log.gz'),(Path('elastic_dataflow/build/yosys.log'),'P4-synthesis-final.log.gz'),(Path('/tmp/dataflow-physical.log'),'P4-physical-suite.log')]:
 if target.endswith('.gz'):(validation/target).write_bytes(gzip.compress(source.read_bytes(),mtime=0))
 else:(validation/target).write_bytes(source.read_bytes())
stress=validation/'rtl-stress';stress.mkdir(exist_ok=True)
for p in Path('elastic_dataflow/build/stress').glob('*.log'):(stress/(p.name+'.gz')).write_bytes(gzip.compress(p.read_bytes(),mtime=0))
summary=[]
for name in ['book','copy','fault','cancel']:
 result=json.loads((validation/f'P4-{name}-physical.json').read_text())[0]
 summary.append({'example':name,'source':result['source'],'outputs':result['outputs'],'counters':result['snapshot']['counters'],'config':result['snapshot']['config']})
(validation/'P4-physical-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps(summary,indent=2))

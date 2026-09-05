#!/usr/bin/env python3
from pathlib import Path
import hashlib,json,re,shutil,sys,yaml
root=Path(__file__).resolve().parents[1]
body=(root/'sources/project-report.md').read_text()
doc=root/'reference/03-project-report-inside-the-physically-qualified-lazy-graph-reducer.md'
header=doc.read_text().split('---',2)[1]
doc.write_text('---'+header+'---\n\n'+body)
assert body.count('\n'+chr(96)*3)%2==0
for target in re.findall(r'!\[[^]]*\]\(([^)]+)\)',body):
    assert (doc.parent/target).is_file(),target
if '--publish' not in sys.argv:
    print(json.dumps({'words':len(body.split()),'screenshots':5,'ticket_report':str(doc)}))
    sys.exit(0)
vault=Path('/home/manuel/code/wesen/go-go-golems/go-go-parc')
dated=vault/'Projects/2026/09/05'
name='ARTICLE - GateMate Symbolic - Inside a Lazy Graph Reducer.md'
dest=dated/name
assert not dest.exists(),'append-only: report already exists'
assets=dated/'_assets'
assets.mkdir(parents=True,exist_ok=True)
receipt={'note':str(dest),'assets':{}}
for target in re.findall(r'!\[[^]]*\]\(([^)]+)\)',body):
    src=doc.parent/target
    asset=assets/('gatemate-lazy-'+src.name)
    assert not asset.exists(),'append-only: asset already exists'
    shutil.copyfile(src,asset)
    body=body.replace('('+target+')','(_assets/'+asset.name+')')
    receipt['assets'][str(asset)]=hashlib.sha256(asset.read_bytes()).hexdigest()
meta={'title':'GateMate Symbolic — Inside a Lazy Graph Reducer','aliases':['GateMate Lab 4 lazy reducer report'],'tags':['article','fpga','gatemate','lazy-evaluation','go'],'status':'complete','type':'article','created':'2026-09-05','project_started':'2026-09-04','repo':'/home/manuel/code/wesen/2026-09-04--gatemate-symbolic','source_commit':'79f28e6','ticket':'GATEMATE-SYMBOLIC-009'}
dest.write_text('---\n'+yaml.safe_dump(meta,sort_keys=False,allow_unicode=True)+'---\n\n'+body)
receipt['sha256']=hashlib.sha256(dest.read_bytes()).hexdigest()
(root/'reference/validation/report-vault-publication.json').write_text(json.dumps(receipt,indent=2)+'\n')
print(json.dumps(receipt,indent=2))

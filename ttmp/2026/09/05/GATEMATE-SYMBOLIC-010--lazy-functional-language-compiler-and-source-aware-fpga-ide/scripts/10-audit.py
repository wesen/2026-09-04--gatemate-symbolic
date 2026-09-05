#!/usr/bin/env python3
from pathlib import Path
import hashlib,json,re,struct,subprocess,yaml
root=Path(__file__).resolve().parents[1]
old=root.parent/'GATEMATE-SYMBOLIC-009--laboratory-4-lazy-graph-reducer-and-heap-inspector'
report={}
for p in [root/'index.md',*root.glob('design-doc/*.md'),*root.glob('reference/*.md')]:
 s=p.read_text();meta=yaml.safe_load(s.split('---',2)[1]);assert meta['Ticket']=='GATEMATE-SYMBOLIC-010',p
 assert s.count('\n'+chr(96)*3)%2==0,p
 for target in re.findall(r'\]\(([^)]+)\)',s):
  if '://' not in target and not target.startswith('#'):assert (p.parent/target.split('#')[0]).exists(),(p,target)
 report[p.name]={'words':len(s.split()),'sha256':hashlib.sha256(p.read_bytes()).hexdigest()}
fixture=json.loads((root/'sources/shared-design-fixture.json').read_text())
assert all(len(v)==20 for v in fixture['objects'])
assert all(len(v)==32 for v in fixture['code'])
assert json.loads((root/'reference/validation/contract-check.json').read_text())['logicalMemoryBits']==540672
figures={}
for p in (root/'reference/figures').glob('*.png'):
 data=p.read_bytes();assert data[:8]==b'\x89PNG\r\n\x1a\n'
 figures[p.name]=struct.unpack('>II',data[16:24])
assert len(figures)==7
receipt=json.loads((old/'reference/validation/report-vault-publication.json').read_text())
assert hashlib.sha256(Path(receipt['note']).read_bytes()).hexdigest()==receipt['sha256']
for path,digest in receipt['assets'].items():assert hashlib.sha256(Path(path).read_bytes()).hexdigest()==digest
pdf=root/'reference/validation/GATEMATE 010 Lazy Functional Language Design.pdf'
assert pdf.read_bytes().startswith(b'%PDF-')
pdftext=subprocess.check_output(['pdftotext','-layout',str(pdf),'-'],text=True)
(root/'reference/validation/design-pdf.txt').write_text(pdftext)
pages=pdftext.split('\f');pages=[p for p in pages if p.strip()]
assert len(pages)>10
for required in ['A hand-derived compiled example','Observation pages and counter numbering','540,672','CYCLIC_THUNK']:
 assert required in pdftext,required
figure_page=next(i+1 for i,p in enumerate(pages) if 'Compiler pipeline from source to a validated artifact' in p)
subprocess.run(['pdftoppm','-f',str(figure_page),'-l',str(figure_page),'-scale-to','1400','-png','-singlefile',str(pdf),str(root/'reference/validation/design-figure-page')],check=True)
slips={}
for name in ['plan','d1-start','d1-done','d2-start','d2-done','d3-start','d3-done']:
 log=root/f'reference/validation/{name}-print.log'
 slips[name]='printed' if log.exists() and 'printed: true' in log.read_text() else 'pending'
upload=root/'reference/validation/remarkable-upload.log'
out={'documents':report,'figures':figures,'pdfPages':len(pages),'inspectedFigurePage':figure_page,'vaultArticleAndAssetsMatch':True,'physicalSlips':slips,'remarkableUploaded':upload.exists() and 'OK: uploaded' in upload.read_text()}
print(json.dumps(out,indent=2))

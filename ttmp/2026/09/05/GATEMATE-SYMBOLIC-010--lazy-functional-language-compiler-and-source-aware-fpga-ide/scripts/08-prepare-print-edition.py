#!/usr/bin/env python3
from pathlib import Path
import re,yaml
root=Path(__file__).resolve().parents[1]
p=root/'sources/design.md';body=p.read_text()
body=body.replace('flowchart LR\n    S[Source text]','flowchart TD\n    S[Source text]')
p.write_text(body)
doc=root/'design-doc/01-lazy-functional-language-intern-analysis-design-and-implementation-guide.md'
header=doc.read_text().split('---',2)[1]
doc.write_text('---'+header+'---\n\n'+body)
i=0
captions=['Compiler pipeline from source to a validated artifact','Evaluation and thunk update transitions','Source-aware IDE and runtime observation architecture']
def diagram(match):
 global i
 i+=1
 size='height=5.5in' if i==1 else 'width=95%'
 return f'![{captions[i-1]}](../reference/figures/design-{i}.png)'+'{'+size+'}'
print_body=re.sub(chr(96)*3+r'mermaid\n.*?\n'+chr(96)*3,diagram,body,flags=re.S)
assert i==3
meta={'Title':'Lazy Functional Language - Intern Design and Implementation Guide','Ticket':'GATEMATE-SYMBOLIC-010','DocType':'design-doc','Status':'review'}
(root/'sources/design-print.md').write_text('---\n'+yaml.safe_dump(meta,sort_keys=False)+'---\n\n'+print_body)
# Correct the investigation's earlier diagram count after rendering all blocks.
old=root.parent/'GATEMATE-SYMBOLIC-009--laboratory-4-lazy-graph-reducer-and-heap-inspector/reference/01-implementation-diary.md'
s=old.read_text().replace('three diagrams, representation and API tables','four diagrams, representation and API tables');old.write_text(s)
p=root/'reference/01-implementation-diary.md';s=p.read_text().replace('all fault codes, three diagrams and five physical screenshots','all fault codes, four diagrams and five physical screenshots');p.write_text(s)
p=root/'sources/d1-diary.json';s=p.read_text().replace('all fault codes, three diagrams and five physical screenshots','all fault codes, four diagrams and five physical screenshots');p.write_text(s)

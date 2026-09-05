#!/usr/bin/env python3
"""Write the initial design and append strict diary steps from JSON."""
from pathlib import Path
import json,sys
root=Path(__file__).resolve().parents[1]
if sys.argv[1]=='diary':
 d=json.loads(Path(sys.argv[2]).read_text());p=root/'reference/01-implementation-diary.md';s=p.read_text()
 if '## Step ' not in s:s=s.split('---',2)[0]+'---'+s.split('---',2)[1]+'---\n\n# Implementation diary\n\n## Goal\n\nDocument the lazy reducer, validation, physical evidence, and phased delivery.\n'
 n=s.count('\n## Step ')+1
 prompt='See Step 1.' if n>1 else (root/'sources/user-request.txt').read_text()
 s+=f'\n## Step {n}: {d["title"]}\n\n{d["prose"]}\n\n### Prompt Context\n\n**User prompt (verbatim):** {prompt}\n\n**Assistant interpretation:** Implement Lab 4 with design, model, RTL, host inspector, physical qualification, diary and printed phases.\n\n**Inferred user intent:** Build an inspectable lazy machine with evidence that sharing evaluates once and cycles terminate precisely.\n'
 if d.get('commit'):s+='\n**Commit (code):** '+d['commit']+'\n'
 for k,h in [('did','What I did'),('why','Why'),('worked','What worked'),('failed',"What didn't work"),('learned','What I learned'),('tricky','What was tricky to build'),('review','What warrants a second pair of eyes'),('future','What should be done in the future'),('validate','Code review instructions'),('details','Technical details')]:s+='\n### '+h+'\n\n'+'\n'.join('- '+x for x in d.get(k,['N/A']))+'\n'
 p.write_text(s)
else:
 p=root/'design-doc/01-lazy-graph-reducer-intern-analysis-design-and-implementation-guide.md';s=p.read_text();p.write_text(s.split('---',2)[0]+'---'+s.split('---',2)[1]+'---\n\n'+(root/'sources/design-body.md').read_text())

#!/usr/bin/env python3
from pathlib import Path
import json,sys
root=Path(__file__).resolve().parents[1]
d=json.loads((root/sys.argv[1]).read_text());p=root/'reference/01-implementation-diary.md';s=p.read_text();n=s.count('\n## Step ')+1
s+=f'\n## Step {n}: {d["title"]}\n\n{d["prose"]}\n\n### Prompt Context\n\n**User prompt (verbatim):** {d.get("prompt","(see Step 5)")}\n\n**Assistant interpretation:** Implement the lexer and handwritten parser described in the preceding exchange, with tested diagnostics, phase slips and commits.\n\n**Inferred user intent:** Turn the agreed parsing design into a usable, documented Go syntax package as the first part of I1.\n'
if d.get('commit'):s+='\n**Commit (code):** '+d['commit']+'\n'
for k,h in [('did','What I did'),('why','Why'),('worked','What worked'),('failed',"What didn't work"),('learned','What I learned'),('tricky','What was tricky to build'),('review','What warrants a second pair of eyes'),('future','What should be done in the future'),('validate','Code review instructions'),('details','Technical details')]:s+='\n### '+h+'\n\n'+'\n'.join('- '+x for x in d.get(k,['N/A']))+'\n'
p.write_text(s)

#!/usr/bin/env python3
from pathlib import Path
import json,sys
root=Path(__file__).resolve().parents[1]
mode=sys.argv[1]
if mode=='body':
 p=root/sys.argv[2];body=(root/sys.argv[3]).read_text()
 p.write_text('---'+p.read_text().split('---',2)[1]+'---\n\n'+body)
elif mode=='diary':
 d=json.loads((root/sys.argv[2]).read_text());p=root/'reference/01-implementation-diary.md';s=p.read_text()
 if '## Step ' not in s:s='---'+s.split('---',2)[1]+'---\n\n# Implementation diary\n\n## Goal\n\nRecord the report delivery and design of a lazy functional language, with exact evidence, decisions, failures and future implementation gates.\n'
 n=s.count('\n## Step ')+1
 prompt=(root/'sources/user-request.txt').read_text().rstrip() if n==1 else '(see Step 1)'
 s+=f'\n## Step {n}: {d["title"]}\n\n{d["prose"]}\n\n### Prompt Context\n\n**User prompt (verbatim):**\n\n{prompt}\n\n**Assistant interpretation:** Publish the completed reducer report, then prepare and deliver the follow-up language design with a detailed implementation plan.\n\n**Inferred user intent:** Preserve a technical explanation of the working machine and make the next project concrete enough to implement task by task.\n'
 if d.get('commit'):s+='\n**Commit (documentation):** '+d['commit']+'\n'
 for k,h in [('did','What I did'),('why','Why'),('worked','What worked'),('failed',"What didn't work"),('learned','What I learned'),('tricky','What was tricky to build'),('review','What warrants a second pair of eyes'),('future','What should be done in the future'),('validate','Code review instructions'),('details','Technical details')]:s+='\n### '+h+'\n\n'+'\n'.join('- '+x for x in d.get(k,['N/A']))+'\n'
 p.write_text(s)

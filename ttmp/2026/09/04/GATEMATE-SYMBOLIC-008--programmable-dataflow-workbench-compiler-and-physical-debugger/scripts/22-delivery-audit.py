#!/usr/bin/env python3
"""Check ticket delivery artifacts without altering live hardware or the UI."""
from pathlib import Path
import hashlib,json,re
import yaml
root=Path(__file__).resolve().parents[1]
slips=list((root/'reference/validation').glob('*-print.log'))
for path in slips:
 assert 'printed: true' in path.read_text(),path
images=list((root/'reference/screenshots').glob('*.png'))
for path in images:assert path.read_bytes().startswith(b'\x89PNG\r\n\x1a\n'),path
for path in list((root/'design-doc').glob('*.md'))+list((root/'reference').glob('*.md')):
 text=path.read_text()
 assert text.count('\n```')%2==0,path
 yaml.safe_load(text.split('---',2)[1])
 for target in re.findall(r'!\[[^]]*\]\(([^)]+)\)',text):assert (path.parent/target).is_file(),(path,target)
assert 'OK: uploaded' in (root/'reference/validation/p1-upload.log').read_text()
print(json.dumps({'printed_slips':len(slips),'screenshots':{p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in images},'design_uploaded':True},indent=2))

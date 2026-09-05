#!/usr/bin/env python3
from pathlib import Path
import re,json,hashlib,yaml
root=Path(__file__).resolve().parents[1]
for p in list((root/'design-doc').glob('*.md'))+list((root/'reference').glob('*.md')):
 s=p.read_text();yaml.safe_load(s.split('---',2)[1]);assert s.count('\n```')%2==0,p
 for target in re.findall(r'!\[[^]]*\]\(([^)]+)\)',s):assert (p.parent/target).is_file(),target
slips=list((root/'reference/validation').glob('*-print.log'))
for p in slips:assert 'printed: true' in p.read_text(),p
assert 'OK: uploaded' in (root/'reference/validation/p1-upload.log').read_text()
assert 'OK: uploaded' in (root/'reference/validation/p5-handoff-upload.log').read_text()
images=list((root/'reference/screenshots').glob('*.png'))
for p in images:assert p.read_bytes().startswith(b'\x89PNG\r\n\x1a\n'),p
print(json.dumps({'printed_slips':len(slips),'physical_qualification':'pending device connection','images':{p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in images},'design_uploaded':True,'handoff_uploaded':True},indent=2))

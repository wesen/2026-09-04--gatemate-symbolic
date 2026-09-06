from pathlib import Path
import json,hashlib,subprocess
T=Path(__file__).resolve().parents[1]
vault=Path('/home/manuel/code/wesen/go-go-golems/go-go-parc')
m=json.loads((T/'reference/validation/project-report-manifest.json').read_text())
def git(*args): return subprocess.check_output(['git','-C',str(vault),*args],text=True).strip()
assert not git('diff','--cached','--name-only'),'Existing staged changes; refusing mixed commit'
paths=[]
for item in m['files']:
    p=vault/item['path']
    assert hashlib.sha256(p.read_bytes()).hexdigest()==item['sha256'],str(p)
    paths.append(item['path'])
print(git('add','--',*paths))
print(git('diff','--cached','--check'))
print(git('diff','--cached','--stat'))
assert set(git('diff','--cached','--name-only').splitlines())==set(paths)
print(git('commit','-m','docs: explain LFL1 lazy language execution on GateMate'))
print(git('rev-parse','HEAD'))

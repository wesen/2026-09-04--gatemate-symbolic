from pathlib import Path
import re, argparse
T=Path(__file__).resolve().parents[1]
vault=Path('/home/manuel/code/wesen/go-go-golems/go-go-parc/Projects/2026/09/05/ARTICLE - GateMate Symbolic - Inside a Lazy Functional Language.md')
def fix(text):
    def block(m):
        s=m[1].replace('flowchart TD','graph TD').replace('flowchart LR','graph LR')
        s=re.sub(r'\[([^"\]\n][^\]\n]*)\]',lambda x:'["'+x[1]+'"]',s)
        s=re.sub(r'\|([^"|\n][^|\n]*)\|',lambda x:'|"'+x[1]+'"|',s)
        return '```mermaid\n'+s+'```'
    return re.sub(r'```mermaid\n(.*?)```',block,text,flags=re.S)
p=argparse.ArgumentParser();p.add_argument('--vault',action='store_true');a=p.parse_args()
paths=[vault] if a.vault else [T/'sources/lfl1-project-report.md',T/'scripts/29-write-project-report.py']
for path in paths:
    before=path.read_text();after=fix(before)
    assert before!=after, str(path)
    path.write_text(after)
    print(path)

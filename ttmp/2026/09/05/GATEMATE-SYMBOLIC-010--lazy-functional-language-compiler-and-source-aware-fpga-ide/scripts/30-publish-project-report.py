from pathlib import Path
import re, json, shutil, hashlib, argparse
T=Path(__file__).resolve().parents[1]
root=T.parents[4]
vault=Path('/home/manuel/code/wesen/go-go-golems/go-go-parc')
dest=vault/'Projects/2026/09/05/ARTICLE - GateMate Symbolic - Inside a Lazy Functional Language.md'
source=T/'sources/lfl1-project-report.md'
body=source.read_text()
assets=re.findall(r'!\[[^\]]*\]\(_assets/([^)]*)\)',body)
files=[]
for name in assets:
    matches=list((T/'reference/screenshots').glob('*-'+name.removeprefix('lfl1-')))
    assert len(matches)==1, (name,matches)
    files.append((matches[0],dest.parent/'_assets'/name))
for path in re.findall(r'https://github.com/wesen/2026-09-04--gatemate-symbolic/blob/[a-f0-9]+/([^)]*)',body):
    assert (root/path).is_file(), (root,path)
for name in re.findall(r'\[\[([^]]+)\]\]',body):
    assert list(vault.rglob(name+'.md')),name
assert body.count('```')%2==0
mermaid=re.findall(r'```mermaid\n(.*?)```',body,re.S)
js='''async (existing) => {
 const page = await existing.context().newPage();
 try {
 await page.goto('about:blank');
 await page.addScriptTag({path:'/home/manuel/.local/share/nvim/lazy/markdown-preview.nvim/app/_static/mermaid.min.js'});
 const sources = SOURCES;
 return await page.evaluate(async (sources) => {
 window.mermaid.initialize({startOnLoad:false,securityLevel:'strict'});
 const results=[];
 for(let i=0;i<sources.length;i++) {
 const result=await window.mermaid.render('article'+i,sources[i]);
 const svg=typeof result==='string'?result:result.svg;
 if(!svg.includes('<svg')) throw new Error('Missing SVG');
 results.push({diagram:i+1,rendered:true});
 }
 return results;
 },sources);
 } finally { await page.close(); }
}'''.replace('SOURCES',json.dumps(mermaid))
(T/'scripts/31-validate-report-diagrams.js').write_text(js)
parser=argparse.ArgumentParser();parser.add_argument('--publish',action='store_true');args=parser.parse_args()
if args.publish:
    for a,b in [(source,dest)]+files:
        b.parent.mkdir(parents=True,exist_ok=True)
        if b.exists(): assert b.read_bytes()==a.read_bytes(),str(b)
        else: shutil.copyfile(a,b)
manifest={'words':len(body.split()),'diagrams':len(mermaid),'published':args.publish,'article':str(dest),'files':[{'path':str(b.relative_to(vault)),'sha256':hashlib.sha256(a.read_bytes()).hexdigest()} for a,b in [(source,dest)]+files]}
(T/'reference/validation/project-report-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
print(json.dumps(manifest,indent=2))

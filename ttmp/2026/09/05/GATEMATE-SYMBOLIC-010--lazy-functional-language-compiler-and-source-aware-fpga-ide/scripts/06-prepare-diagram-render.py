#!/usr/bin/env python3
from pathlib import Path
import json,re
root=Path(__file__).resolve().parents[1]
old=root.parent/'GATEMATE-SYMBOLIC-009--laboratory-4-lazy-graph-reducer-and-heap-inspector'
items=[]
for kind,path in [('design',root/'sources/design.md'),('report',old/'sources/project-report.md')]:
 for i,source in enumerate(re.findall(chr(96)*3+r'mermaid\n(.*?)\n'+chr(96)*3,path.read_text(),re.S),1):
  items.append({'name':f'{kind}-{i}','source':source})
(root/'reference/figures').mkdir(exist_ok=True)
code="""async (existing) => {
 const page = await existing.context().newPage();
 const data = DATA;
 const dir = DIR;
 await page.goto('about:blank');
 await page.setViewportSize({width:1500,height:1200});
 await page.addScriptTag({path:'/home/manuel/.local/share/nvim/lazy/markdown-preview.nvim/app/_static/mermaid.min.js'});
 const results=[];
 for (const item of data) {
  await page.evaluate(async ({name,source})=>{
   document.body.innerHTML='<div id="figure"></div>';
   document.body.style.cssText='margin:24px;background:white';
   window.mermaid.initialize({startOnLoad:false,theme:'neutral',securityLevel:'strict'});
   const result=await window.mermaid.render(name,source);
   document.getElementById('figure').innerHTML=typeof result==='string'?result:result.svg;
  },item);
  const figure=page.locator('#figure svg');
  await figure.waitFor();
  await figure.screenshot({path:dir+'/'+item.name+'.png'});
  results.push({name:item.name,svg:true,bounds:await figure.boundingBox()});
 }
 await page.close();
 return results;
}
""".replace('DATA',json.dumps(items)).replace('DIR',json.dumps(str(root/'reference/figures')))
(root/'scripts/07-render-diagrams.js').write_text(code)
print(json.dumps({'diagrams':len(items)}))

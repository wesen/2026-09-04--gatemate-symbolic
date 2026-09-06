async(page)=>{
 const base='/home/manuel/code/wesen/2026-09-04--gatemate-symbolic/ttmp/2026/09/05/GATEMATE-SYMBOLIC-010--lazy-functional-language-compiler-and-source-aware-fpga-ide/reference/screenshots';
 await page.goto('http://127.0.0.1:18091');await page.setViewportSize({width:1600,height:1100});
 await page.waitForFunction(()=>document.querySelector('textarea')?.value.length>0);
 await page.getByRole('button',{name:'Compile',exact:true}).click();await page.getByText('✓ Checked and compiled',{exact:false}).waitFor();
 await page.getByRole('button',{name:'Load compiled artifact',exact:true}).click();await page.waitForFunction(()=>document.querySelector('.state-badge')?.textContent==='IDLE');
 await page.getByRole('button',{name:'Force main',exact:true}).click();
 for(let i=0;i<3;i++)await page.getByRole('button',{name:'Step 1',exact:true}).click();
 await page.getByLabel('Heap filter').selectOption('mutable');await page.locator('.heap-scroll tbody tr').filter({hasText:'BLACKHOLE'}).first().click();
 await page.getByRole('button',{name:'Stack',exact:true}).click();
 await page.screenshot({path:base+'/09-model-claimed-thunk-and-update-frame.png',fullPage:true});
 const options=await page.getByLabel('History frame').locator('option').evaluateAll(nodes=>nodes.map(n=>({value:n.value,text:n.textContent})));
 const load=options.find(o=>o.text.includes('load'));if(!load)throw new Error('missing load history');
 await page.getByLabel('History frame').selectOption(load.value);await page.getByText('Historical frame is read-only.',{exact:false}).waitFor();
 if(await page.getByRole('button',{name:'Step 1',exact:true}).isEnabled())throw new Error('historical control enabled');
 await page.screenshot({path:base+'/10-model-read-only-history.png',fullPage:true});
 await page.getByLabel('History frame').selectOption('0');
 await page.getByLabel('Language source').fill('def main : Int = true;');
 if(await page.getByRole('button',{name:'Load compiled artifact',exact:true}).isEnabled())throw new Error('stale compilation load enabled');
 await page.getByRole('button',{name:'Compile',exact:true}).click();await page.locator('.diagnostic').waitFor();
 await page.screenshot({path:base+'/11-model-type-diagnostic.png',fullPage:true});
 return {screenshots:3,historyReadOnly:true,staleLoadBlocked:true,diagnostic:await page.locator('.diagnostic').innerText()};
}

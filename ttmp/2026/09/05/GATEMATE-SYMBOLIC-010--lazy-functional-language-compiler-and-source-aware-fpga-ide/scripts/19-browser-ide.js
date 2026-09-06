async (page) => {
 const base='/home/manuel/code/wesen/2026-09-04--gatemate-symbolic/ttmp/2026/09/05/GATEMATE-SYMBOLIC-010--lazy-functional-language-compiler-and-source-aware-fpga-ide/reference/screenshots';
 await page.setViewportSize({width:1600,height:1100});
 await page.goto('http://127.0.0.1:18091');
 await page.waitForFunction(()=>document.querySelector('textarea')?.value.length>0);
 await page.getByRole('button',{name:'Compile',exact:true}).click();
 await page.getByText('✓ Checked and compiled',{exact:false}).waitFor();
 await page.getByRole('button',{name:'Load compiled artifact',exact:true}).click();
 await page.locator('.heap-scroll tbody tr').first().waitFor();
 await page.getByLabel('Heap filter').selectOption('mutable');
 await page.screenshot({path:base+'/01-model-loaded-source.png',fullPage:true});
 await page.getByRole('button',{name:'Force main',exact:true}).click();
 await page.getByRole('button',{name:'Run 1000 ticks',exact:true}).click();
 await page.getByRole('button',{name:'Poll result',exact:true}).waitFor({state:'visible'});
 await page.waitForFunction(()=>document.querySelector('.state-badge')?.textContent?.includes('OUTPUT'));
 await page.locator('.result-strip button').click();
 await page.getByRole('button',{name:'Mutations',exact:true}).click();
 await page.screenshot({path:base+'/02-model-shared-result-and-updates.png',fullPage:true});
 await page.getByLabel('Example program').selectOption('closure');
 await page.getByRole('button',{name:'Compile',exact:true}).click();
 await page.getByText('✓ Checked and compiled',{exact:false}).waitFor();
 await page.getByRole('button',{name:'Load compiled artifact',exact:true}).click();
 await page.waitForFunction(()=>document.querySelector('.state-badge')?.textContent==='IDLE');
 await page.getByRole('button',{name:'Force main',exact:true}).click();
 await page.getByRole('button',{name:'Run 1000 ticks',exact:true}).click();
 await page.waitForFunction(()=>document.querySelector('.state-badge')?.textContent==='OUTPUT');
 await page.getByLabel('Heap filter').selectOption('FUN');
 await page.locator('.heap-scroll tbody tr').last().click();
 await page.getByRole('button',{name:'Code',exact:true}).click();
 await page.screenshot({path:base+'/03-model-closure-environment.png',fullPage:true});
 await page.getByLabel('Example program').selectOption('squares');
 await page.getByRole('button',{name:'Compile',exact:true}).click();
 await page.getByText('✓ Checked and compiled',{exact:false}).waitFor();
 await page.getByRole('button',{name:'Load compiled artifact',exact:true}).click();
 await page.waitForFunction(()=>document.querySelector('.state-badge')?.textContent==='IDLE');
 await page.getByRole('button',{name:'Start stream',exact:true}).click();
 for(let i=0;i<8;i++){
  await page.getByRole('button',{name:'Next element',exact:true}).click();
  await page.waitForFunction(n=>document.querySelectorAll('.stream-values>span').length===n||[...document.querySelectorAll('button')].some(b=>b.textContent==='Resume demand'&&!b.disabled),i+1);
  while(await page.locator('.stream-values>span').count()<i+1){
   await page.getByRole('button',{name:'Resume demand',exact:true}).click();
   await page.waitForFunction(n=>document.querySelectorAll('.stream-values>span').length===n||[...document.querySelectorAll('button')].some(b=>b.textContent==='Resume demand'&&!b.disabled),i+1);
  }
 }
 await page.getByLabel('Heap filter').selectOption('CONS');
 await page.locator('.heap-scroll tbody tr').first().click();
 await page.screenshot({path:base+'/04-model-eight-square-stream.png',fullPage:true});
 const values=await page.locator('.stream-values>span').allTextContents();
 if(values.join(',')!=='1,4,9,16,25,36,49,64')throw new Error('unexpected stream '+values);
 return {screenshots:4,values,title:await page.title()};
}

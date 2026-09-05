async (page) => {
 const dir='/home/manuel/code/wesen/2026-09-04--gatemate-symbolic/ttmp/2026/09/05/GATEMATE-SYMBOLIC-009--laboratory-4-lazy-graph-reducer-and-heap-inspector/reference/screenshots';
 await page.setViewportSize({width:1800,height:1100});await page.goto('http://127.0.0.1:18089/');
 await page.getByRole('heading',{name:'Lazy Graph Reducer',exact:true}).waitFor();
 const state=()=>page.evaluate(async()=>await(await fetch('/api/lazy/state')).json());
 const change=async(name)=>{const before=await state();await page.getByRole('button',{name,exact:true}).click();await page.waitForFunction(async id=>(await(await fetch('/api/lazy/state')).json()).current.id>id,before.current.id);return state()};
 await change('Load graph');await change('Force root');
 let current=await state();for(let n=0;n<30&&current.current.snapshot.counters.claims===0;n++)current=await change('Step 1 cycle');
 if(current.current.snapshot.counters.claims!==1||current.current.snapshot.counters.updates!==0||current.current.snapshot.heap[3]!==5*2**36)throw Error('missing live thunk claim');
 const claimed=current.current;const prefix=claimed.snapshot.source==='serial'?'fpga':'model';
 await page.screenshot({path:dir+'/'+prefix+'-claimed.png',fullPage:true});
 await page.getByLabel('Cycles to advance').fill('1000');current=await change('Advance');
 if(current.current.snapshot.result!==168||!current.current.snapshot.valid||current.current.snapshot.heap[3]!==42||current.current.snapshot.counters.muls!==1)throw Error('shared result or memoization incorrect');
 const completed=current.current;await page.screenshot({path:dir+'/'+prefix+'-result.png',fullPage:true});
 await change('Poll result');await change('Force root');current=await change('Advance');if(current.current.snapshot.counters.muls!==1||current.current.snapshot.result!==168)throw Error('repeated force recomputed body');
 await page.getByLabel('Snapshot history').selectOption(String(claimed.id));if(await page.getByRole('button',{name:'Force root',exact:true}).isEnabled())throw Error('historical mutation enabled');await page.screenshot({path:dir+'/'+prefix+'-history.png',fullPage:true});await page.getByRole('button',{name:'Return to live',exact:true}).click();
 await page.getByLabel('Example graph').selectOption('cycle');await change('Load graph');await change('Force root');current=await change('Advance');if(current.current.snapshot.result!==13*2**36+3||current.current.snapshot.heap[1]!==13*2**36+3)throw Error('cycle not memoized');
 const cycle=current.current;await page.screenshot({path:dir+'/'+prefix+'-cycle.png',fullPage:true});
 const old=await page.getByLabel('Graph source').inputValue();await page.getByLabel('Graph source').fill('{');await page.getByRole('button',{name:'Load graph',exact:true}).click();await page.getByRole('alert').filter({hasText:'SyntaxError'}).waitFor();if((await state()).current.id!==cycle.id)throw Error('invalid image changed state');await page.getByLabel('Graph source').fill(old);await page.getByLabel('Cycles to advance').fill('1');await change('Advance');
 await page.setViewportSize({width:390,height:844});await page.screenshot({path:dir+'/'+prefix+'-mobile.png',fullPage:true});const width=await page.evaluate(()=>({scroll:document.documentElement.scrollWidth,client:document.documentElement.clientWidth}));if(width.scroll>width.client)throw Error('mobile document overflow');
 await page.setViewportSize({width:1800,height:1100});return {claimed,completed,cycle,width};
}

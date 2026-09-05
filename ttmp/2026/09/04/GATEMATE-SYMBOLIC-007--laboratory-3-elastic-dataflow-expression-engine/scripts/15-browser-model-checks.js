async (page) => {
  const root='/home/manuel/code/wesen/2026-09-04--gatemate-symbolic/ttmp/2026/09/04/GATEMATE-SYMBOLIC-007--laboratory-3-elastic-dataflow-expression-engine';
  const state=()=>page.evaluate(async()=>await(await fetch('/api/dataflow/state')).json());
  const waitStep=async(n)=>page.waitForFunction(async n=>(await(await fetch('/api/dataflow/state')).json()).step===n,n);
  await page.setViewportSize({width:1600,height:1100});
  await page.goto('http://127.0.0.1:8087/');
  await page.getByText('8 actions · valid scenario').waitFor();
  await page.getByRole('button',{name:'Run from start',exact:true}).click();
  await waitStep(8);
  await page.waitForFunction(()=>[...document.querySelectorAll('.df-result-value')].map(e=>e.textContent).join(',')==='58,12');
  await page.screenshot({path:root+'/reference/screenshots/02-model-book-complete.png',fullPage:true});
  await page.getByRole('textbox',{name:'Project ID'}).fill('dataflow-intern-review');
  await page.getByRole('button',{name:'Save',exact:true}).click();
  await page.getByRole('status').filter({hasText:'Saved dataflow-intern-review'}).waitFor();
  const source=await page.getByRole('textbox',{name:'Scenario JSON source'}).inputValue();
  await page.getByRole('textbox',{name:'Scenario JSON source'}).fill('invalid JSON');
  await page.waitForFunction(()=>document.querySelector('.df-validation')?.classList.contains('invalid'));
  if(!await page.getByRole('button',{name:'Run from start',exact:true}).isDisabled())throw Error('Invalid source remained executable');
  await page.getByRole('combobox',{name:'Open saved project'}).selectOption('dataflow-intern-review');
  await page.waitForFunction(source=>document.querySelector('textarea')?.value===source,source);
  const live=await state();const recorded=live.history.find(f=>f.label.includes('tick'));
  await page.getByRole('combobox',{name:'Snapshot history'}).selectOption(String(recorded.id));
  await page.getByText('HISTORICAL SNAPSHOT '+recorded.id).waitFor();
  for(const name of ['Step 1 cycle','Reset engine','Run from start'])if(!await page.getByRole('button',{name,exact:true}).isDisabled())throw Error('Historical mutation enabled: '+name);
  await page.screenshot({path:root+'/reference/screenshots/03-model-history.png',fullPage:true});
  await page.getByRole('button',{name:'Return to live'}).click();
  await page.getByRole('combobox',{name:'EXAMPLE'}).selectOption('cancel');
  await page.getByText('10 actions · valid scenario').waitFor();
  for(let n=1;n<=4;n++){await page.getByRole('button',{name:'Next action',exact:true}).click();await waitStep(n)}
  await page.locator('.df-context').nth(2).click();
  await page.waitForFunction(()=>document.querySelector('.df-inspector')?.textContent.includes('Context 2'));
  const inFlight=await state();if(!inFlight.current.snapshot.mul.some(Boolean))throw Error('No occupied model multiplier');
  await page.screenshot({path:root+'/reference/screenshots/04-model-in-flight.png',fullPage:true});
  for(let n=5;n<=10;n++){await page.getByRole('button',{name:'Next action',exact:true}).click();await waitStep(n)}
  const canceled=await state();if(canceled.results.length!==1||canceled.results[0].value!==12||canceled.results[0].epoch!==1)throw Error('Cancellation result mismatch');
  await page.setViewportSize({width:390,height:844});
  if(await page.evaluate(()=>document.body.scrollWidth>innerWidth))throw Error('Mobile horizontal overflow');
  await page.screenshot({path:root+'/reference/screenshots/05-model-mobile.png',fullPage:false});
  await page.setViewportSize({width:1600,height:1100});
  return {bookResults:live.results,savedProject:'dataflow-intern-review',historicalControlsDisabled:true,inFlight:inFlight.current.snapshot.mul,cancellationResults:canceled.results,mobileOverflow:false};
}

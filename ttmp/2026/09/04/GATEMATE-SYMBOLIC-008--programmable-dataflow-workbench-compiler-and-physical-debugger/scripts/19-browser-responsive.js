async (page) => {
 await page.setViewportSize({width:390,height:844});
 const overflow=await page.evaluate(()=>({viewport:innerWidth,body:document.documentElement.scrollWidth}));
 await page.screenshot({path:'/home/manuel/code/wesen/2026-09-04--gatemate-symbolic/ttmp/2026/09/04/GATEMATE-SYMBOLIC-008--programmable-dataflow-workbench-compiler-and-physical-debugger/reference/screenshots/04-model-mobile.png',fullPage:true});
 if(overflow.body>overflow.viewport)throw new Error(`horizontal overflow: ${JSON.stringify(overflow)}`);
 await page.setViewportSize({width:1800,height:1100});
 return overflow;
}

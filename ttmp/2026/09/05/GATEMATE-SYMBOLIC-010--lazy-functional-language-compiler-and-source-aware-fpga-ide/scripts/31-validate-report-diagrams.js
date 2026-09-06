async (existing) => {
 const page = await existing.context().newPage();
 try {
 await page.goto('about:blank');
 await page.addScriptTag({path:'/home/manuel/.local/share/nvim/lazy/markdown-preview.nvim/app/_static/mermaid.min.js'});
 const sources = ["flowchart TD\n  S[Source text] --> P[Lexer and parser]\n  P --> C[Type and lexical checker]\n  C --> R[AST reference evaluator]\n  C --> A[Compiler and immutable artifact]\n  A --> G[Finite Go machine]\n  A --> U[UART loader and client]\n  U --> F[GateMate FPGA machine]\n  G --> H[Go session and HTTP API]\n  F --> U\n  U --> H\n  H --> I[React source and machine inspector]\n", "flowchart LR\n  T[THUNK: code and environment] -->|claim and push UPDATE| B[BLACKHOLE]\n  B -->|body returns a value reference| I[IND: result reference]\n  B -->|recursive demand while evaluating| E[Cycle error]\n  E -->|unwind pending updates| I\n"];
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
}
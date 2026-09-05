async (existing) => {
 const page = await existing.context().newPage();
 const data = [{"name": "design-1", "source": "flowchart TD\n    S[Source text] --> T[Tokens and source spans]\n    T --> A[Parsed syntax tree]\n    A --> B[Resolved bindings and types]\n    B --> C[Expression code records]\n    C --> H[Initial heap and provenance]\n    H --> P[Validated immutable artifact]\n    P --> M[Go machine]\n    P --> F[FPGA loader]"}, {"name": "design-2", "source": "flowchart TD\n    E[Eval code with environment] -->|variable| L[Resolve ENV binding]\n    L --> N[Enter heap reference]\n    E -->|lambda or constructor| A[Allocate complete objects]\n    A --> R[Return WHNF reference]\n    N -->|THUNK| U[Reserve UPDATE and claim]\n    U --> E\n    N -->|IND| N\n    N -->|value or ERROR| R\n    R --> K[Dispatch continuation]\n    K -->|body or selected branch| E\n    K -->|UPDATE| W[BLACKHOLE to IND]\n    W --> R\n    K -->|empty stack| O[Hold result reference]"}, {"name": "design-3", "source": "flowchart TD\n    ED[Source editor and diagnostics] --> CP[Compile API]\n    CP --> AR[Immutable artifact]\n    AR --> LV[Typed code and binding viewer]\n    AR --> LD[Explicit load control]\n    LD --> SS[Serialized session]\n    SS --> RT[Model or serial Runtime]\n    RT --> SN[Detached snapshot]\n    SN --> HP[Heap and closure inspector]\n    SN --> ST[Continuation and allocation views]\n    SN --> TR[Source-linked mutation timeline]\n    SN --> LS[Lazy list demand panel]"}, {"name": "report-1", "source": "flowchart LR\n    T[\"THUNK(body)\"] -->|\"reserve UPDATE; claim\"| B[\"BLACKHOLE\"]\n    B -->|\"body returns value\"| V[\"INT(value)\"]\n    B -->|\"body returns error\"| E[\"ERROR(code)\"]\n    V -->|\"later demand\"| RV[\"Return stored value\"]\n    E -->|\"later demand\"| RE[\"Return stored error\"]"}, {"name": "report-2", "source": "flowchart TD\n    R[\"6: ADD\"] --> L[\"4: ADD\"]\n    R --> Q[\"5: ADD\"]\n    L -->|\"left and right\"| X[\"3: THUNK\"]\n    Q -->|\"left and right\"| X\n    X --> M[\"2: MUL\"]\n    M --> A[\"0: INT 21\"]\n    M --> B[\"1: INT 2\"]"}, {"name": "report-3", "source": "flowchart TD\n    F[\"FETCH: issue heap address\"] --> W[\"Heap wait\"]\n    W --> D[\"EVAL: dispatch node\"]\n    D -->|\"value or error\"| R[\"RET: inspect stack depth\"]\n    D -->|\"enter child\"| F\n    R -->|\"nonempty\"| SW[\"Stack wait\"]\n    SW --> RD[\"Return dispatch\"]\n    RD -->|\"evaluate right\"| F\n    RD -->|\"UPDATE\"| UW[\"Ownership read and update\"]\n    UW --> R\n    RD -->|\"MUL\"| M[\"32 iterative multiply steps\"]\n    M --> R\n    RD -->|\"ADD or error propagation\"| R\n    R -->|\"empty\"| O[\"OUTPUT: hold result\"]"}, {"name": "report-4", "source": "flowchart LR\n    UI[\"React + Redux / RTK Query\"] --> HTTP[\"Go HTTP session\"]\n    HTTP --> E[\"Engine interface\"]\n    E --> GM[\"Stepped Go model\"]\n    E --> SE[\"Serial engine\"]\n    SE --> UART[\"UART command controller\"]\n    UART --> CORE[\"FPGA reducer\"]\n    CORE --> MEM[\"Heap / stack / trace RAM\"]"}];
 const dir = "/home/manuel/code/wesen/2026-09-04--gatemate-symbolic/ttmp/2026/09/05/GATEMATE-SYMBOLIC-010--lazy-functional-language-compiler-and-source-aware-fpga-ide/reference/figures";
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

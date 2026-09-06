import {useEffect,useRef,useState} from 'react';
import {useDispatch,useSelector} from 'react-redux';
import {actions,api,RootState} from './store';
import {address,ArtifactView,CodeView,Frame,frameFields,ObjectView,sourceSlice,states} from './types';

function SourcePreview({view,span}:{view?:ArtifactView;span:number}){
 if(!view)return <p className="muted">Compile source to inspect its artifact.</p>;
 const source=view.artifact.source,range=view.artifact.spans[span];
 if(!range)return <pre className="source-preview">{source}</pre>;
 return <pre className="source-preview">{sourceSlice(source,{start:0,end:range.start})}<mark>{sourceSlice(source,range)}</mark>{sourceSlice(source,{start:range.end,end:new TextEncoder().encode(source).length})}</pre>;
}
function Environment({object,heap,choose}:{object?:ObjectView;heap:ObjectView[];choose:(ref:number)=>void}){
 if(!object)return null;
 const start=object.tag==='ENV'?object.address:object.tag==='FUN'||object.tag==='THUNK'?object.b:65535;
 const chain:ObjectView[]=[],seen=new Set<number>();let ref=start;
 while(ref!==65535&&chain.length<10&&!seen.has(ref)){seen.add(ref);const node=heap[ref];if(!node||node.tag!=='ENV')break;chain.push(node);ref=node.b;}
 if(!chain.length)return <p className="muted">This object has no captured environment.</p>;
 return <div className="environment"><div className="eyebrow">Captured environment · depth 0 first</div><svg viewBox={`0 0 290 ${chain.length*58+12}`} role="img" aria-label="Captured environment chain">{chain.map((env,i)=><g key={env.address} transform={`translate(0 ${i*58})`}>
 {i>0&&<path d="M45 -8 V10" stroke="#64748b"/>}<rect x="5" y="10" width="88" height="37" rx="5" fill="#26344f" onClick={()=>choose(env.address)}/><text x="15" y="33" fill="#cbd5e1" onClick={()=>choose(env.address)}>ENV {address(env.address)}</text>
 <path d="M93 28 H135" stroke="#8a9bbb"/><rect x="137" y="10" width="147" height="37" rx="5" fill="#203f3b" onClick={()=>choose(env.a)}/><text x="146" y="33" fill="#a7f3d0" onClick={()=>choose(env.a)}>{i}: {address(env.a)} {heap[env.a]?.tag}</text>
 </g>)}</svg>{ref!==65535&&<small className="muted">Chain continues at {address(ref)}.</small>}</div>;
}
function ObjectDetails({object,frame,view,choose}:{object?:ObjectView;frame:Frame;view?:ArtifactView;choose:(ref:number)=>void}){
 if(!object)return <div className="empty">Select a heap object to inspect its value, source and environment.</div>;
 const hasA=['CONS','FUN','THUNK','IND','ENV'].includes(object.tag),hasB=['CONS','FUN','THUNK','ENV'].includes(object.tag);
 return <><div className="object-heading"><span className={`tag tag-${object.tag}`}>{object.tag}</span><strong>{address(object.address)}</strong></div>
 <dl className="object-fields"><dt>Origin span</dt><dd>{object.span}</dd><dt>Region</dt><dd>{object.committed?'committed':'private allocation'}</dd>
 {hasA&&<><dt>{['FUN','THUNK'].includes(object.tag)?'Body code':'Reference A'}</dt><dd>{['FUN','THUNK'].includes(object.tag)?`#${object.a}`:<button className="ref-button" onClick={()=>choose(object.a)}>{address(object.a)}</button>}</dd></>}
 {hasB&&<><dt>{object.tag==='CONS'?'Tail':'Environment / parent'}</dt><dd><button className="ref-button" onClick={()=>choose(object.b)}>{address(object.b)}</button></dd></>}
 <dt>Value</dt><dd>{object.tag==='INT'?object.integer:object.tag==='BOOL'?String(!!object.payload):object.tag==='ERROR'?`fault ${object.payload}`:'—'}</dd></dl>
 <code className="packed">{frame.snapshot?.heap[object.address]}</code>
 <Environment object={object} heap={frame.heap} choose={choose}/>
 <div className="eyebrow mt-3">Allocation source</div><pre className="origin-source">{view?sourceSlice(view.artifact.source,view.artifact.spans[object.span]):''}</pre></>;
}
function CodeTable({codes,view,selectSpan}:{codes:CodeView[];view:ArtifactView;selectSpan:(span:number)=>void}){
 return <table className="table table-sm lab-table"><thead><tr><th>Code</th><th>Instruction</th><th>A / B / C</th><th>Immediate</th><th>Type</th><th>Source</th></tr></thead><tbody>{codes.map(c=><tr key={c.id} onClick={()=>selectSpan(c.span)}><td>#{c.id}</td><td className="op">{c.op}</td><td>{c.a} / {c.b} / {c.c}</td><td>{c.immediate}</td><td>{c.type}</td><td className="source-cell">{sourceSlice(view.artifact.source,view.artifact.spans[c.span])}</td></tr>)}</tbody></table>;
}
export function App(){
 const dispatch=useDispatch(),editor=useSelector((s:RootState)=>s.editor);
 const {data:state,refetch}=api.useStateQuery(),{data:examples}=api.useExamplesQuery();
 const [compile,compileRequest]=api.useCompileMutation(),[control,controlRequest]=api.useControlMutation();
 const {data:historical}=api.useHistoryQuery(editor.historyID,{skip:!editor.historyID});
 const frame=editor.historyID?historical:state?.frame;
 useEffect(()=>{dispatch(actions.selectRef(-1));dispatch(actions.selectSpan(0));},[frame?.runId,dispatch]);
 const loadedID=frame?.artifactId||'',inspectID=loadedID||editor.compiledID;
 const {data:view}=api.useArtifactQuery(inspectID,{skip:!inspectID});
 const [tab,setTab]=useState('Code'),[filter,setFilter]=useState('all'),[message,setMessage]=useState('');
 const textarea=useRef<HTMLTextAreaElement>(null);
 useEffect(()=>{if(examples&&!editor.source)dispatch(actions.edit(examples.shared));},[examples,editor.source,dispatch]);
 const busy=controlRequest.isLoading,history=!!editor.historyID,snap=frame?.snapshot;
 const selected=frame?.heap.find(o=>o.address===editor.selectedRef);
 const choose=(ref:number)=>{dispatch(actions.selectRef(ref));const o=frame?.heap[ref];if(o)dispatch(actions.selectSpan(o.span));};
 async function act(kind:string,extra:Record<string,string|number>={}){
  if(!state||history)return;setMessage('');
  try{await control({kind,expectedId:state.frame.id,runId:state.frame.runId,...extra}).unwrap();}catch(e){const error=e as {data?:{error?:string}};setMessage(error.data?.error||'Request failed. Refresh state before taking another action.');}
 }
 async function compileSource(){setMessage('');try{const result=await compile({source:editor.source,clientRevision:editor.revision}).unwrap();dispatch(actions.compiled(result));}catch{setMessage('Compilation request failed.');}}
 const heap=frame?.heap.filter(o=>(filter==='all'||filter==='mutable'&&o.address>=(snap?.constantEnd??0)||filter===o.tag))??[];
 const frozen=busy||history||!loadedID||!!frame?.needsReset;
 const manual=frozen||!!frame?.stream;
 const sourceDiffers=!!view&&view.artifact.source!==editor.source;
 return <div className="language-app">
 <header className="lab-header"><div><div className="eyebrow">GATEMATE / SYMBOLIC LABORATORY 010</div><h1>Lazy Language <span>IDE</span></h1><p>Typed source · shared evaluation · finite hardware</p></div><div className="run-meta"><span className="backend">{frame?.backend==='serial'?'● FPGA / LFL1':'● GO MODEL / LFL1'}</span><div>RUN {frame?.runId.slice(0,8)||'—'} · FRAME {frame?.id??'—'}</div><small>{loadedID?`ARTIFACT ${loadedID.slice(0,16)}`:'No program loaded'}</small></div></header>
 <div className="toolbar"><button className="btn btn-sm btn-outline-light" disabled={busy||history||!editor.compiledID||editor.compiledRevision!==editor.revision} onClick={()=>act('load',{artifactId:editor.compiledID})}>Load compiled artifact</button><button className="btn btn-sm btn-outline-secondary" disabled={busy||history} onClick={()=>act('reset')}>Reset</button><span className="divider"/>
 <button className="btn btn-sm btn-primary" disabled={manual||snap?.state!==0} onClick={()=>act('force',{ref:snap?.root??0})}>Force main</button><button className="btn btn-sm btn-outline-light" disabled={manual} onClick={()=>act('tick',{ticks:1})}>Step 1</button><button className="btn btn-sm btn-outline-light" disabled={manual} onClick={()=>act('tick',{ticks:1000})}>Run 1000 ticks</button><button className="btn btn-sm btn-outline-light" disabled={manual||!snap?.valid} onClick={()=>act('poll')}>Poll result</button>
 <span className="state-badge">{history?'HISTORY · ':''}{snap?states[snap.state]:'READY'}</span><button className="btn btn-sm btn-link" onClick={()=>refetch()}>Refresh</button></div>
 {(message||frame?.error)&&<div className="alert alert-danger mx-3 my-2">{message||frame?.error}{frame?.needsReset&&' Reset is required.'}</div>}
 <div className="counters">{[['CLAIMS','claims'],['UPDATES','updates'],['MULTIPLICATIONS','multiplies'],['ALLOCATIONS','allocations'],['HEAP HIGH WATER','maxHeap'],['STACK HIGH WATER','maxStack']].map(([label,key])=><div key={key}><span>{label}</span><strong>{snap?.counters[key]??0}</strong></div>)}</div>
 <main className="workbench">
 <section className="panel source-panel"><div className="panel-title"><h2>01 / Source</h2><span>revision {editor.revision}</span></div><div className="source-actions"><select aria-label="Example program" className="form-select form-select-sm" defaultValue="shared" onChange={e=>{if(examples)dispatch(actions.edit(examples[e.target.value]));}}>{Object.keys(examples??{}).map(name=><option key={name}>{name}</option>)}</select><button className="btn btn-sm btn-primary" disabled={compileRequest.isLoading} onClick={compileSource}>{compileRequest.isLoading?'Compiling…':'Compile'}</button></div>
 <textarea ref={textarea} aria-label="Language source" spellCheck={false} value={editor.source} onChange={e=>dispatch(actions.edit(e.target.value))}/>
 <div className="source-status">{editor.diagnostics.length?`${editor.diagnostics.length} diagnostics`:editor.compiledRevision===editor.revision&&editor.compiledID?'✓ Checked and compiled':'Source requires compilation'}{sourceDiffers&&<span>Editor differs from loaded source</span>}</div>
 {editor.diagnostics.map((d,i)=><button className="diagnostic" key={i} onClick={()=>{const start=sourceSlice(editor.source,{start:0,end:d.span.start}).length,end=sourceSlice(editor.source,{start:0,end:d.span.end}).length;textarea.current?.focus();textarea.current?.setSelectionRange(start,end);}}><strong>{d.code}</strong> {d.message}<small>bytes {d.span.start}–{d.span.end}</small></button>)}
 <details open={editor.selectedSpan>0}><summary>{loadedID?'Loaded source':'Compiled source'} · span {editor.selectedSpan}</summary><SourcePreview view={view} span={editor.selectedSpan}/></details>
 </section>
 <section className="panel heap-panel"><div className="panel-title"><h2>02 / Heap</h2><span>{snap?.committedTop??0} / 2048 objects</span></div><div className="heap-controls"><select className="form-select form-select-sm" aria-label="Heap filter" value={filter} onChange={e=>setFilter(e.target.value)}>{['all','mutable','THUNK','IND','FUN','ENV','CONS','INT','ERROR'].map(value=><option key={value}>{value}</option>)}</select><span className="muted">root {address(snap?.root??65535)} · constants &lt; {snap?.constantEnd??0}</span></div>
 {snap?.allocation.active&&<div className="allocation-status">Allocation [{snap.allocation.base}, {snap.allocation.reservedEnd}) · stage {snap.allocation.stage} · write {snap.allocation.nextWrite}/{snap.allocation.objectCount}</div>}
 <div className="heap-scroll"><table className="table table-sm lab-table"><thead><tr><th>Ref</th><th>Object</th><th>Contents</th><th>Span</th></tr></thead><tbody>{heap.map(o=><tr key={o.address} className={`${editor.selectedRef===o.address?'selected ':''}${editor.selectedSpan&&o.span===editor.selectedSpan?'same-origin':''}`} onClick={()=>choose(o.address)}><td>{address(o.address)}</td><td><span className={`tag tag-${o.tag}`}>{o.tag}</span></td><td>{o.tag==='INT'?o.integer:o.tag==='BOOL'?String(!!o.payload):o.tag==='ERROR'?`fault ${o.payload}`:['FUN','THUNK'].includes(o.tag)?`code #${o.a} · env ${address(o.b)}`:o.tag==='IND'?`→ ${address(o.a)}`:['CONS','ENV'].includes(o.tag)?`${address(o.a)} · ${address(o.b)}`:'—'}{!o.committed&&' (private)'}</td><td>{o.span}</td></tr>)}</tbody></table>{!heap.length&&<div className="empty">Compile and load a program to inspect its initial heap.</div>}</div>
 <div className="result-strip">{snap?.valid?<>OUTPUT READY <button onClick={()=>choose(snap.resultRef)}>{address(snap.resultRef)} · {frame?.heap[snap.resultRef]?.tag} {frame?.heap[snap.resultRef]?.tag==='INT'?frame.heap[snap.resultRef].integer:''}</button></>:frame?.polled!=null?<>POLLED <button onClick={()=>choose(frame.polled!)}>{address(frame.polled)}</button></>:<>Output is held until explicitly polled.</>}</div>
 </section>
 <section className="panel inspector-panel"><div className="panel-title"><h2>03 / Inspector</h2><span>source + environment</span></div>{frame&&<ObjectDetails object={selected} frame={frame} view={view} choose={choose}/>}</section>
 </main>
 <section className="stream-panel"><div><div className="eyebrow">LAZY STREAM</div><h2>Demand one element at a time</h2><p>{frame?.stream?`${frame.stream.phase} · cursor ${address(frame.stream.cursor)}${frame.stream.pending?' · pending demand':''}${frame.stream.complete?' · complete':''}`:'ListInt programs expose a constructor and suspended fields.'}</p></div><div className="stream-values">{frame?.stream?.values.map((v,i)=><span key={i}>{v}</span>)}{frame?.stream?.fault?<span className="fault">fault {frame.stream.fault}</span>:null}</div><div className="stream-actions"><button className="btn btn-sm btn-outline-light" disabled={frozen||!!frame?.stream||view?.artifact.entryType!=='ListInt'||snap?.state!==0} onClick={()=>act('stream-start')}>Start stream</button><button className="btn btn-sm btn-primary" disabled={frozen||!frame?.stream||frame.stream.pending||frame.stream.complete} onClick={()=>act('stream-next',{ticks:1000})}>Next element</button><button className="btn btn-sm btn-outline-light" disabled={frozen||!frame?.stream?.pending} onClick={()=>act('stream-resume',{ticks:1000})}>Resume demand</button></div></section>
 <section className="panel lower-panel"><div className="lower-toolbar"><nav>{['Code','Stack','Mutations','Bindings','Raw state'].map(name=><button className={tab===name?'active':''} key={name} onClick={()=>setTab(name)}>{name}</button>)}</nav><label>History <select aria-label="History frame" value={editor.historyID} onChange={e=>dispatch(actions.history(Number(e.target.value)))}><option value={0}>Live frame</option>{state?.history.slice().reverse().map(h=><option key={h.id} value={h.id}>#{h.id} · {h.operation} · {h.runId.slice(0,6)}</option>)}</select></label></div>
 {history&&<div className="history-notice">Historical frame is read-only. Select “Live frame” to control the current run.</div>}
 <div className="lower-content">
 {tab==='Code'&&view&&<CodeTable codes={view.code} view={view} selectSpan={span=>dispatch(actions.selectSpan(span))}/>}
 {tab==='Stack'&&<table className="table table-sm lab-table"><thead><tr><th>Depth</th><th>Continuation</th><th>A / B / C</th><th>Saved ref</th><th>Span</th><th>Packed 128-bit frame</th></tr></thead><tbody>{snap?.stack.map((packed,i)=>{const f=frameFields(packed);return <tr key={i} onClick={()=>dispatch(actions.selectSpan(f.span))}><td>{i}</td><td>{['','ARG','PRIM RIGHT','PRIM APPLY','IF','CASE','UPDATE'][f.kind]}</td><td>{f.a} / {f.b} / {f.c}</td><td>{address(f.saved)}</td><td>{f.span}</td><td><code>{packed}</code></td></tr>;})}</tbody></table>}
 {tab==='Mutations'&&<><p className="muted">First 64 mutation events. Dropped: {snap?.counters.traceDropped??0}. Source span is preserved across updates.</p><table className="table table-sm lab-table"><thead><tr><th>Cycle</th><th>Mutation</th><th>Object</th><th>Span</th><th>Old → New</th></tr></thead><tbody>{snap?.trace.map((e,i)=><tr key={i} onClick={()=>choose(e.address)}><td>{e.cycle}</td><td>{['','ALLOCATE','CLAIM','UPDATE'][e.kind]}</td><td>{address(e.address)}</td><td>{e.span}</td><td><code>{e.old} → {e.new}</code></td></tr>)}</tbody></table></>}
 {tab==='Bindings'&&<table className="table table-sm lab-table"><thead><tr><th>ID</th><th>Name</th><th>Type</th><th>Source bytes</th><th>Uses / lexical depths</th></tr></thead><tbody>{view?.artifact.bindings.map(b=><tr key={b.id}><td>{b.id}</td><td>{b.name}</td><td>{b.type}</td><td>{b.span.start}–{b.span.end}</td><td>{view.artifact.uses.filter(u=>u.binding===b.id).map(u=>`${u.span.start}: depth ${u.depth}`).join(' · ')}</td></tr>)}</tbody></table>}
 {tab==='Raw state'&&<pre>{JSON.stringify(frame,null,2)}</pre>}
 </div></section><footer>LFL1 v1 · 2048 heap slots · 512 continuation frames · 64 mutation events · no garbage collection</footer>
 </div>;
}

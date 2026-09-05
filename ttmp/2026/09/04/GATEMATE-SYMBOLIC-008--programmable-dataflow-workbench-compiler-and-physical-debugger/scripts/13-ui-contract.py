from pathlib import Path
p=Path('web/src/dataflow/types.ts');s=p.read_text();s='''export interface Destination {Node:number;Port:number}
export interface Descriptor {Op:number;Required:number;Destinations:[Destination,Destination];Count:number;Final:boolean}
export interface GraphImage {count:number;descriptors:Descriptor[]}
export interface Program {source:string;graph:GraphImage;inputs:{name:string;type:string;destinations:Destination[]}[];constants:{destination:Destination;value:number}[];nodes:{node:number;name:string;line:number;type:string}[]}
export interface DebugControl {mask:number;node:number;context:number;resume?:boolean;clear?:boolean}
export interface DebugState {halted:boolean;reason:number;stopCycle:number;mask:number;node:number;context:number;dropped:number;events:{cycle:number;kind:number;token:Token}[]}
export const opNames=['MUL','ADD','LT','BOOL → INT','COPY','SUB'];
'''+s;s=s.replace('export interface Snapshot {','export interface Snapshot {\n graph:GraphImage;debug:DebugState;').replace('export interface Frame extends FrameInfo {snapshot:Snapshot}','export interface Frame extends FrameInfo {snapshot:Snapshot;program:Program|null}').replace('export interface Operation {kind:string;','export interface Operation {kind:string;debug?:DebugControl;graph?:GraphImage;')
p.write_text(s)
p=Path('web/src/dataflow/store.ts');s=p.read_text().replace('type {Frame,Operation,State,Validation}', 'type {Frame,Operation,State,Validation,Program}')
s=s.replace('state:build.query', '''compile:build.mutation<{program:Program|null;diagnostics:string[]},string>({query:source=>({url:'/program/compile',method:'POST',body:{source}})}),
 loadProgram:build.mutation<State,{source:string;expectedId:number}>({query:body=>({url:'/program/load',method:'POST',body}),invalidatesTags:['State']}),
 programInputs:build.mutation<State,{expectedId:number;context:number;values:Record<string,number>}>({query:body=>({url:'/program/inputs',method:'POST',body}),invalidatesTags:['State']}),
 programs:build.query<string[],void>({query:()=>'/programs',providesTags:['Projects']}),
 program:build.query<{source:string},string>({query:id=>`/programs/${encodeURIComponent(id)}`}),
 saveProgram:build.mutation<{saved:boolean},{id:string;source:string}>({query:({id,source})=>({url:`/programs/${encodeURIComponent(id)}`,method:'PUT',body:{source}}),invalidatesTags:['Projects']}),
 state:build.query''',1)
s=s.replace('interface Workspace {source:string;', 'interface Workspace {programSource:string;source:string;').replace("initialState:{source:'',", "initialState:{programSource:'input a, b, c: int16\\n\\nlet square = a * a\\nlet offset = b * c\\nlet selected = int(a < b)\\n\\noutput square + offset + selected\\n',source:'',")
s=s.replace('setSource(s,a:PayloadAction<string>)','setProgramSource(s,a:PayloadAction<string>){s.programSource=a.payload},setSource(s,a:PayloadAction<string>)')
p.write_text(s)
p=Path('web/src/dataflow/Inspectors.tsx');s=p.read_text().replace('nodeNames,','opNames,').replace('{nodeNames[node]}','{node<s.graph.count?opNames[s.graph.descriptors[node].Op]:\'Inactive node\'}').replace("node===4||node===6?'A':'A + B'","s.graph.descriptors[node]?.Required===1?'A':'A + B'");p.write_text(s)
p=Path('web/src/dataflow/App.tsx');s=p.read_text().replace("import {Graph} from './Graph';", "import {Graph} from './Graph';\nimport {ProgramEditor,Debugger} from './Workbench';")
s=s.replace('<h1>Elastic Dataflow</h1><p>Scenario editor · execution debugger · FPGA instrument</p>','<h1>Programmable Dataflow</h1><p>Typed expressions · compiled graphs · physical execution</p>')
s=s.replace('<aside className="df-panel df-editor"><div', '<aside className="df-panel df-editor"><ProgramEditor state={state} frame={frame} locked={locked}/><details className="mt-4"><summary>Low-level scenario experiments</summary><div',1)
s=s.replace('   </aside>','   </details></aside>',1)
s=s.replace('<span>a × b + c × d + int(e &lt; f)</span>','<span>{snapshot.graph.count} active nodes · {frame?.program?\'compiled program\':\'reset graph\'}</span>')
s=s.replace('<Graph snapshot={snapshot}', '<Graph program={frame?.program??null} snapshot={snapshot}')
s=s.replace('<Pipelines snapshot={snapshot}/>','<Debugger snapshot={snapshot} program={frame?.program??null} selected={workspace.node} context={workspace.context} locked={locked||Boolean(state?.needsReset)} invoke={invoke}/><Pipelines snapshot={snapshot}/>')
s=s.replace('Fixed graph / 4 contexts','Programmable graph / 4 contexts').replace('GATEMATE-SYMBOLIC-007','GATEMATE-SYMBOLIC-008')
p.write_text(s)

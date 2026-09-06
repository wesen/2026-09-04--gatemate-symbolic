export interface Span { start:number; end:number }
export interface Diagnostic { code:string; message:string; span:Span }
export interface ObjectView {address:number;tag:string;a:number;b:number;payload:number;integer:number;span:number;committed:boolean}
export interface CodeView {id:number;op:string;a:number;b:number;c:number;immediate:number;span:number;type:string}
export interface Artifact {id:string;version:string;profile:string;source:string;entryType:string;root:number;constantEnd:number;code:string[];heap:string[];provenance:number[];spans:Span[];bindings:{id:number;name:string;span:Span;type:string}[];uses:{binding:number;depth:number;span:Span}[];codeTypes:string[]}
export interface ArtifactView {artifact:Artifact;code:CodeView[]}
export interface Mutation {cycle:number;address:number;span:number;kind:number;old:string;new:string}
export interface Snapshot {artifactId:string;state:number;root:number;constantEnd:number;currentRef:number;codeId:number;envRef:number;resultRef:number;valid:boolean;committedTop:number;indSteps:number;envSteps:number;heap:string[];provenance:number[];stack:string[];trace:Mutation[];counters:Record<string,number>;allocation:{active:boolean;base:number;reservedEnd:number;objectCount:number;nextWrite:number;kind:number;stage:number}}
export interface Stream {cursor:number;phase:string;pending:boolean;sentDemand:boolean;values:number[];complete:boolean;fault:number}
export interface Frame {id:number;runId:string;artifactId:string;backend:string;operation:string;needsReset:boolean;error:string;snapshot:Snapshot|null;heap:ObjectView[];stream:Stream|null;polled:number|null}
export interface State {frame:Frame;history:{id:number;runId:string;artifactId:string;operation:string}[]}
export interface Operation {kind:string;expectedId:number;runId:string;artifactId?:string;ref?:number;ticks?:number}
export interface CompileResult {clientRevision:number;artifactId:string;diagnostics:Diagnostic[]}
export const states=['IDLE','CODE ISSUE','CODE WAIT','EVAL','HEAP ISSUE','HEAP WAIT','ENTER','ENV ISSUE','ENV WAIT','ENV DISPATCH','RETURN','STACK WAIT','RETURN DISPATCH','ALLOCATE','UPDATE ISSUE','UPDATE WAIT','UPDATE WRITE','MULTIPLY','OUTPUT','LOADING'];
export const address=(n:number)=>n===65535?'∅':`@${n}`;
export function sourceSlice(source:string,span?:Span){if(!span)return '';return new TextDecoder().decode(new TextEncoder().encode(source).slice(span.start,span.end));}
export function frameFields(packed:string){return {kind:parseInt(packed.slice(0,2),16),op:parseInt(packed.slice(2,4),16),a:parseInt(packed.slice(4,8),16),b:parseInt(packed.slice(8,12),16),c:parseInt(packed.slice(12,16),16),saved:parseInt(packed.slice(20,24),16),span:parseInt(packed.slice(24,28),16)};}

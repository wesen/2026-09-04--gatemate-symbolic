export interface Destination {Node:number;Port:number}
export interface Descriptor {Op:number;Required:number;Destinations:[Destination,Destination];Count:number;Final:boolean}
export interface GraphImage {count:number;descriptors:Descriptor[]}
export interface Program {source:string;graph:GraphImage;inputs:{name:string;type:string;destinations:Destination[]}[];constants:{destination:Destination;value:number}[];nodes:{node:number;name:string;line:number;type:string}[]}
export interface DebugControl {mask:number;node:number;context:number;resume?:boolean;clear?:boolean}
export interface DebugState {halted:boolean;reason:number;stopCycle:number;mask:number;node:number;context:number;dropped:number;events:{cycle:number;kind:number;token:Token}[]}
export const opNames=['MUL','ADD','LT','BOOL → INT','COPY','SUB'];
export interface Token { context:number; epoch:number; node:number; port:number; final:boolean; producer:number; subtype:number; value:number }
export interface Config { InputDepth:number; CompletionDepth:number; OutputDepth:number; MulLatency:number; EpochBits:number }
export interface Slot { context:number; node:number; values:[number,number]; valid:number; issued:boolean; pending:boolean }
export interface Snapshot {
 graph:GraphImage;debug:DebugState;
 source:'model'|'serial'; config:Config; epochs:number[]; closed:boolean[]; slots:Slot[];
 issue:{token:Token;phase:number;values:[number,number]}|null;
 router:{token:Token;delivered:number}|null;
 mul:(Token|null)[]; alu:(Token|null)[]; input:Token[]; completion:Token[]; output:Token[];
 errors:(Token|null)[]; counters:Record<string,number>; quiescent:boolean;
}
export interface FrameInfo {id:number;generation:number;label:string;cycle:number;at:string}
export interface Frame extends FrameInfo {snapshot:Snapshot;program:Program|null}
export interface State {current:Frame;history:FrameInfo[];results:Token[];events:{label:string;detail:string;at:string;error:boolean}[];running:boolean;scenario:string;step:number;actions:number;error:string;needsReset:boolean}
export interface Operation {kind:string;debug?:DebugControl;graph?:GraphImage;token?:Token;context?:number;ticks?:number;config?:Config}
export interface Diagnostic {action:number;message:string}
export interface Validation {diagnostics:Diagnostic[];formatted:string;actions:number}
export const nodeNames=['MUL','MUL','ADD','LT','BOOL → INT','ADD / FINAL','COPY'];
export function valueText(value:number):string {
 const tag=Math.floor(value/2**36);const payload=value>>>0;
 if(tag===0)return `${payload|0}`;
 if(tag===1)return payload===0?'false':'true';
 if(tag===13)return `ERROR ${payload}`;
 return `tag ${tag}: ${payload}`;
}
export function valueType(value:number):string{return ['INT','BOOL'][Math.floor(value/2**36)]??(Math.floor(value/2**36)===13?'ERROR':'INVALID')}
export function intValue(value:number):number {if(!Number.isInteger(value)||value<-(2**31)||value>2**31-1)throw new Error('Value must be a signed 32-bit integer');return value>>>0}

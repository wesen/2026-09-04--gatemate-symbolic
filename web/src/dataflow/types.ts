export interface Token { context:number; epoch:number; node:number; port:number; final:boolean; producer:number; subtype:number; value:number }
export interface Config { InputDepth:number; CompletionDepth:number; OutputDepth:number; MulLatency:number; EpochBits:number }
export interface Slot { context:number; node:number; values:[number,number]; valid:number; issued:boolean; pending:boolean }
export interface Snapshot {
 source:'model'|'serial'; config:Config; epochs:number[]; closed:boolean[]; slots:Slot[];
 issue:{token:Token;phase:number;values:[number,number]}|null;
 router:{token:Token;delivered:number}|null;
 mul:(Token|null)[]; alu:(Token|null)[]; input:Token[]; completion:Token[]; output:Token[];
 errors:(Token|null)[]; counters:Record<string,number>; quiescent:boolean;
}
export interface FrameInfo {id:number;generation:number;label:string;cycle:number;at:string}
export interface Frame extends FrameInfo {snapshot:Snapshot}
export interface State {current:Frame;history:FrameInfo[];results:Token[];events:{label:string;detail:string;at:string;error:boolean}[];running:boolean;scenario:string;step:number;actions:number;error:string;needsReset:boolean}
export interface Operation {kind:string;token?:Token;context?:number;ticks?:number;config?:Config}
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

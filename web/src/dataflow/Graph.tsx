import type {Snapshot,Slot} from './types';
import {nodeNames,valueText} from './types';
const positions:Record<number,[number,number]>={6:[30,160],0:[225,40],1:[225,220],2:[430,110],3:[225,390],4:[430,360],5:[650,225]};
const edges:[number,number,string][]=[[6,0,'A'],[6,1,'A'],[0,2,'A'],[1,2,'B'],[2,5,'A'],[3,4,'A'],[4,5,'B']];
export function slotStatus(s:Snapshot,slot:Slot|undefined):string{
 if(!slot)return 'empty';if(s.errors[slot.context])return 'fault';
 if([...s.mul,...s.alu].some(t=>t&&t.context===slot.context&&t.node===slot.node&&t.epoch===s.epochs[slot.context]))return 'executing';
 if(s.issue?.token.context===slot.context&&s.issue.token.node===slot.node)return 'issue';
 if(slot.pending)return 'ready';if(slot.issued)return 'issued';if(slot.valid)return 'waiting';return 'empty';
}
export function Graph({snapshot,context,selected,onSelect}:{snapshot:Snapshot;context:number;selected:number;onSelect:(n:number)=>void}){
 return <svg className="df-graph" viewBox="0 0 850 540" role="img" aria-label={`Fixed expression graph for context ${context}`}>
  <defs><marker id="df-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse"><path d="M 0 0 L 10 5 L 0 10 z" fill="currentColor"/></marker></defs>
  {edges.map(([a,b,port])=>{const [x,y]=positions[a];const [u,v]=positions[b];const destY=v+(port==='A'?34:70);return <g key={`${a}-${b}`} className="df-edge"><path d={`M${x+155},${y+48} C${x+185},${y+48} ${u-40},${destY} ${u},${destY}`} markerEnd="url(#df-arrow)"/><text x={u-18} y={destY-6}>{port}</text></g>})}
  {Object.entries(positions).map(([id,[x,y]])=>{const n=Number(id);const slot=snapshot.slots.find(s=>s.context===context&&s.node===n);const status=slotStatus(snapshot,slot);return <g key={n} transform={`translate(${x},${y})`} className={`df-node ${status} ${selected===n?'selected':''}`} role="button" tabIndex={0} aria-label={`Node ${n} ${nodeNames[n]} ${status}`} onClick={()=>onSelect(n)} onKeyDown={e=>{if(e.key==='Enter'||e.key===' '){e.preventDefault();onSelect(n)}}}>
   <rect width="155" height="94" rx="8"/><text className="df-node-id" x="12" y="20">N{n} · {status.toUpperCase()}</text><text className="df-node-op" x="12" y="45">{nodeNames[n]}</text>
   <text className="df-node-value" x="12" y="69">A {slot&&slot.valid&1?valueText(slot.values[0]):'—'}{n!==4&&n!==6?`   B ${slot&&slot.valid&2?valueText(slot.values[1]):'—'}`:''}</text>
  </g>})}
  <text className="df-graph-note" x="28" y="516">Fixed descriptors · shared arithmetic units · explicit destinations</text>
 </svg>
}

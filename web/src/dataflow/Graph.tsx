import type {Snapshot,Slot,Program,GraphImage} from './types';
import {opNames,valueText} from './types';
export function slotStatus(s:Snapshot,slot:Slot|undefined):string{
 if(!slot)return 'empty';if(s.errors[slot.context])return 'fault';
 if([...s.mul,...s.alu].some(t=>t&&t.context===slot.context&&t.node===slot.node&&t.epoch===s.epochs[slot.context]))return 'executing';
 if(s.issue?.token.context===slot.context&&s.issue.token.node===slot.node)return 'issue';
 if(slot.pending)return 'ready';if(slot.issued)return 'issued';if(slot.valid)return 'waiting';return 'empty';
}
export function layout(graph:GraphImage){
 const depth=Array(graph.count).fill(0) as number[];
 // Bounded relaxation also supports the reset graph's non-topological IDs.
 for(let pass=0;pass<graph.count;pass++)graph.descriptors.slice(0,graph.count).forEach((d,n)=>d.Destinations.slice(0,d.Count).forEach(t=>{depth[t.Node]=Math.max(depth[t.Node],depth[n]+1)}));
 const rows:Record<number,number>={};const positions=depth.map(d=>{const row=rows[d]??0;rows[d]=row+1;return [30+d*205,35+row*135] as [number,number]});
 return {positions,width:Math.max(...depth)*205+225,height:Math.max(1,...Object.values(rows))*135+70};
}
export function Graph({snapshot,program,context,selected,onSelect}:{snapshot:Snapshot;program:Program|null;context:number;selected:number;onSelect:(n:number)=>void}){
 const {positions,width,height}=layout(snapshot.graph);
 return <svg className="df-graph" viewBox={`0 0 ${width} ${height}`} role="img" aria-label={`Active expression graph for context ${context}`}>
  <defs><marker id="df-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse"><path d="M 0 0 L 10 5 L 0 10 z" fill="currentColor"/></marker></defs>
  {snapshot.graph.descriptors.slice(0,snapshot.graph.count).flatMap((d,a)=>d.Destinations.slice(0,d.Count).map(({Node:b,Port:port},i)=>{const [x,y]=positions[a];const [u,v]=positions[b];const destY=v+(port===0?38:75);return <g key={`${a}-${b}-${i}`} className="df-edge"><path d={`M${x+165},${y+48} C${x+190},${y+48} ${u-35},${destY} ${u},${destY}`} markerEnd="url(#df-arrow)"/><text x={u-18} y={destY-6}>{port?'B':'A'}</text></g>}))}
  {positions.map(([x,y],n)=>{const d=snapshot.graph.descriptors[n];const slot=snapshot.slots.find(s=>s.context===context&&s.node===n);const status=slotStatus(snapshot,slot);const info=program?.nodes.find(v=>v.node===n);return <g key={n} transform={`translate(${x},${y})`} className={`df-node ${status} ${selected===n?'selected':''}`} role="button" tabIndex={0} aria-label={`Node ${n} ${opNames[d.Op]} ${status}`} onClick={()=>onSelect(n)} onKeyDown={e=>{if(e.key==='Enter'||e.key===' '){e.preventDefault();onSelect(n)}}}>
   <title>{info?`${info.name} · source line ${info.line} · ${info.type}`:opNames[d.Op]}</title><rect width="165" height="110" rx="8"/><text className="df-node-id" x="10" y="18">N{n} · {status.toUpperCase()}</text><text className="df-node-op" x="10" y="43">{opNames[d.Op]}{d.Final?' / OUT':''}</text>
   <text className="df-node-value" x="10" y="64">{info?.name.slice(0,21)??'descriptor'}</text><text className="df-node-value" x="10" y="88">A {slot&&slot.valid&1?valueText(slot.values[0]):'—'}{d.Required===3?`  B ${slot&&slot.valid&2?valueText(slot.values[1]):'—'}`:''}</text>
  </g>})}
  <text className="df-graph-note" x="28" y={height-10}>Active device descriptors · explicit operand ports</text>
 </svg>
}

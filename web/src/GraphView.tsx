import type { Graph, Snapshot } from './types';
import { candidates, palette } from './graph';

export function GraphView({ graph, snapshot, selected, onSelect }: { graph: Graph; snapshot: Snapshot; selected: number | null; onSelect: (v: number) => void }) {
  const points = Array.from({ length: graph.vertices }, (_, i) => ({ x: 240 + 140 * Math.cos(2 * Math.PI * i / graph.vertices - Math.PI / 2), y: 180 + 125 * Math.sin(2 * Math.PI * i / graph.vertices - Math.PI / 2) }));
  return <svg className="graph-view" viewBox="0 0 480 360" aria-label="Graph and candidate colors">
    {graph.edges.map(([a, b]) => <line key={`${a}-${b}`} x1={points[a].x} y1={points[a].y} x2={points[b].x} y2={points[b].y} stroke="#b8c4c0" strokeWidth="2" />)}
    {points.map((p, v) => {
      const values = candidates(snapshot.domains[v]); const color = values.length === 1 ? palette[values[0]] : values.length === 0 ? '#991b1b' : '#fff';
      const label = `Vertex ${v}, candidates ${values.length ? values.join(', ') : 'none'}`;
      return <g key={v} role="button" tabIndex={0} aria-label={label} aria-pressed={selected === v} onClick={() => onSelect(v)} onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onSelect(v) } }} className="vertex">
        <circle cx={p.x} cy={p.y} r={selected === v ? 30 : 26} fill={color} stroke={selected === v ? '#111827' : '#72827c'} strokeWidth={selected === v ? 3 : 1.5} />
        <text x={p.x} y={p.y + 5} textAnchor="middle" fill={values.length <= 1 ? 'white' : '#17251f'} fontSize="15" fontWeight="700">v{v}</text>
        <text x={p.x} y={p.y + 46} textAnchor="middle" fill="#40564b" fontSize="12">{values.length ? values.map((c) => `C${c}`).join(' ') : 'EMPTY DOMAIN'}</text>
        <g transform={`translate(${p.x - values.length * 6},${p.y + 57})`}>{values.map((c, i) => <rect key={c} x={i * 12} width={9} height={5} rx={2} fill={palette[c]} />)}</g>
      </g>;
    })}
  </svg>;
}

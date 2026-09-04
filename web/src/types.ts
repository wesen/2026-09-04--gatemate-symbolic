export interface Graph { vertices: number; colors: number; edges: [number, number][]; firstOnly: boolean }
export interface Choice { vertex: number; remaining: number; mark: number; propagated: number }
export interface TrailEntry { vertex: number; old: number; level: number }
export interface Snapshot {
  sequence: number; kind: number; name: string; domains: number[]; propagated: number;
  choiceTop: number; trailTop: number; base: number; count: number; result: number; fault: number;
  choiceWord: number; trailWord: number; choices: Choice[]; trail: TrailEntry[];
}
export interface Summary { sequence: number; kind: number; name: string; count: number; choiceTop: number; trailTop: number }
export interface State { engine: 'model' | 'serial'; loaded: boolean; graph: Graph; generation: number; running: boolean; error: string; latest: Snapshot; history: Summary[] }
export type Action = 'step' | 'run' | 'pause' | 'reset';

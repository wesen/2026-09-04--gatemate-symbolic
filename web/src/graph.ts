import type { Graph } from './types';

export const palette = ['#047857', '#c2410c', '#4338ca', '#be185d', '#0369a1', '#854d0e', '#6d28d9', '#334155'];
export const candidates = (mask: number) => Array.from({ length: 8 }, (_, i) => i).filter((i) => (mask & (1 << i)) !== 0);
export const hex = (mask: number) => mask.toString(16).toUpperCase().padStart(2, '0');
export const presets: { name: string; graph: Graph }[] = [
  { name: 'Triangle · 3 colors', graph: { vertices: 3, colors: 3, edges: [[0, 1], [1, 2], [0, 2]], firstOnly: false } },
  { name: 'Triangle · 2 colors', graph: { vertices: 3, colors: 2, edges: [[0, 1], [1, 2], [0, 2]], firstOnly: false } },
  { name: 'Path · 2 colors', graph: { vertices: 4, colors: 2, edges: [[0, 1], [1, 2], [2, 3]], firstOnly: false } },
  { name: 'Square · 3 colors', graph: { vertices: 4, colors: 3, edges: [[0, 1], [1, 2], [2, 3], [3, 0]], firstOnly: false } },
];
export const edgeText = (g: Graph) => g.edges.map(([a, b]) => `${a}-${b}`).join(', ');
export function parseGraph(vertices: number, colors: number, text: string, firstOnly: boolean): Graph {
  if (!Number.isInteger(vertices) || vertices < 1 || vertices > 8 || !Number.isInteger(colors) || colors < 1 || colors > 8) throw new Error('Vertices and colors must be whole numbers from 1 to 8.');
  const edges: [number, number][] = []; const seen = new Set<string>();
  for (const token of text.trim().split(/[\s,]+/).filter(Boolean)) {
    const match = /^(\d+)-(\d+)$/.exec(token);
    if (!match) throw new Error(`Use vertex pairs such as 0-1. Invalid edge: ${token}`);
    const a = Number(match[1]), b = Number(match[2]);
    if (a >= vertices || b >= vertices || a === b) throw new Error(`Edge ${token} needs two distinct vertices in 0…${vertices - 1}.`);
    const key = [a, b].sort((x, y) => x - y).join('-');
    if (seen.has(key)) throw new Error(`Edge ${token} duplicates an undirected edge.`);
    seen.add(key); edges.push([a, b]);
  }
  return { vertices, colors, edges, firstOnly };
}
export function errorText(error: unknown): string {
  if (error && typeof error === 'object' && 'data' in error) {
    const data = error.data;
    if (data && typeof data === 'object' && 'error' in data && typeof data.error === 'string') return data.error;
  }
  return error instanceof Error ? error.message : 'The instrument request failed. Check the connection and reset or reload if needed.';
}

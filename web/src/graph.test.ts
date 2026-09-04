import { describe, expect, it } from 'vitest';
import { candidates, parseGraph } from './graph';
import { ui } from './store';

describe('graph editing and view state', () => {
  it('parses labeled undirected edges and preserves first-only', () => {
    expect(parseGraph(3, 3, '0-1, 1-2\n0-2', true)).toEqual({ vertices: 3, colors: 3, edges: [[0, 1], [1, 2], [0, 2]], firstOnly: true });
    expect(parseGraph(2, 1, '', false).edges).toEqual([]);
    expect(candidates(0x82)).toEqual([1, 7]);
  });
  it.each(['0-0', '0-3', '0-1,1-0', '0:1', '0-1-2'])('rejects invalid edge input %s', (text) => {
    expect(() => parseGraph(3, 2, text, false)).toThrow();
  });
  it('bounds dimensions and separates history from vertex selection', () => {
    expect(() => parseGraph(9, 2, '', false)).toThrow();
    expect(() => parseGraph(3, 1.5, '', false)).toThrow();
    let view = ui.reducer(undefined, ui.actions.inspect(8));
    view = ui.reducer(view, ui.actions.selectVertex(2));
    expect(view).toEqual({ sequence: 8, vertex: 2 });
    expect(ui.reducer(view, ui.actions.live())).toEqual({ sequence: null, vertex: 2 });
    expect(ui.reducer(view, ui.actions.reset())).toEqual({ sequence: null, vertex: null });
  });
});

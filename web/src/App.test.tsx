import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Provider } from 'react-redux';
import App from './App';
import { api, makeStore, type AppStore } from './store';
import type { Graph, Snapshot, State } from './types';

const graph: Graph = { vertices: 3, colors: 3, edges: [[0, 1], [1, 2], [0, 2]], firstOnly: false };
const initial = (g: Graph): Snapshot => ({ sequence: 0, kind: 0, name: 'READY', domains: Array.from({ length: 8 }, (_, v) => v < g.vertices ? (1 << g.colors) - 1 : 1), propagated: 255 ^ ((1 << g.vertices) - 1), choiceTop: 0, trailTop: 0, base: 0, count: 0, result: 0, fault: 0, choiceWord: 0, trailWord: 0, choices: [], trail: [] });
let state: State, records: Map<number, Snapshot>, store: AppStore, failRun: boolean;
beforeEach(() => {
  state = { engine: 'serial', loaded: false, graph, generation: 0, running: false, error: '', latest: initial(graph), history: [] };
  records = new Map(); failRun = false;
  vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const req = input instanceof Request ? input : new Request(input, init);
    const path = new URL(req.url).pathname;
    let data: unknown = state;
    if (req.method === 'POST' && path === '/api/graph') {
      const g = await req.json() as Graph;
      state = { ...state, graph: g, loaded: true, generation: state.generation + 1, latest: initial(g), history: [] }; records.clear(); data = state;
    } else if (req.method === 'POST' && path === '/api/control') {
      const { action } = await req.json() as { action: string };
      if (action === 'run' && failRun) return new Response(JSON.stringify({ error: 'USB cable disconnected' }), { status: 502, headers: { 'Content-Type': 'application/json' } });
      if (action === 'step') {
        const e = structuredClone(state.latest); e.sequence++;
        if (e.sequence === 1) { e.kind = 1; e.name = 'CREATE'; e.choiceTop = 1; e.choices = [{ vertex: 0, remaining: e.domains[0] & ~1, mark: 0, propagated: e.propagated }] }
        else { e.kind = 3; e.name = 'WRITE'; e.trail = [{ vertex: 0, old: e.domains[0], level: 1 }]; e.trailTop = 1; e.domains[0] = 1 }
        records.set(e.sequence, structuredClone(e));
        state = { ...state, latest: e, history: [...state.history, { sequence: e.sequence, kind: e.kind, name: e.name, count: 0, choiceTop: e.choiceTop, trailTop: e.trailTop }] };
      }
      data = state;
    } else if (path.startsWith('/api/events/')) data = records.get(Number(path.split('/').at(-1)));
    return new Response(JSON.stringify(data), { status: 200, headers: { 'Content-Type': 'application/json' } });
  }));
  store = makeStore();
});
afterEach(() => { store.dispatch(api.util.resetApiState()); vi.unstubAllGlobals() });
const mount = () => render(<Provider store={store}><App /></Provider>);

it('loads a preset, steps, selects trail records, and inspects history safely', async () => {
  const user = userEvent.setup(); mount();
  await screen.findByText('FPGA · live device');
  await user.click(screen.getByRole('button', { name: 'Triangle · 2 colors' }));
  await user.click(screen.getByRole('button', { name: 'Load graph' }));
  await waitFor(() => expect(screen.getByRole('button', { name: 'Step event' })).toBeEnabled());
  expect(state.graph.colors).toBe(2);
  await user.click(screen.getByRole('button', { name: 'Step event' }));
  await screen.findByRole('button', { name: /Event 1: CREATE/ });
  await user.click(screen.getByRole('button', { name: 'Step event' }));
  const trail = await screen.findByRole('button', { name: 'Trail entry 0 restores vertex 0' });
  await user.click(trail);
  expect(screen.getByRole('button', { name: 'Vertex 0, candidates 0' })).toHaveAttribute('aria-pressed', 'true');
  await user.click(screen.getByRole('button', { name: /Event 1: CREATE/ }));
  await screen.findByText(/Inspecting recorded event 1/);
  expect(screen.getByRole('button', { name: 'Step event' })).toBeDisabled();
  expect(screen.getByRole('button', { name: 'Load graph' })).toBeDisabled();
  await user.click(screen.getByRole('button', { name: 'Return to live' }));
  expect(screen.getByRole('button', { name: 'Step event' })).toBeEnabled();
});
it('shows graph validation and device errors rather than implying success', async () => {
  const user = userEvent.setup(); mount(); await screen.findByText('FPGA · live device');
  await user.clear(screen.getByLabelText('Edges')); await user.type(screen.getByLabelText('Edges'), '0-0');
  await user.click(screen.getByRole('button', { name: 'Load graph' }));
  expect(await screen.findByRole('alert')).toHaveTextContent('distinct vertices');
  expect(state.loaded).toBe(false);
  await user.click(screen.getByRole('button', { name: 'Path · 2 colors' }));
  await user.click(screen.getByRole('button', { name: 'Load graph' }));
  await waitFor(() => expect(screen.getByRole('button', { name: 'Run' })).toBeEnabled());
  failRun = true; await user.click(screen.getByRole('button', { name: 'Run' }));
  expect(await screen.findByRole('alert')).toHaveTextContent('USB cable disconnected');
});
it('identifies the software simulator explicitly', async () => {
  state.engine = 'model'; mount(); expect(await screen.findByText('MODEL · software simulator')).toBeVisible();
});

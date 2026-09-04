import { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { skipToken } from '@reduxjs/toolkit/query';
import { api, ui, type AppDispatch, type RootState } from './store';
import { candidates, edgeText, errorText, hex, palette, parseGraph, presets } from './graph';
import { GraphView } from './GraphView';
import type { Action, Graph, State } from './types';

export default function App() {
  const dispatch = useDispatch<AppDispatch>(); const view = useSelector((s: RootState) => s.view);
  const { data: state, error: connectionError, isLoading } = api.useStateQuery(undefined, { pollingInterval: 250 });
  const [load, loadStatus] = api.useLoadMutation(); const [control, controlStatus] = api.useControlMutation();
  const historical = api.useEventQuery(view.sequence !== null && state ? { sequence: view.sequence, generation: state.generation } : skipToken);
  const [vertices, setVertices] = useState(3), [colors, setColors] = useState(3), [edges, setEdges] = useState('0-1, 1-2, 0-2'), [firstOnly, setFirstOnly] = useState(false);
  const [message, setMessage] = useState('');
  useEffect(() => { dispatch(ui.actions.reset()) }, [state?.generation, dispatch]);
  const busy = loadStatus.isLoading || controlStatus.isLoading;
  const inspecting = view.sequence !== null;
  const snapshot = inspecting ? historical.currentData : state?.latest;
  const terminal = state?.latest.kind === 10 || state?.latest.kind === 11;
  const disabled = !state?.loaded || !!connectionError || busy || inspecting;
  const choosePreset = (g: Graph) => { setVertices(g.vertices); setColors(g.colors); setEdges(edgeText(g)); setFirstOnly(g.firstOnly); setMessage('') };
  const accept = (next: State) => { dispatch(api.util.upsertQueryData('state', undefined, next)); setMessage('') };
  const action = async (name: Action) => { try { accept(await control(name).unwrap()) } catch (error) { setMessage(errorText(error)) } };
  const submit = async () => { try { accept(await load(parseGraph(vertices, colors, edges, firstOnly)).unwrap()) } catch (error) { setMessage(errorText(error)) } };
  const status = state?.error ? 'LINK ERROR' : state?.running ? 'RUNNING' : state?.latest.kind === 11 ? 'FAULT' : state?.latest.kind === 10 ? state.latest.count === 0 ? 'UNSATISFIABLE' : 'COMPLETE' : state?.loaded ? 'PAUSED' : 'NO GRAPH';
  return <div className="microscope">
    <header className="instrument-header">
      <div><div className="eyebrow">GATEMATE / LAB 03</div><h1>Graph search microscope</h1><p>Inspect propagation, choices, and reversible state.</p></div>
      <div className="source"><span className={`source-dot ${connectionError ? 'offline' : ''}`} />{state?.engine === 'serial' ? 'FPGA · live device' : state?.engine === 'model' ? 'MODEL · software simulator' : 'Connecting to instrument'}</div>
    </header>
    <main className="container-fluid px-4 pb-4">
      {connectionError && <div className="alert alert-danger mt-3" role="alert">Connection unavailable. The last displayed state may be stale. {errorText(connectionError)}</div>}
      {(message || state?.error) && <div className="alert alert-danger mt-3" role="alert">{message || state?.error}</div>}
      <div className="toolbar my-3">
        <div className="d-flex gap-2 flex-wrap">
          <button className="btn btn-dark" disabled={disabled || state?.running || terminal || !!state?.error} onClick={() => void action('step')}>Step event</button>
          <button className="btn btn-success" disabled={disabled || state?.running || terminal || !!state?.error} onClick={() => void action('run')}>Run</button>
          <button className="btn btn-outline-dark" disabled={disabled || !state?.running} onClick={() => void action('pause')}>Pause</button>
          <button className="btn btn-outline-dark" disabled={disabled || state?.running} onClick={() => void action('reset')}>Reset graph</button>
        </div>
        <div className="status-block" aria-live="polite"><span className="status-label">{status}</span><span>event {state?.latest.sequence ?? 0}</span><span>{state?.latest.count ?? 0} accepted</span></div>
      </div>
      {inspecting && <div className="alert alert-warning d-flex justify-content-between align-items-center"><span>Inspecting recorded event {view.sequence}. Device controls require the live view.</span><button className="btn btn-sm btn-dark" onClick={() => dispatch(ui.actions.live())}>Return to live</button></div>}
      <div className="workspace">
        <aside className="panel editor-panel">
          <div className="section-number">01 / CONFIGURE</div><h2>Graph input</h2><p className="text-secondary small">Vertices are numbered from 0. An edge requires different endpoint colors.</p>
          <div className="preset-grid">{presets.map((p) => <button key={p.name} className="btn btn-sm btn-outline-secondary" onClick={() => choosePreset(p.graph)}>{p.name}</button>)}</div>
          <form onSubmit={(e) => { e.preventDefault(); void submit() }}>
            <div className="row g-2 my-2"><label className="col-6 form-label">Vertices<input className="form-control" type="number" min="1" max="8" value={vertices} onChange={(e) => setVertices(Number(e.target.value))} /></label><label className="col-6 form-label">Colors<input className="form-control" type="number" min="1" max="8" value={colors} onChange={(e) => setColors(Number(e.target.value))} /></label></div>
            <label className="form-label w-100">Edges<textarea className="form-control font-monospace" rows={3} value={edges} onChange={(e) => setEdges(e.target.value)} aria-describedby="edge-help" /></label><div id="edge-help" className="small text-secondary mb-3">Example: 0-1, 1-2. Leave empty for isolated vertices.</div>
            <label className="form-check mb-3"><input className="form-check-input" type="checkbox" checked={firstOnly} onChange={(e) => setFirstOnly(e.target.checked)} /><span className="form-check-label">Stop after first solution</span></label>
            <button className="btn btn-success w-100" disabled={busy || state?.running || !!connectionError || isLoading}>Load graph</button>
          </form>
          {state?.loaded && <div className="loaded-note"><strong>Loaded configuration</strong><br />{state.graph.vertices} vertices · {state.graph.colors} colors<br />{state.graph.firstOnly ? 'First solution only' : 'Enumerate all labeled colorings'}<br /><span className="text-secondary">Run generation {state.generation}</span></div>}
          <p className="small text-secondary mt-3 mb-0">Loading replaces the current search. Preset and form edits take effect only when loaded.</p>
        </aside>
        <section className="panel graph-panel">
          <div className="section-number">02 / OBSERVE</div><div className="d-flex justify-content-between"><h2>Domains & propagation</h2><span className="event-badge">{snapshot?.name || 'READY'}</span></div>
          {historical.error ? <div className="alert alert-warning">This event is no longer retained. Return to live to continue.</div> : !state?.loaded ? <div className="empty-view">Load a graph to inspect its search state.</div> : !snapshot ? <div className="empty-view">Loading recorded state…</div> : <>
            <GraphView graph={state.graph} snapshot={snapshot} selected={view.vertex} onSelect={(v) => dispatch(ui.actions.selectVertex(v))} />
            <div className="table-responsive"><table className="table table-sm domain-table"><thead><tr><th>Vertex</th><th>Mask</th><th>Candidate colors</th><th>Propagated</th></tr></thead><tbody>{snapshot.domains.slice(0, state.graph.vertices).map((d, v) => <tr key={v} className={view.vertex === v ? 'table-active' : ''}><td><button className="btn btn-sm p-0 fw-bold" onClick={() => dispatch(ui.actions.selectVertex(v))}>v{v}</button></td><td className="font-monospace">0x{hex(d)}</td><td>{candidates(d).map((c) => <span key={c} className="color-chip" style={{ backgroundColor: palette[c] }}>C{c}</span>)}{!d && <span className="text-danger fw-bold">Contradiction</span>}</td><td>{snapshot.propagated & (1 << v) ? 'Yes' : 'No'}</td></tr>)}</tbody></table></div>
            {snapshot.count > 0 && <p className="small mb-0"><strong>Latest accepted coloring:</strong> <span className="font-monospace">[{Array.from({ length: state.graph.vertices }, (_, v) => (snapshot.result >> (3 * v)) & 7).join(', ')}]</span> · {snapshot.count} accepted at this event</p>}
          </>}
        </section>
        <aside className="panel history-panel">
          <div className="section-number">03 / INSPECT</div><h2>Event history</h2><p className="small text-secondary">Newest first. Up to 256 immutable snapshots; viewing history does not reverse the FPGA.</p>
          <div className="timeline">{state?.history.slice().reverse().map((e) => <button key={e.sequence} aria-label={`Event ${e.sequence}: ${e.name}`} className={`timeline-event ${view.sequence === e.sequence ? 'selected' : ''}`} aria-pressed={view.sequence === e.sequence} onClick={() => dispatch(ui.actions.inspect(e.sequence))}><span className="font-monospace">{e.sequence}</span><strong>{e.name}</strong><small>C {e.choiceTop} / T {e.trailTop}</small></button>)}{!state?.history.length && <p className="text-secondary small">Step or run to record events.</p>}</div>
        </aside>
        <section className="panel memory-panel">
          <div className="section-number">04 / RECOVER</div><div className="row g-4">
            <div className="col-lg-5"><h2>Choice stack <span className="count-badge">{snapshot?.choiceTop ?? 0}</span></h2><p className="small text-secondary">Each checkpoint retains alternatives and a restoration mark.</p><div className="table-responsive"><table className="table table-sm"><thead><tr><th>Level</th><th>Vertex</th><th>Remaining</th><th>Mark</th><th>Saved P</th></tr></thead><tbody>{snapshot?.choices.map((c, i) => <tr key={i}><td>{i + 1}</td><td>v{c.vertex}</td><td>{candidates(c.remaining).map((v) => `C${v}`).join(' ') || 'Exhausted'}</td><td>{c.mark}</td><td className="font-monospace">{hex(c.propagated)}</td></tr>)}</tbody></table></div>{!snapshot?.choices.length && <p className="small text-secondary">No live alternatives.</p>}</div>
            <div className="col-lg-7"><h2>Mutation trail <span className="count-badge">{snapshot?.trailTop ?? 0}</span></h2><p className="small text-secondary">Select a record to locate its vertex. Undo restores old masks in reverse order. Base: {snapshot?.base ?? 0}.</p><div className="trail-grid">{snapshot?.trail.map((entry, i) => <button key={i} className={`trail-entry ${i < snapshot.base ? 'committed' : ''} ${view.vertex === entry.vertex ? 'selected' : ''}`} aria-label={`Trail entry ${i} restores vertex ${entry.vertex}`} onClick={() => dispatch(ui.actions.selectVertex(entry.vertex))}><span>#{i} · level {entry.level}</span><strong>v{entry.vertex} ← 0x{hex(entry.old)}</strong><small>{i < snapshot.base ? 'Retained after cut' : `Restore ${candidates(entry.old).map((c) => `C${c}`).join(' ')}`}</small></button>)}</div>{!snapshot?.trail.length && <p className="small text-secondary">No logged mutations.</p>}</div>
          </div>
        </section>
      </div>
      <footer>Deterministic search · singleton propagation · synchronous rollback · {state?.engine === 'serial' ? 'events from physical GateMate hardware' : 'events from the Go reference model'}</footer>
    </main>
  </div>;
}

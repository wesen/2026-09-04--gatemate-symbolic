---
Title: Graph Coloring Search Microscope - Intern Analysis Design and Implementation Guide
Ticket: GATEMATE-SYMBOLIC-006
Status: active
Topics:
    - fpga
    - gatemate
    - symbolic-computers
    - architecture
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: repo://queens_rollback/rtl/queens_core.sv
      Note: Source of deterministic recovery sequencing
    - Path: repo://queens_rollback/tools/queens_model.py
      Note: Prior semantic model and event format
    - Path: repo://symbolic_eval/rtl/sync_sdp_ram.sv
      Note: Synchronous memory timing
    - Path: repo://symbolic_eval/rtl/uart_tx.sv
      Note: Shared transmitter handshake
ExternalSources: []
Summary: Intern guide for a configurable graph-coloring FPGA, lossless semantic stepping, a Go device service, and a React state inspector.
LastUpdated: 2026-09-04T18:26:11.714531969-04:00
WhatFor: ""
WhenToUse: ""
---


# Graph Coloring Search Microscope - Intern Analysis Design and Implementation Guide

## Executive Summary

Build a browser application that loads a small graph into the GateMate FPGA, advances its coloring search through observable semantic events, and displays domain propagation, choices, restoration, and accepted solutions. A Go service owns the serial connection and presents a typed HTTP API. A TypeScript/React frontend uses Redux Toolkit and RTK Query to edit graphs, control execution, and inspect the current or a retained historical state.

The first implementation supports one through eight vertices and one through eight colors. Edges are undirected, self-loops are rejected, and every active vertex initially permits every configured color. Search is deterministic: select the first unresolved vertex and its lowest permitted color; propagate singleton colors to adjacent vertices; undo changes through a mutation trail. Full enumeration and first-solution cut are supported. Historical inspection changes the browser's selected snapshot; it does not reverse the physical FPGA.

The design preserves the previous laboratories as standalone experiments. New graph RTL derives its scheduling and recovery discipline from the proven queens engine, with graph configuration and a genuine event handshake added. The software model is an explicitly labeled simulator and verification reference. It must never be presented as FPGA execution.

## Problem Statement

The current solver computes a fixed geometric problem and exports only solved boards and terminal UART records. Its internal trace is visible in simulation, but the board cannot receive a graph or stop under host control at each mutation. A useful search inspector needs both capabilities. Sampling debug registers asynchronously would produce inconsistent or missing states because propagation is much faster than UART transmission.

The main engineering requirement is therefore ownership at event boundaries. When the FPGA publishes an event, its supporting semantic state must remain stable until the host accepts that event. The host must own one serial exchange at a time. The browser must distinguish the latest device state from a previously captured state and must never infer that a dropped connection means successful completion.

### Acceptance criteria

- Load a validated graph from the browser through Go and UART into a newly configured graph solver without rebuilding the bitstream.
- Step exactly one semantic event; run continuously by repeating the same step exchange; pause without losing an event or corrupting serial framing; reset the loaded graph.
- Display vertices and edges, candidate colors, propagation bits, live choices, trail records, solution count, terminal status, and event history.
- Select a trail entry and identify its affected vertex; select a retained historical event and inspect its complete recorded state.
- Demonstrate a triangle with three colors (six labeled colorings), a triangle with two colors (unsatisfiable), and a path with two colors (two colorings), plus first-only termination.
- Compare Go model and RTL semantic events and verify actual graph loading and stepping on the connected FPGA.
- Serve the built UI and API from one Go binary; keep compiled JS/CSS under `/static/` URLs and use Bootstrap styling.

## Current system and reusable principles

The source baseline is `f6fce4e` in `/home/manuel/code/wesen/2026-09-04--gatemate-symbolic`. The implementation is currently SystemVerilog and Python; no root Go module or frontend package exists. The root Go module introduced by this lab must be the only module in the repository.

| Existing source | Observed behavior and relevance |
|---|---|
| `queens_rollback/rtl/queens_core.sv:106` | Scans zero domains before singleton propagation, then output, then choices. |
| `queens_rollback/rtl/queens_core.sv:117` | Reserves choice and trail capacity before choosing; writes a checkpoint before publishing its top. |
| `queens_rollback/rtl/queens_core.sv:137` | Logs changed domains before applying them; no-ops do not consume trail space. |
| `queens_rollback/rtl/queens_core.sv:177` | Waits for synchronous RAM reads and restores trail records in reverse order. |
| `queens_rollback/rtl/queens_core.sv:220` | Accepts a result before applying first-only cut. |
| `queens_rollback/rtl/queens_top.sv` | Connects reset, core, formatter, UART, and LED; receive pin is unused. |
| `symbolic_eval/rtl/sync_sdp_ram.sv` | Registered synchronous read output and a separate write port. |
| `symbolic_eval/rtl/uart_tx.sv` | Ready/start/data transmitter with 8N1 framing and a rounded baud divider. |
| `queens_rollback/tools/queens_model.py` | Executable event semantics and complete checkpoint/trail state. |
| `queens_rollback/sim/tb_queens.sv` | Event and between-event state checks, reset injection, corruption checks. |

The previous physical result is 92 exact ordered queen boards from both recovery modes. This establishes a useful recovery baseline, not proof that graph input or trace pacing works. Those are new mechanisms and require their own tests.

## Foundations: graph coloring as domain propagation

A graph has vertices `0..n-1` and undirected edges. A valid coloring assigns each vertex a value in `0..k-1` such that connected vertices have different values. The adjacency row `A[v]` is an eight-bit mask; bit `u` identifies an edge from `v` to `u`. For an undirected graph, `A[v][u]` must equal `A[u][v]`. Bits referring to inactive vertices and all diagonal bits must be zero.

The domain `D[v]` is an eight-bit set of candidate colors. Active domains begin at `(1 << k) - 1`. Inactive physical slots are initialized to singleton `01` and marked propagated, so fixed-size hardware scans do not treat them as unresolved vertices. They are excluded from the displayed result and from graph constraints.

If `D[v]` is singleton and `v` has not propagated, remove that singleton color from each neighbor's domain. Non-neighbors are unchanged. A zero domain is a contradiction. The propagated bitmap records completion of the entire source scan, not simply that a vertex is singleton.

```text
propagate(v):
    require D[v] is singleton
    for u in increasing vertex order:
        if adjacency[v][u]:
            mutate(u, D[u] AND NOT D[v])
            if D[u] == 0: return contradiction
    propagated[v] = true

choose():
    v = first non-singleton active vertex
    color = lowest set bit of D[v]
    reserve checkpoint and trail capacity
    save(v, D[v] without color, trail_top, propagated)
    mutate(v, color)
```

This is singleton propagation, not a general consistency algorithm. Colors remain labeled; symmetry-equivalent assignments are still separate solutions. A three-vertex triangle with three colors has six outputs. With two colors, the same triangle eventually exhausts all alternatives and reports completion with zero outputs.

## Architecture and state ownership

```mermaid
flowchart LR
    Editor[React graph editor] --> API[Go HTTP API]
    Controls[Step run pause reset] --> API
    API --> Session[Serialized session owner]
    Session --> Model[Explicit Go simulator]
    Session --> Serial[Serial transport and codec]
    Serial --> RX[FPGA UART receiver and command parser]
    RX --> Core[Graph search core]
    Core --> Hold[Stable pending semantic event]
    Hold --> TX[Event formatter and UART transmitter]
    TX --> Serial
    Session --> Projection[Checked state reconstruction and history]
    Projection --> Views[Graph domains choices trail and timeline]
```

The service selects its engine at startup: `--engine model` for a simulator or `--engine serial --device /dev/ttyACM0` for hardware. The API always reports the selected engine prominently. The browser cannot choose arbitrary filesystem device paths. A single session owns the engine; this lab is a local instrument, not a multi-user hosted service.

### Semantic state versus implementation state

Semantic state comprises domain masks, propagation bits, live choice/trail records and tops, trail base, accepted solution count, pending result, fault, and completion. Controller states, staging registers, UART bit counters, and temporary parser bytes are implementation state. The GUI receives complete semantic observations at events. It does not require a dump of every register on every clock.

The graph core publishes `trace_valid` and accepts `trace_ready`. When valid is asserted without ready, all semantic state and the event metadata remain stable. Its sequential controller and both RAM write enables must be gated by the same advance condition. Gating the controller alone would leave a write-enable state repeatedly writing RAM while the event is blocked.

```text
advance = NOT trace_valid OR trace_ready
if reset: initialize semantic state from accepted graph configuration
else if advance:
    clear old trace_valid
    perform one controller state transition
    if a semantic event occurs: publish trace_valid and event kind

choice_ram_write_enable = advance AND checkpoint_write_state
trail_ram_write_enable  = advance AND trail_write_state
```

An acknowledged event permits the core to continue until its next event. The transport does not grant a large free-running window. This bounds buffering and makes pause independent of FPGA event rate. While an event is being serialized, the core may prepare the next one and then hold it; the browser's accepted sequence still advances one event per step exchange.

## Record and trace contracts

The graph trail uses the established formats: choice metadata is 40 bits (column 3, remaining mask 8, mark 7, saved propagated 8, reserved 14), and a trail record is 20 bits (vertex 3, old mask 8, choice level 5, reserved 4). Physical depths remain eight choices and 64 trail entries. Logical capacities are elaboration parameters for testing. Runtime graph input cannot request arbitrary memory sizes.

**Implementation finding:** Unlike queens, a one-color graph begins with singleton domains and can perform root propagation before creating a checkpoint. Its valid trail entries have choice level zero. Graph integrity checks must permit zero when it matches the current root depth; copying the queens nonzero-level guard would incorrectly reject this valid case. A root contradiction completes with zero solutions and no checkpoint to restore.

| Event code | Name | Projection update |
|---|---|---|
| 1 | CREATE | Append the published choice metadata. |
| 2 | UPDATE | Replace the newest choice metadata. |
| 3 | WRITE | Append the published old-domain trail record. |
| 4 | PROPAGATED | Install the reported propagated bitmap. |
| 5 | CONTRADICTION | Preserve records and report the zero domain. |
| 6 | RESTORE | Remove the newest trail record; reported domains contain its old mask. |
| 7 | RESTORED | Domains and propagated bits now represent the complete checkpoint. |
| 8 | POP | Remove the newest choice. |
| 9 | OUTPUT | Increment accepted count; first-only may clear choices and set base. |
| 10 | COMPLETE | Mark terminal success, including zero-solution exhaustion. |
| 11 | FAULT | Mark terminal fault; protect the attempted operation's state. |

Every event also carries packed domains, propagated bitmap, tops/base, count, result, and fault code. The newest choice word is meaningful for CREATE and UPDATE; the trail word is meaningful for WRITE and RESTORE. Other events may carry stale staging words, which the projection must ignore. This distinction allows exact architecture checking without pretending that internal staging has semantic meaning.

The host reconstructs full live stacks from these deltas and checks lengths against reported tops. A missing sequence number, invalid checksum, impossible top transition, or invalid restoration stops the run and records a transport/projection error. It must not silently fill gaps from the model. Historical snapshots are deep copies, so later restoration cannot mutate earlier GUI views.

## Serial wire protocol

Use a bounded ASCII hexadecimal request/response protocol at nominal 115200 baud, 8N1. Hex encoding is simple to inspect in captures and avoids delimiter escaping. The host permits one outstanding exchange. There is no automatic retry of a side-effecting command after timeout because the device may already have accepted it.

### Commands

| Command | Bytes | Effect and response |
|---|---|---|
| Load | `L` + 24 hex digits + LF | Twelve bytes encoded as hex: n, k, first-only, eight adjacency rows, XOR checksum. Validate completely, then atomically install configuration/reset; reply `A` + LF. |
| Step | `S` + LF | Accept one pending event and send its event record. |
| Reset | `R` + LF | Reset the current accepted graph and event numbering; reply `A` + LF. |

The load checksum is the XOR of the first eleven decoded bytes. `first-only` must be zero or one. Both Go and FPGA validate bounds, symmetry, self-loops, and inactive bits. A malformed or partial input must not replace the accepted graph. The parser bounds its line length and abandons a partial command after an idle timeout. The UART receiver samples the start bit and data centers, checks stop framing, and synchronizes the asynchronous input.

Errors use `!` followed by a two-digit code and LF. Reserve separate codes for syntax/framing, invalid graph, and checksum. Terminal stepping can return `Z` + LF once no further event exists. The host normally prevents this by recognizing COMPLETE or FAULT. Commands sent while a response is active are outside the stop-and-wait contract; the host serial owner prevents them.

### Event record

An event is `E`, 66 hexadecimal digits representing the 33-byte payload below, two hexadecimal XOR-checksum digits, and LF: 70 bytes total. Numeric fields are serialized most significant byte first. The checksum is XOR of all payload bytes. This detects simple line corruption; it is not an authentication mechanism or a strong error-correcting code.

| Payload field | Width | Notes |
|---|---:|---|
| Sequence | 32 bits | Starts at one after load/reset; strictly increments for delivered events. |
| Event kind | 8 bits | Codes 1–11. |
| Domains | 64 bits | Vertex zero in least significant byte. |
| Propagated | 8 bits | Includes pre-propagated inactive slots. |
| Choice top | 8 bits | Logical range 0–8. |
| Trail top | 8 bits | Logical range 0–64. |
| Trail base | 8 bits | No greater than trail top. |
| Accepted count | 32 bits | Permanent within a run. |
| Result | 24 bits | Three bits per vertex; ignore inactive slots. |
| Fault | 8 bits | Zero or existing solver fault code. |
| Choice word | 40 bits | Meaningful for CREATE/UPDATE. |
| Trail word | 24 bits | Low 20 bits contain the record; upper four are zero. |

At 10 MHz, the UART divider is 87 clocks per bit. Seventy bytes require at least 60,900 clocks of serial framing, about 6.09 ms, plus command and controller overhead. The microscope intentionally prioritizes complete observability over uninstrumented solver throughput.

## Go packages and API boundaries

Create module `github.com/wesen/2026-09-04--gatemate-symbolic` at repository root. Keep reusable domain types under `pkg/microscope`, HTTP/session code under `internal/microscope`, and the command under `cmd/search-microscope`. The frontend lives in `web/`. No nested go.mod files are introduced.

```go
type Graph struct {
    Vertices int      `json:"vertices"`
    Colors   int      `json:"colors"`
    Edges    [][2]int `json:"edges"`
    FirstOnly bool    `json:"firstOnly"`
}

type Engine interface {
    Load(context.Context, Graph) error
    Step(context.Context) (Event, error)
    Reset(context.Context) error
    Close() error
}
```

Graph validation constructs symmetric eight-row adjacency and rejects duplicates or normalizes them consistently before serial encoding. The model implements the same Engine contract as serial hardware, but maintains its own explicit semantic state machine. The independent test oracle enumerates assignments and checks edge inequalities directly, without reusing propagation or rollback.

Serial reads need bounded timeouts and cancellation checks. Use `go.bug.st/serial` with `Open`, a baud `Mode`, `SetReadTimeout`, and close on shutdown. Short reads and short writes are normal transport conditions and must be handled. After a timeout or malformed response, mark the link unsynchronized and require explicit reset/reload recovery; do not keep issuing steps into an unknown response boundary.

### HTTP API

| Endpoint | Contract |
|---|---|
| `GET /api/state` | Engine identity, accepted graph, generation, running/terminal/error status, latest snapshot, retained history summaries. |
| `POST /api/graph` | Strict JSON Graph. Validate before touching the engine; load atomically and clear prior history on success. |
| `POST /api/control` | `{ "action": "step" | "run" | "pause" | "reset" }`. Reject incompatible actions with a structured error. |
| `GET /api/events/{sequence}` | Deep immutable snapshot for a retained event; return 404 when no longer retained. |
| `GET /` | Application HTML. |
| `GET /static/{path...}` | Embedded JS/CSS and other compiled assets. |

Use `http.ServeMux`, bounded request bodies, strict JSON decoding, and JSON error responses. Unknown API paths must return API errors or 404, not the application's HTML. Default listening address is loopback. The device path is a CLI field, not an HTTP argument. Read-only status requests never drive the solver.

The session serializes all engine exchanges. Run executes the same Step operation repeatedly in an errgroup-managed worker, with a modest pacing interval. Pause stops scheduling new exchanges and waits for any already-started bounded exchange to finish. Holding a session lock over a bounded exchange is acceptable for this single-device lab; an unbounded read while holding it is not. Shutdown stops the worker, closes the device, and shuts down HTTP with a deadline.

Retain at most 256 complete event snapshots in memory. State responses include short summaries rather than every full snapshot. Historical lookup is explicitly bounded: an evicted sequence is unavailable. A generation counter distinguishes runs after load/reset, and the frontend clears its historical selection when generation changes.

The CLI uses Glazed fields and a BareCommand for the long-running server, with flags for listen address, engine, device, and log level. Zerolog records load/control/transport failures without logging device authentication secrets (none are needed here). Contexts propagate through engine operations; background work uses errgroup. Build and analyzer versions are pinned to the selected Glazed module.

## React interface and state flow

The interface has a graph editor and a live inspector. Presets include triangle, path, square, and a graph that requires rollback. The editor exposes vertex count, color count, first-only mode, and an edge list or adjacency toggles. Invalid input gets a concrete error before device state is changed. The accepted graph remains visible separately from an unsubmitted edit.

An SVG graph places vertices around a circle and draws edges underneath. A singleton domain fills the vertex with its assigned color; unresolved vertices show candidate swatches; zero domains display a contradiction marker. Color names/numbers accompany swatches so color is not the only information channel. A domain table shows numeric masks and propagated status.

The choice table shows variable, remaining candidates, mark, and saved propagation bits. The trail table shows index, vertex, old mask, and choice level. Clicking a trail row selects that vertex in the graph and explains what restoration would install. Cut history below the base is visibly distinguished from undoable entries.

The timeline lists sequence, event name, and compact state changes. Selecting an event fetches its retained snapshot and displays a historical-inspection label. A Return to live control restores the latest view. Step, Run, Pause, and Reset always address the current device session, not the selected historical snapshot; controls must make that distinction explicit.

Use one RTK Query API slice for state polling, control/load mutations, and historical lookup. Use a Redux UI slice for selected sequence and vertex. Polling observes current server state; it does not create events. Mutations invalidate the state cache. Failed HTTP requests must show their errors instead of leaving the UI apparently running. Bootstrap provides layout and accessible controls; source files use TypeScript throughout.

## Decision records

### Decision: bounded graph coloring first

- **Context:** The prior suggestion included a generic finite-domain engine and a live microscope.
- **Options considered:** Arbitrary compatibility tables, graph coloring only, or visualization of fixed queens.
- **Decision:** Implement runtime graph coloring with up to eight vertices/colors and complete event inspection.
- **Rationale:** This adds a new programmable problem and physical input while retaining a small, testable propagation rule.
- **Consequences:** Arbitrary binary constraints and larger memories remain future extensions, not hidden features.
- **Status:** accepted.

### Decision: stop-and-wait semantic events

- **Context:** UART cannot carry every event at core clock rate.
- **Options considered:** Lossy sampling, a large event FIFO, or backpressure at every semantic boundary.
- **Decision:** The core holds one pending event until the host requests it.
- **Rationale:** Complete history and deterministic stepping matter more than peak throughput for this instrument.
- **Consequences:** Runtime includes observation latency. Every RAM write enable must obey the advance condition.
- **Status:** accepted.

### Decision: bounded server snapshots and historical inspection

- **Context:** Full stack/trail views need more information than a result stream, but unbounded history is unsafe for enumeration.
- **Options considered:** Full RAM dumps per event, delta reconstruction with bounded snapshots, or browser-only reconstruction.
- **Decision:** Go validates deltas and retains 256 full snapshots; the browser requests selected history.
- **Rationale:** One checked projection serves all views and bounds memory independently of run duration.
- **Consequences:** Old history can be evicted, and the interface must report that explicitly. Historical viewing is not reverse execution.
- **Status:** accepted.

### Decision: preserve prior laboratory implementations

- **Context:** The measured queen backend is valuable evidence and does not have trace backpressure or graph input.
- **Options considered:** Rewrite queens as a configurable solver or create a separate graph laboratory with direct reuse of low-level components.
- **Decision:** Add graph-specific RTL and retain earlier laboratories unchanged.
- **Rationale:** The new experiment has a distinct input/observation contract. Shared RAM/reset/UART modules are reused directly; no compatibility shim is introduced.
- **Consequences:** Some controller structure is intentionally repeated across laboratory snapshots. Later generalization should be a separately justified refactor.
- **Status:** accepted.

## Proposed Solution

The preceding contracts define the implementation. Start with the Go model and protocol codec so the same typed events can feed unit tests, serial transport, HTTP, and the browser. Adapt the graph core with runtime masks and adjacency, then add real UART reception and the bounded command/event formatter. Verify the wire path before treating the GUI as a debugging tool.

```text
HTTP step:
    acquire serialized session ownership
    require loaded, not running, not terminal, synchronized transport
    event = engine.Step(context with bounded timeout)
    require event.sequence == previous.sequence + 1
    next = checked_projection(previous, event)
    append immutable next to bounded history
    publish latest state

HTTP load:
    validate graph without device mutation
    acquire serialized session ownership; require paused
    engine.Load(graph)
    only after acknowledgement:
        install graph, increment generation, clear history and terminal/error
```

## Design Decisions

The accepted decisions above are implementation constraints. If measurements force a change, update this document with the final contract and record the evidence in the diary rather than leaving code and guide inconsistent.

## Alternatives Considered

A browser-only simulator would make a useful teaching tool but would not satisfy graph loading and live FPGA inspection. A generic compatibility-table engine would broaden problem scope but add table memory and configuration complexity before the input/trace transport is proven. A high-speed trace FIFO would permit short uninterrupted runs, but capacity overflow would require either stalling anyway or losing precisely the evidence the microscope is intended to preserve.

## Implementation Plan

### P1 — Design and delivery

Create this guide, relate existing evidence, archive consulted APIs under sources using Defuddle, initialize the diary and task list, print plan/start slips, run docmgr doctor, and dry-run then upload the guide to reMarkable. Commit the concrete design before beginning implementation.

### P2 — Go model and serial contract

Add root go.mod; `pkg/microscope/graph.go`, `model.go`, `protocol.go`, and `projection.go`; add oracle and codec/projection tests. Exercise satisfiable and unsatisfiable graphs, all labeled outputs, first-only, shallow capacity, zero-domain logging, malformed input, checksum corruption, and sequence gaps. Commit the executable contract.

### P3 — FPGA graph execution and transport

Add `graph_microscope/rtl/graph_core.sv`, `uart_rx.sv`, `graph_link.sv`, and `graph_top.sv`; add simulation benches and build scripts. Compare semantic events with the Go model, hold trace readiness low, test invalid loads and reset, synthesize/route at 10 MHz, and load real graphs over UART. Record complete captured events and timing/resource evidence. Commit hardware and tests.

### P4 — Go service

Add serial Engine implementation, session owner, HTTP handlers, embedded assets contract, and Glazed server command. Test cancellation, bounded exchanges, load validation, run/pause sequencing, retained snapshots, error responses, and concurrent requests with the race detector. Commit a usable API before building the final UI.

### P5 — React microscope

Add pnpm/Vite/React/TypeScript, Bootstrap, Redux and RTK Query; implement graph editor, controls, SVG view, masks, choice/trail tables, and historical event selection. Add focused interaction tests, type checking, and deterministic embedded build generation. Commit the interface.

### P6 — Integrated evidence and handoff

Run the Go tests/race checks/build/vet/analyzer, frontend tests/build, graph RTL simulations, and the existing queens regression suite where shared infrastructure is involved. Start the server in tmux and exercise the real browser. Demonstrate physical triangle/path/unsatisfiable and first-only runs through the service; inspect selected trail/history entries in the UI. Update final docs and diary, archive screenshots/evidence, print completion, commit and push.

Every phase has a printed start and completion slip, with immutable commit references on completion. The overall plan adds one slip, for thirteen total. Scripts belong in scripts directories; downloaded resources belong in sources; build outputs remain ignored.

## Validation strategy and evidence limits

The independent oracle enumerates color assignments and directly checks every edge. The model supplies the exact semantic order. RTL tests compare all architectural fields at events plus meaningful choice/trail deltas, then test stability between accepted events. UART simulation verifies encoded bytes and malformed commands. Physical capture establishes configuration, clock/reset, pin mapping, serial framing, and real solver events.

HTTP tests use both a deterministic model and a controlled failing engine to exercise transport errors without hardware. Concurrent controls must never cause overlapping serial exchanges. Browser tests check what the user can see and do, including clear engine identity, errors, history selection, and return to live. A software-mode browser test is not physical FPGA evidence; records must identify their engine.

Hardware acceptance includes six triangle colorings with three colors, zero with two, two path colorings with two, and one with first-only. Timing must pass the actual 10 MHz constraint. Resource reports retain tool units such as RAM_HALF and CPE_LT. Instrumentation adds cost; compare with the old solver only with that difference stated.

## Open Questions

No user decision blocks this scoped implementation. Remaining engineering risks are UART resynchronization after a truncated command, event-stall gating of memory writes, input validation in both Go and RTL, and loss of ownership during reset. These require tests, not speculative workarounds. The design does not promise arbitrary-size graphs, multiple concurrent devices, hardware reverse stepping, external-network authentication, or unbounded event retention.

Any deviation discovered during implementation must be recorded with its effect on the public contract. A physical-board failure must not be hidden by switching the service to model mode. A failed upload must not be described as reMarkable delivery. A queued print job must not be described as printed without the printer receipt.

## References

- The existing queens core, model, UART wrapper, shared RAM, and validation evidence are mapped in the current-system table above. The detailed prior technical report is ticket GATEMATE-SYMBOLIC-004 reference/02-inside-the-eight-queens-rollback-engine-technical-project-report.md.
- [RTK Query overview](https://redux-toolkit.js.org/rtk-query/overview): API slices define fetching/mutation endpoints, generated React hooks, and Redux middleware integration. Archived as sources/rtk-query.md.
- [Go serial package API](https://pkg.go.dev/go.bug.st/serial): Open, Mode, Port read/write/close, and bounded read timeouts. Archived as sources/go-serial.md.
- [Go HTTP ServeMux API](https://pkg.go.dev/net/http#ServeMux): method/path patterns and request routing. Pattern specificity determines matching; route registration order is not used as application precedence. Archived as sources/go-http.md.
- Local Glazed APIs are inspected under /home/manuel/code/wesen/go-go-golems/glazed/pkg/cmds and pkg/cli before command authoring. The chosen module version must match the analyzer version.

#!/usr/bin/env python3
"""Reconcile the original design with measured implementation and API contracts."""
from pathlib import Path

ticket = Path(__file__).resolve().parents[1]
guide = next((ticket / 'design-doc').glob('*.md'))
text = guide.read_text()
text = text.replace('Build a browser application that loads', 'This implemented laboratory provides a browser application that loads', 1)
text = text.replace('The implementation is currently SystemVerilog and Python; no root Go module or frontend package exists.', 'At that baseline the implementation was SystemVerilog and Python; this ticket added the root Go module and frontend package.')
text = text.replace('rejects duplicates or normalizes them consistently before serial encoding', 'rejects duplicate edges, including reversed duplicates, before serial encoding')
text = text.replace('`GET /api/events/{sequence}` | Deep immutable snapshot for a retained event; return 404 when no longer retained.', '`GET /api/events/{sequence}?generation=N` | Deep immutable snapshot; generation is required (400 if absent), wrong generation returns 409, and evicted history returns 404.')
text = text.replace('Embedded JS/CSS and other compiled assets.', 'Only the compiled app.js and app.css assets.')
text = text.replace('an edge list or adjacency toggles', 'a comma/whitespace-separated edge list')
text = text.replace('Step, Run, Pause, and Reset always address the current device session, not the selected historical snapshot; controls must make that distinction explicit.', 'Step, Run, Pause, Reset, and Load are disabled while inspecting history. Return to live must be selected before mutating the current session.')
text = text.replace('Mutations invalidate the state cache.', 'Successful mutations install the returned state into the query cache immediately; polling refreshes it every 250 ms.')
marker = '\n## Implemented system and measured walkthrough\n'
text = text.split(marker)[0]
text += marker + '''
The implementation is complete across RTL, Go, and React. The original design and phase plan above explain the intended contracts; this section records their concrete realization and verification. The standalone earlier laboratories remain available. The root module now requires Go 1.26.8 after the final vulnerability scan identified advisories in the installed 1.26.1 toolchain.

### Read the source in dependency order

An intern should first read `pkg/microscope/graph.go`, which defines legal inputs, event codes, packed words, and the Engine interface. Then read `model.go` to understand semantic execution independently of clock cycles. `projection.go` explains what evidence a device event must carry for the service to reconstruct a valid state. `protocol.go` maps those values to UART bytes, and `serial.go` implements bounded exchanges and explicit recovery.

Next read `graph_microscope/rtl/graph_core.sv` beside the model. Its controller introduces extra cycles around synchronous memory reads and writes, but publishes equivalent semantic events. `graph_link.sv` owns graph loading, command parsing, sequence numbering, event serialization, and terminal responses. `uart_rx.sv` samples physical bytes, while `graph_top.sv` joins the core, transport, clock/reset, pins, and indicator.

Finally read `internal/microscope/session.go`, `http.go`, and `web/src/store.ts` before the React view. This order makes ownership clear: the session owns the engine, HTTP calls the session, and React renders server snapshots. `App.tsx` implements interaction and view selection; `GraphView.tsx` renders topology and candidate colors; `graph.ts` validates editor input and defines presets. `types.ts` mirrors the JSON boundary rather than importing hardware encodings into component logic.

| Concrete API | Responsibility and invariant |
|---|---|
| `Graph.Adjacency` | Reject invalid dimensions, endpoints, self-loops, and repeated undirected edges before constructing eight symmetric adjacency rows. |
| `Engine.Load`, `Step`, `Reset`, `Close` | Shared execution boundary for the physical serial engine and explicitly labeled software model. |
| `Serial.exchange` | Write all command bytes, accumulate a bounded response, check framing, and poison synchronization after an ambiguous result. |
| Checked projection `Apply` | Verify sequence and mutation/checkpoint meaning, then construct an independent snapshot. |
| `Session.Control` | Serialize Step/Run/Pause/Reset; stop scheduling on failure or completion. |
| `Session.Event` | Retrieve a retained immutable snapshot only in its matching generation. |
| `NewHandler` | Enforce strict request decoding, bounded bodies, route-specific responses, and static asset allowlisting. |
| RTK Query state/event hooks | Observe latest state and selected history; neither query advances hardware. |

### A concrete triangle execution

Load vertices 0, 1, 2 with edges 0–1, 1–2, 0–2 and three colors. Each active domain starts at hexadecimal 07, representing candidates 0, 1, 2. Inactive slots 3–7 have domain 01 and propagation bits already set, so the initial propagation mask is F8. Inactive slots are internal fixed values and do not add graph vertices or extra solution multiplicity.

Event 1 is CREATE. The engine chooses vertex 0, saves the old propagation mask F8 and trail mark 0, and retains candidates 1 and 2 as alternatives (mask 06). This event publishes the checkpoint before narrowing the domain. Event 2 is WRITE: trail entry 0 records vertex 0, old domain 07, and level 1; the domain becomes 01. The UI therefore shows one checkpoint and one trail entry. Selecting the trail entry highlights vertex 0 and explains that restoration would install 07.

Propagation removes color 0 from the two neighbors. These are separate WRITE events because each change must be recoverable. Once all neighbors have been processed, PROPAGATED records that vertex 0 has been handled. The next unresolved vertex is 1, whose domain is now 06. A second checkpoint chooses color 1 and retains color 2. Narrowing and propagation eventually leave domains 01, 02, 04, yielding the first accepted coloring [0, 1, 2]. OUTPUT increments the permanent accepted count.

In enumeration mode, recovery reads trail records in reverse order until the checkpoint mark is reached, reinstalls its saved propagation mask, and tries the next remaining color. The accepted count does not roll back. In first-only mode, the accepted result is preserved, live choices are cut, and the trail base advances to the retained top. The UI marks those retained records as cut history rather than claiming they remain undoable.

```text
load triangle(3 colors): D = [07,07,07], P = F8
CREATE: choices = [(vertex=0, remaining=06, mark=0, savedP=F8)]
WRITE:  trail = [(vertex=0, old=07, level=1)], D = [01,07,07]
WRITE:  D = [01,06,07]    // neighbor 1 loses color 0
WRITE:  D = [01,06,06]    // neighbor 2 loses color 0
PROPAGATED: P includes vertex 0
... choose vertex 1, propagate, publish OUTPUT ...
enumeration: restore, retry, repeat until no choices remain
first-only:  preserve accepted result, cut choices, COMPLETE
```

This exact first branch was visible through the physical browser. Complete measured runs produced the following results. Event totals include COMPLETE and all semantic writes/recovery events; they are not clock-cycle counts.

| Graph | Physical events | Accepted labeled colorings |
|---|---:|---:|
| Triangle, 3 colors | 86 | 6 |
| Triangle, 2 colors | 26 | 0 |
| Four-vertex path, 2 colors | 32 | 2 |
| Triangle, 3 colors, first-only | 12 | 1 |
| Root contradiction test | 3 | 0 |

### Why event stepping is lossless

The core advances when no event is pending or the transport accepts the pending event. The same condition gates controller transitions and checkpoint/trail write enables. This is essential: freezing only the controller while leaving a RAM write enabled would repeatedly perform an operation during UART transmission. The simulation hold monitor checks architectural stability between accepted events.

```mermaid
sequenceDiagram
    participant UI as React
    participant S as Go session
    participant L as FPGA UART link
    participant C as Graph core
    UI->>S: POST control step
    S->>L: S followed by newline
    L->>C: accept held event
    C-->>L: stable event and meaningful delta
    L-->>S: E frame with sequence and checksum
    S->>S: decode, validate, project, retain
    S-->>UI: latest state and history summaries
    C->>C: prepare next semantic event and hold
```

The FPGA can prepare the next pending event after accepting the previous one. The UI describes the latest delivered event, not an asynchronous view of all current physical registers. No additional semantic event is delivered until the next step. Continuous Run uses the same exchange, paced by a 25 ms worker ticker; the serial framing alone costs about 6.09 ms per event at the configured clock/divider. Observed runtime is therefore dominated by instrumentation and host pacing.

### Failure and retention behavior

Malformed or timed-out serial responses leave uncertainty about which response boundary the device reached. The serial engine marks itself unsynchronized and rejects further Step calls. Explicit reset or reload waits 250 ms, longer than the FPGA parser's 200 ms partial-command timeout, drains pending input, and begins a new acknowledged exchange. No automatic retry can silently advance the solver twice.

HTTP graph input uses temporary variable-length edge arrays and requires exactly two endpoints before constructing `[2]int` pairs. This avoids Go JSON array decoding silently accepting extra endpoints. Unknown fields, trailing JSON, inappropriate content types, and bodies beyond 4096 bytes are rejected. The frontend validates the same obvious input rules for immediate feedback; backend and FPGA validation remain authoritative.

History retains 256 complete snapshots with independent slices. A generation identifies a load/reset run, so sequence 1 from an older graph cannot be confused with sequence 1 from the new graph. A historical request requires both sequence and generation. A wrong generation returns 409; a missing or evicted sequence returns 404. The browser uses the query's currentData field so a previous successful response cannot appear under a newly selected event label while a request is pending.

### Build and validation evidence

The production asset contract is exactly index.html, app.js, and app.css. `scripts/build-web.py`, invoked by `go generate ./internal/microscope`, runs the TypeScript/Vite build and copies those three files into ignored embed inputs. The default build reads generated assets from disk; the `embed` build uses blank-imported Go embedding to include them in the executable. Only JS/CSS are exposed under `/static/`. See `graph_microscope/README.md` for runnable commands.

The routed graph design used 4,242 CPE_LT, 1,007 CPE_FF, and 2 RAM_HALF. Final routed timing was 26.76 MHz, passing the actual 10 MHz constraint. These units are the implementation tool's resource categories. The graph design includes runtime input and complete event transport, so its resource count should not be treated as an isolated comparison of propagation algorithms.

The final evidence set includes Go race tests with seven real UART simulations, an independent exhaustive oracle over all 64 simple four-vertex graphs at one through three colors, ten frontend tests, TypeScript and production builds, default/embedded Go builds, vet and the version-matched Glazed analyzer, and 58 original queens tests. Physical event captures compare the board against the model; the embedded HTTP smoke confirms complete user-facing runs. Desktop and 390-pixel mobile browser checks verified trail selection, historical labeling, disabled historical controls, return to live, and no horizontal overflow. The embedded page produced no console errors.

The first vulnerability scan found 12 reachable Go 1.26.1 standard-library advisories. Updating the root requirement to Go 1.26.8 resolved the reachable findings on the first repair attempt. The final scanner reported zero affected vulnerabilities, while noting one advisory in imported packages and five in required modules whose vulnerable symbols the project does not appear to call. The original and final logs are preserved; this is a point-in-time scan, not a claim about future dependencies. The consulted [official Go release history](https://go.dev/doc/devel/release) is archived in sources/go-release-history.md.

Evidence files reside in reference/validation: P3 hardware and routing logs, api-serial-smoke.json, P6 test/build/lint/vulnerability logs, and desktop/mobile screenshots. The detailed chronological diary explains failed attempts and corrections. Printer receipts reside in reference/slips. The initial design was uploaded before implementation; the completed guide is delivered as a separate reMarkable document to preserve any annotations on the original.
'''
guide.write_text(text.rstrip() + '\n')

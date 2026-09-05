from pathlib import Path
import json
root=Path(__file__).resolve().parents[1]
p=root/'design-doc/01-programmable-dataflow-workbench-intern-analysis-design-and-implementation-guide.md'
front=p.read_text().split('---',2)[1]
p.write_text('---'+front+'---\n\n'+'''# Programmable dataflow workbench: intern analysis, design, and implementation guide

## Purpose and deliverables

This project extends the working elastic dataflow laboratory into a programmable expression workbench. A user writes a small typed expression language, compiles it to a bounded acyclic graph, loads that graph into either the Go model or FPGA, supplies named inputs for a context, and observes execution through the existing Go/React IDE. Hardware breakpoints and a bounded event trace add observations at computational clock boundaries rather than relying only on snapshots taken after long tick commands.

The implementation is grounded in the preceding ticket, GATEMATE-SYMBOLIC-007, at repository revision aaad4ff. That machine already evaluates typed tokens with four contexts, synchronous operand RAM, independent pending activations, elastic arithmetic units, finite queues, epoch cancellation, and UART inspection. This guide explains that foundation before specifying the changes. It is a design contract; qualification results and any justified implementation changes must be recorded in the diary and final reference documentation.

The scope is six implementation phases: design and delivery; graph model and compiler; programmable RTL and transport; physical debug control and trace; the workbench UI; and end-to-end qualification. Every phase has a printed start and completion slip and a focused commit. The guide is uploaded before implementation so the full system can be reviewed independently of the source changes.

## 1. Existing state and ownership

A context represents one expression instance. Each of four contexts owns seven operand slots. A slot has A and B values, validity bits, an issued bit, and a pending bit. The physical values live in two synchronous 40-bit RAMs with 32 addresses, of which 28 are used. Metadata lives in registers. The scheduler finds a pending slot, reads its operands, captures the returning values, and dispatches to a shared multiplier or ALU.

The existing graph calculates `a*b + c*d + int(e<f)`. Descriptors currently live in Go's `Descriptors` array and RTL package functions. This hardcoding extends into routing, final-token construction, required-port checks, and the React graph layout. Programmability therefore requires changing each consumer, not only adding an editor or a descriptor upload command.

A token is 80 bits: context, epoch, node/port/final, producer, subtype, and a 40-bit typed value. INT uses tag zero and a signed 32-bit payload. BOOL uses tag one and payload zero or one. Flags are zero. ERROR uses tag thirteen with a fault code. MUL accepts signed 16-bit operands and returns a signed 32-bit result; ADD and SUB check signed overflow; LT returns BOOL; BOOL_TO_INT performs an explicit conversion; COPY preserves a canonical input value.

The machine performs one activation per `(context, epoch, node)`. The issued bit is not cleared when execution completes. Reusing a context requires cancellation or global reset. This work retains that execution contract. It does not add cyclic graphs, streaming iterations, or repeated node firing.

```mermaid
flowchart TD
    Editor[Typed source editor] --> Compiler[Go parser and graph compiler]
    Compiler --> Program[Validated program and source map]
    Program --> Session[Single Go session owner]
    Session --> Model[Transaction model]
    Session --> UART[UART graph and debug commands]
    UART --> FPGA[Programmable descriptor engine]
    FPGA --> Snapshot[Paused state and bounded event trace]
    Model --> Snapshot
    Snapshot --> UI[Graph, operands, timeline and stop reason]
```

The engine owns computation. The session owns operations, serial transactions, and retained frames. React owns drafts and view selection. Compiling has no device side effects. Loading and running are separate explicit actions. A historical frame must carry the graph that was active for that frame so a later program does not change the meaning of old node numbers.

## 2. Programmable graph representation

Retain the existing seven-node capacity in the first version. This fits the operand RAM, pending selection, counters, and token node field without expanding the control datapath before its new semantics are qualified. A graph contains between one and seven active descriptors. Unused slots are disabled. Four contexts execute the same loaded graph with separate operands and epochs.

Each descriptor contains an opcode, a required-port mask, up to two destination node/port pairs, a destination count, and a final flag. Required ports derive from the opcode: COPY and BOOL_TO_INT use A; arithmetic and comparison use A and B. A final node has no destinations. Every other node must lead toward the single final result.

For loaded programs, node IDs are topologically ordered: every internal destination has a greater node ID than its producer. That representation makes cycle rejection a local edge comparison in hardware. It also makes the compiler's output easy to read from left to right. The built-in laboratory graph remains a reset program used by existing directed experiments; it is not a wire-protocol compatibility layer. Its descriptors are initialized directly, and loaded-program validation has its own explicit constraints.

Validation must reject:

- An active count outside one through seven or an unsupported opcode.
- An inconsistent required mask, an inactive destination, or a destination port not required by its target.
- Backward or self edges, duplicate destination writes, more than two destinations, or multiple producers for one operand port.
- Multiple final nodes, a nonfinal node with no consumer, or a final node with destinations.
- A graph whose nodes cannot contribute to the final node.

External bindings identify the required ports that have no internal producer. The compiler assigns every such port exactly one named input or constant value. The physical token interface still checks runtime duplicate writes and types. Host compilation establishes static consistency; runtime checks remain necessary because raw injection is a supported experiment operation.

```text
validate_graph(graph):
    check node count and each opcode
    incoming = empty set of destination ports
    for each active descriptor n:
        check mask, finality, destination count
        for each destination d:
            require n < d.node < active_count
            require d.port is required by descriptor[d.node]
            require d not in incoming
            add d to incoming
    require exactly one final node
    require all nodes reach that final
```

The seven-node limit is a visible resource limit. Compiler diagnostics must report the required count after any inserted COPY nodes. Silently dropping expressions or evaluating excess nodes on the host would produce a different machine and is not permitted.

## 3. The typed source language

The initial language has declarations, named intermediate values, and one output statement. Semicolons or newlines delimit statements. Names are identifiers. Inputs declare `int16`, `int32`, or `bool`; `let` binds an expression once. Expressions include identifiers, signed integer literals, parentheses, multiplication, addition, subtraction, less-than, and `int(boolean_expression)`.

```text
input a, b, c: int16
let square = a * a
let offset = b * c
let selected = int(a < b)
output square + offset + selected
```

The parser observes ordinary precedence: multiplication before addition/subtraction, followed by comparison. The explicit conversion is required to combine a predicate with integer arithmetic. Undefined names, duplicate definitions, invalid conversions, and missing or repeated output statements produce source diagnostics with locations. Inputs are values, not graph operations; supplying one input to several ports produces several source tokens under a defined host injection policy.

Type inference distinguishes the signed 16-bit input range from general signed 32-bit arithmetic. Multiplying a general INT32 expression is rejected unless the compiler can establish the operand is in the supported range. Addition and subtraction can overflow at runtime and remain checked by the engine. A literal must fit signed 32 bits; its inferred range can establish that it is a legal multiplier operand.

The compiler builds a dependency DAG with a source map from node IDs to expressions and binding names. Reusing a named intermediate reuses its producer. A value consumed twice can use the descriptor's two destinations directly. More consumers require explicit COPY distribution nodes, each with at most two outgoing edges. These consume the same node budget as arithmetic. Topological ordering occurs after fanout lowering so every loaded edge points forward.

```text
compile(source):
    parse declarations and expressions with locations
    resolve names and infer/check types
    retain operations reachable from the output
    build explicit producer-to-port edges
    lower fanout beyond two using COPY nodes
    topologically order all operation nodes
    reject if lowered node count exceeds seven
    build descriptors, named input bindings, constants, source map
    validate the complete graph
    return immutable compiled program
```

An output that directly names an input or literal still requires a final COPY node, because terminal delivery is a node operation. A program with an unused binding may omit that operation from the emitted graph. Diagnostics and the graph source map should make such elimination understandable.

## 4. Loading a graph without mixing executions

Descriptor mutation during execution would be unsafe. An old completion could emerge with a producer ID whose destinations had changed. Epoch cancellation alone cannot fix this because epochs identify contexts, not descriptor revisions. This version therefore permits graph loading only into a freshly reset execution state, before any input, ticks, or cancellation have been accepted.

The host load operation validates the entire graph before touching the device, then performs staged writes and an activation command while holding serial ownership. The device receives descriptors into a shadow bank. A write bitmap proves that every requested active descriptor was supplied. Activation checks structural invariants before copying the staged graph into active descriptors. Failure leaves the previously active graph unchanged. Reset is the explicit way to leave an uncertain partial load.

The proposed descriptor payload uses four bytes: node index, opcode/finality/count bits, destination zero, and destination one. Each destination packs node and port; unused destination bytes are zero. A commit payload carries active count. The wire decoder must reject reserved bits and noncanonical unused fields. The detailed implemented byte map belongs in a separate API reference and its golden encoding tests.

New commands extend the existing stop-and-wait UART framing: `W` stages a descriptor, `G` activates the staged graph, and `B` controls breakpoint/trace state. `Q` continues reading ten-byte pages. The capability version changes so the host does not interpret an old bitstream as supporting new registers. No old-protocol adapter is added.

The existing rules for uncertain mutations still apply. A definite full-input rejection can be retried after credit is obtained. A missing acknowledgement to graph activation is an uncertain mutation and requires explicit reset. The host must not infer that loading failed merely because it did not receive the response.

## 5. Core RTL changes

Replace fixed operation lookup, required-port lookup, destination lookup, and final-node assumptions with active descriptor fields. Keep the physical operand address `7*context + node`. Disabled nodes are rejected as destinations and excluded from scheduler eligibility. The issue token captures finality from the descriptor when the activation is selected. Arithmetic admission uses the selected opcode.

Generalize routing so every node can have one or two destinations. The existing delivered mask already retains successful commits during fanout; remove the special case that identifies COPY by node six. A router record is released when its configured destination count is satisfied. Both the Go transaction model and semantic model must use their own graph state rather than a mutable process-global descriptor table.

Descriptor debug pages expose the actual active graph. A physical snapshot must be decoded from those pages, not populated from the compiler's last requested program. This catches mismatched loading and preserves the provenance of the UI. The source map is host metadata tied to the loaded program; raw hardware descriptors remain independently inspectable.

The RTL must preserve the existing synchronous RAM read-ownership guard. A paused debug query can change RAM addresses while an issue is waiting. Resuming must not capture the debug-selected operands into that issue. Programmable graphs increase the number of possible access patterns but do not change this invariant.

## 6. Hardware breakpoints and stopping semantics

A breakpoint configuration includes a mask of supported predicates and an optional node/context selector. Initial predicates are node issue, completion queue full, and stale-discard occurrence. The selection identifies a node ID from the compiled source map; wildcard context allows stopping on any of the four expression instances.

A breakpoint stops after the clock edge on which its condition is observed. Other legal operations on that same edge may also complete. The resulting snapshot represents the entire post-edge state, not an artificially serialized account in which only the selected event happened. The stop reason records which condition triggered and the enabled-cycle count.

```text
on each computational edge:
    perform normal simultaneous state transitions
    capture eligible debug events
    if configured breakpoint matches this edge:
        latch halted and stop reason

while halted:
    retain computational state
    permit UART queries and explicit debug control
```

The UART tick loop terminates early when the engine halts and acknowledges that bounded run request. The caller reads the stop reason and actual cycle count from the snapshot. Resume explicitly clears the halt latch. Disarming breakpoints and clearing trace are separate explicit controls or distinct bits in a validated debug operation. A single-cycle control must not accidentally ignore an already latched breakpoint.

A queue-full predicate can match a level that persists. Resuming without changing capacity can therefore halt again. The UI should display the reason and allow polling or disarming rather than concealing this valid behavior.

## 7. Trace representation and truthfulness

Retain a bounded trace of 32 event records in the first physical implementation. A record includes enabled-cycle number, event kind, and an 80-bit token when the event has token identity. Useful events include operand commit, issue, unit completion, route delivery, output acceptance, cancellation, and stale discard. Events from one enabled edge share the same cycle number.

The core can perform several transitions on one edge. If the physical recorder accepts only one record per edge, use a documented priority and increment a dropped-event count for all unrecorded candidates. Once the buffer is full, preserve the captured prefix and count later losses. Never present a partial trace as complete. A single valid bit or an unqualified event list is insufficient evidence for a reconstructed timeline.

Each trace record occupies two ten-byte debug pages: token and metadata. Capacity and overflow count have dedicated status fields. This fits the existing eight-bit page address space. Source maps allow the UI to label node events, but the token's physical producer, context, and epoch remain visible.

The Go model implements the same public debug concepts while retaining its own cycle timing. Tests compare program results and meaningful activation values across implementations, not exact physical stage positions. A breakpoint test does assert the stopping boundary within its own implementation. Trace overflow and simultaneous-event loss must have directed tests because they affect what the user can conclude from the display.

## 8. Go APIs and service integration

Extend the existing `Engine.Execute(context.Context, Operation)` boundary with graph load and debug operations. `Operation.Validate` checks payload presence, graph structure, configuration ranges, and mutually inappropriate fields. `Snapshot` gains active graph descriptors and a detached debug snapshot. Clone must deep-copy any newly introduced slices so historical frames remain immutable.

The compiler should be an ordinary Go package API, callable without an HTTP server. Proposed functions are `Compile(source) (Program, error)`, `ValidateGraph(Graph) error`, and a named-input token builder that validates supplied values and maps them to ports at a specified context/epoch. Use `github.com/pkg/errors` for contextual errors and existing package conventions for cancellation and logging.

The service adds compilation and program loading endpoints. Compile returns diagnostics and a program preview without modifying the engine. Load uses the session's expected-frame guard, resets explicitly as a user-visible operation, loads the graph, and captures the physical state. Running named inputs validates all values before injecting any token. A full-input rejection can obtain bounded credit; uncertain operations retain the existing reset-required behavior.

The source editor, input form, compile result, and loaded program identity must be distinct state. Editing a draft does not change the loaded graph. A failed compilation must not clear the working device's graph or fabricate a new source map for it. Loading a new graph starts a new history generation, making node identities unambiguous.

## 9. React workbench behavior

Retain React, TypeScript, Redux Toolkit, RTK Query, Bootstrap, and the existing embedded Go frontend path. Add a typed-program editor alongside the existing low-level scenario tools. The UI offers Compile, Load, named context inputs, Tick/Run, breakpoint configuration, Resume, and Clear Trace. Every mutation uses the current expected frame and is disabled in historical inspection.

The graph renderer must derive nodes and edges from `snapshot.graph`. A topological layout can assign columns by dependency depth and rows within a column. The built-in graph may use the same general layout; there is no need to maintain a special diagram adapter. Node labels combine operation name with loaded source-map names where available. Clicking a node selects its operand state and filters trace records.

The timeline renders actual captured events grouped by cycle, with token type/value, context, epoch, and producer. A prominent loss indicator explains when the retained list is incomplete. A details table provides exact fields for review. Queue occupancy at a snapshot is a real observation; intermediate occupancy between snapshots must not be invented from an incomplete trace.

Save/import/export can preserve typed source and named inputs using an explicit new document shape or a separate program endpoint. Existing scenario JSON remains useful for raw scheduling experiments. The implementation must document the final persistence contract rather than quietly accepting ambiguous mixed schemas.

## 10. Validation strategy

Compiler tests cover precedence, signed constants, undefined/duplicate names, BOOL conversion, multiplier range restrictions, repeated subexpressions, fanout lowering, direct-input output, and capacity diagnostics. Independent expected values for multiple input assignments ensure the compiler does not merely test itself by reusing its own intermediate result calculation.

Graph tests cover malformed destinations, repeated writers, cycles, dead nodes, finality, and active count. Model tests execute several distinct compiled graphs across multiple contexts, stalls, cancellation, and output hold. Snapshot mutation tests include the new graph and debug fields.

RTL tests load a graph that differs from the built-in one and verify every relevant activation and final result. They attempt partial/invalid activation, graph writes after execution, reset during staging, and graph readback. Existing directed queue/latency tests continue to check the original ready/valid and cancellation contract after descriptor generalization.

Debug tests halt on a specific issue, retain state across additional requested ticks, resume, trigger a full completion queue, and cancel in-flight work to observe stale handling. Trace tests exercise capacity exhaustion and simultaneous candidates and check the loss counter. UART tests cover new lengths, checksums, bounds, acknowledgements, and early tick completion on halt.

Physical qualification rebuilds and routes the expanded design, checks timing against 10 MHz, programs the board, and runs compiled programs over real UART. A browser test compiles, loads, supplies inputs, stops at a hardware breakpoint, reads actual descriptor pages and trace, resumes to an expected result, and captures screenshots. Timing failure is a real implementation failure; model success cannot replace physical qualification.

## 11. File map for the intern

| Existing path | Responsibility and planned work |
|---|---|
| `pkg/dataflow/types.go` | Tagged values, descriptors, per-engine graph identity |
| `pkg/dataflow/semantic.go` | Timing-independent programmable execution |
| `pkg/dataflow/transaction.go` | Bounded programmable model and debug transitions |
| `pkg/dataflow/engine.go` | Operations, snapshots, validation, clone ownership |
| `pkg/dataflow/protocol.go`, `serial.go` | New wire commands, graph/debug page decoding |
| `elastic_dataflow/rtl/dataflow_core.sv` | Descriptor banks, generic routing, breakpoints and trace |
| `elastic_dataflow/rtl/dataflow_link.sv` | Load/debug request sequencing and early tick stop |
| `elastic_dataflow/rtl/dataflow_pkg.sv` | Shared arithmetic and reset descriptor helpers |
| `internal/dataflowide/session.go`, `http.go` | Atomic session operations and compiler routes |
| `web/src/dataflow/Graph.tsx`, `App.tsx` | Actual graph layout and workbench controls |
| `web/src/dataflow/types.ts`, `store.ts` | Typed API payloads and query integration |

New compiler and graph-validation files belong under `pkg/dataflow`, with focused tests beside them. Experimental authoring, verification, printing, and upload helpers belong in this ticket's `scripts/` folder. Any downloaded reference belongs in `sources/`; use Defuddle for web-page Markdown. This implementation can rely primarily on local code and the archived lab material.

## 12. Review gates and implementation order

P1 completes when this guide is linked to source files, tasks are registered, the design receipt is recorded, and the guide is uploaded. P2 establishes executable graph/compiler semantics before hardware changes. P3 qualifies descriptor loading and execution in RTL and host transport. P4 adds stopping and trace with explicit loss semantics. P5 connects those already-defined contracts to the browser. P6 validates the real board, captures screenshots, updates the final API reference, and closes the ticket only when all required work has passed.

The principal review risks are descriptor/execution mixing, compiler fanout mistakes, stale source maps, premature UART acknowledgement of a halted run, lost trace events presented as complete, and new debug logic on the timing-critical path. Each has an explicit state boundary or test above. Record exact failures in the diary as they occur. If two consecutive repair attempts fail, stop under the repository's debugging instruction and request a broader review rather than continuing unbounded local repairs.
''')
prompt='yes do it. Create a new docmgr ticket, then Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable.\n\nThen implement, commit at appropriate intervals and keep a detailed diary as you work (using the diary format from the skill). slPrint out a brutalist work slip with the plan / different phases for the ticket. then before stsarting a phase, plrint a split about the phase, and print one when the phase is done.'
(root/'sources/user-request.json').write_text(json.dumps({'prompt':prompt},indent=2)+'\n')

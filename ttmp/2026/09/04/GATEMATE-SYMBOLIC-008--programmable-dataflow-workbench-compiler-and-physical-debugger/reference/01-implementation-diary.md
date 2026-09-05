---
Title: Implementation diary
Ticket: GATEMATE-SYMBOLIC-008
Status: active
Topics:
    - fpga
    - gatemate
    - architecture
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: repo://elastic_dataflow/rtl/dataflow_core.sv
      Note: Programmable descriptors, breakpoint state and trace storage
    - Path: repo://pkg/dataflow/compiler.go
      Note: Compiler implementation and source-map ownership
    - Path: repo://pkg/dataflow/workbench_physical_test.go
      Note: Compiled physical qualification contract
    - Path: repo://ttmp/2026/09/04/GATEMATE-SYMBOLIC-008--programmable-dataflow-workbench-compiler-and-physical-debugger/scripts/16-parallel-graph-validation.py
      Note: Recorded targeted timing repair
    - Path: repo://web/src/dataflow/Workbench.tsx
      Note: Workbench implementation and browser evidence
ExternalSources: []
Summary: ""
LastUpdated: 2026-09-04T22:09:27.769234375-04:00
WhatFor: ""
WhenToUse: ""
---


# Implementation diary

## Goal

Record the programmable workbench design, implementation, validation, physical evidence, and delivery.

## Step 1: Intern guide, execution contracts, and design delivery

Created the new workbench ticket and an intern guide that explains the existing engine before specifying programmable descriptors, a typed compiler, hardware stopping, bounded trace, and React integration. The design preserves seven physical nodes per context and makes capacity and trace loss explicit.

Printed the complete six-phase plan and phase-one start slip. Uploaded the completed guide to reMarkable before implementation; the receipt identifies the ticket-specific remote directory.

### Prompt Context

**User prompt (verbatim):** Exact request as JSON:

```json
"yes do it. Create a new docmgr ticket, then Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable.\n\nThen implement, commit at appropriate intervals and keep a detailed diary as you work (using the diary format from the skill). slPrint out a brutalist work slip with the plan / different phases for the ticket. then before stsarting a phase, plrint a split about the phase, and print one when the phase is done."
```

**Assistant interpretation:** Build the programmable graph/compiler/debugger workbench with an intern guide, physical evidence, and phased delivery.

**Inferred user intent:** Make the existing FPGA execution engine programmable and inspectable while preserving a reviewable account of its correctness.

### What I did

- Created GATEMATE-SYMBOLIC-008, six tracked phase tasks, the design document, and this diary.
- Read existing model, RTL, UART, session, and graph-renderer code; related five implementation paths to the design.
- Retained authoring, printing, upload, and diary helpers in scripts/ and the exact user request in sources/.
- Uploaded GATEMATE 008 Programmable Dataflow Workbench Design.pdf to /ai/2026/09/04/GATEMATE-SYMBOLIC-008.

### Why

- Programmability affects finality, routing, readiness, loading, and historical graph identity; changing only the editor would be incomplete.
- Staged loading after reset prevents old tokens from acquiring a new descriptor interpretation.

### What worked

- Both initial printer receipts report printed: true.
- The upload reports OK: uploaded; no redundant cloud listing was needed.

### What didn't work

- No failures in this phase.

### What I learned

- The current router special-cases COPY node six and final node five; both assumptions must be removed from runtime execution.
- Seven physical nodes are sufficient for the proposed example, but compiler-inserted fanout can consume the remaining capacity.

### What was tricky to build

- Trace candidates can coincide on one clock. The design requires explicit dropped-event accounting rather than pretending a one-record-per-cycle trace is complete.
- Source maps must remain associated with loaded graphs and historical generations.

### What warrants a second pair of eyes

- Review graph activation preconditions, byte encoding, breakpoint post-edge semantics, and trace loss before implementing hardware.

### What should be done in the future

- Implement P2 graph validation and compiler tests, followed by programmable RTL and debug contracts.

### Code review instructions

- Read design-doc/01-programmable-dataflow-workbench-intern-analysis-design-and-implementation-guide.md and reference/validation/p1-upload.log.
- Inspect plan-print.log and p1-start-print.log for physical print receipts.

### Technical details

- Source baseline aaad4ff; active count 1..7; four contexts; two outgoing destinations; 32 proposed trace records.
- Phase task IDs: 13x4, ycgx, g08b, 2bb4, cho9, 54tm.

## Step 2: Typed compiler and per-engine executable graphs

Implemented a bounded graph image and moved semantic and transaction execution to per-engine descriptors. The compiler parses declarations and expressions, checks integer and Boolean types, lowers internal fanout through COPY nodes, orders nodes topologically, and emits named-input bindings and source metadata.

Directed tests execute compiled programs through both models across all four contexts with one-entry queues. They cover a different arithmetic graph, repeated intermediate values, fanout beyond two, signed limits, constants, Boolean conversion, and direct input output.

### Prompt Context

**User prompt (verbatim):** See Step 1.

**Assistant interpretation:** Build the programmable graph/compiler/debugger workbench with an intern guide, physical evidence, and phased delivery.

**Inferred user intent:** Make the existing FPGA execution engine programmable and inspectable while preserving a reviewable account of its correctness.

**Commit (code):** 429e2a2 — feat(dataflow): compile typed expressions into validated per-engine graphs

### What I did

- Added graph.go, compiler.go and compiler_test.go; generalized descriptor access and finality in existing models.
- Added validated load operations and graph snapshots with detached fixed-size state.
- Ran go test ./pkg/dataflow ./internal/dataflowide -count=1 and race-enabled package tests.

### Why

- A mutable global descriptor table would allow one engine or historical snapshot to change another engine's graph.
- Checking complete named inputs before token generation prevents partial injection caused by a later validation error.

### What worked

- All focused model, compiler, session and race checks passed.
- The four-consumer intermediate program lowers within the seven-node budget and produces the expected result.
- Malformed graph edges, duplicate writers, invalid types, excessive nesting and graph capacity are rejected.

### What didn't work

- No failed tests or repair attempts in P2.

### What I learned

- External input fanout needs multiple source bindings; only internal producer fanout consumes COPY nodes.
- Finality must be generated from the active descriptor rather than assuming node five.

### What was tricky to build

- COPY insertion changes dependency ordering; final node IDs are assigned after fanout lowering.
- Reset retains the original laboratory graph as a built-in program while loaded programs use stricter forward-edge validation.

### What warrants a second pair of eyes

- Review type/range diagnostics and fanout lowering, particularly shared intermediates used three or four times.
- Graph loading rejects execution state that has already accepted input, ticks or cancellation.

### What should be done in the future

- P3 implements the same descriptor contract in RTL and versioned UART readback.

### Code review instructions

- go test -race ./pkg/dataflow ./internal/dataflowide
- Read compiler_test.go for independent expected arithmetic results and malformed graph cases.

### Technical details

- P1 design commit af8e50f; P2 implementation commit 429e2a2.
- reference/validation/p2-go-tests.log retains race-test results.

## Step 3: Programmable RTL descriptors and UART graph transactions

Replaced runtime fixed-node assumptions with active descriptor fields throughout operand validation, issue, routing and finality. Added staged descriptor writes, a fresh-reset load guard, full-image structural validation, atomic activation, version-two capabilities and active graph readback.

The Go serial engine now sends validated descriptor images and commits them under one ownership lock. Snapshot decoding reads physical graph pages. Directed RTL and UART tests execute a two-node multiply fanout graph yielding 84 and a four-node graph yielding 43.

### Prompt Context

**User prompt (verbatim):** See Step 1.

**Assistant interpretation:** Build the programmable graph/compiler/debugger workbench with an intern guide, physical evidence, and phased delivery.

**Inferred user intent:** Make the existing FPGA execution engine programmable and inspectable while preserving a reviewable account of its correctness.

### What I did

- Added shadow and active descriptor banks, write bitmap, validation and graph pages 148..155.
- Added W/G UART commands and Go serial load encoding with golden-frame tests.
- Updated all directed simulation harnesses for explicit graph control inputs and capability version two.
- Ran twelve existing depth/latency configurations, the UART suite, programmable graph tests and focused Go tests.

### Why

- Descriptors must remain stable while tokens are in flight; only a pristine execution state may activate a new graph.
- Physical snapshots must read actual descriptors rather than assuming the host's last compile request succeeded.

### What worked

- All twelve original RTL configurations passed after generic routing changes.
- Programmable tests passed changed operations/finality, fanout, graph readback, incomplete activation and invalid edge rejection.
- UART tests passed staged loading, graph readback, result 84 at node one and late-write rejection.

### What didn't work

- Initial elaboration reported: rtl/dataflow_core.sv:69: error: Unable to bind wire/reg/memory `cancel_fire' in `tb_dataflow.dut'; declaration was later at line 84.
- Moved cancel_fire before its new use. The next full directed/UART run passed; no second repair attempt was needed.

### What I learned

- The RTL also encoded unary-port validation using node numbers; this was changed to required_ports from the loaded opcode.
- Forward-only edges, nonempty nonfinal fanout and exactly one final sink establish reachability in this bounded graph.

### What was tricky to build

- G selects graph_size before evaluating validation on the following clock, avoiding acknowledgement against a stale count.
- Unused physical descriptors read as zero and are decoded separately from active descriptors.

### What warrants a second pair of eyes

- Review graph bitmap/activation ordering and the explicit reset-program exception in readback validation.
- Hardware placement and timing qualification remain P6 work; simulation success is not physical qualification.

### What should be done in the future

- P4 adds post-edge hardware breakpoints, early tick completion and a bounded trace with loss accounting.

### Code review instructions

- Run scripts/08-rtl-checks.sh under the installed OSS CAD environment.
- go test ./pkg/dataflow ./internal/dataflowide -count=1

### Technical details

- Wire descriptor: index, opcode<<4 | final<<3 | count, destination0, destination1; destination=node<<1 | port.
- Capability version is two. All new command payloads retain XOR framing and stop-and-wait ownership.

## Step 4: Post-edge hardware stopping and bounded trace

Added issue, completion-full, and stale-discard breakpoints with an explicit halt latch and stop cycle. UART tick commands return early on halt, and additional ticks preserve computational state until explicit resume. The model exposes the same debug concepts with its own timing.

A 32-record physical trace uses synchronous token and metadata RAMs. Its documented priority chooses one candidate per edge and counts every unrecorded simultaneous candidate as well as buffer overflow. The UI can therefore state when a trace is incomplete instead of reconstructing missing events.

### Prompt Context

**User prompt (verbatim):** See Step 1.

**Assistant interpretation:** Build the programmable graph/compiler/debugger workbench with an intern guide, physical evidence, and phased delivery.

**Inferred user intent:** Make the existing FPGA execution engine programmable and inspectable while preserving a reviewable account of its correctness.

**Commit (code):** 455637e — feat(dataflow): stop on hardware breakpoints and retain bounded execution trace

### What I did

- Added B command, debug status pages 156/157, and token/metadata page pairs 160..223.
- Added Go debug controls, detached snapshots, model breakpoints and bounded event retention.
- Expanded UART simulation to check issue halt, trace value 42, no additional halted cycles, resume result 84, overflow after repeated cancellations, and stale halt.
- Added a dedicated RTL test for completion-full stopping and simultaneous-event loss.
- Started synthesis and routing in tmux for early physical feasibility while frontend work continues.

### Why

- A bounded tick request must report the actual stopped state when a breakpoint ends it early.
- Trace observation must disclose losses; a complete-looking partial timeline would mislead the user.

### What worked

- Original RTL regression, programmable graph suite, debug suite and expanded UART tests passed.
- Go model/debug tests and race checks passed, including snapshot ownership and trace clear.
- Physical trace read addresses are normalized from page 160 before indexing the RAM.

### What didn't work

- Starting the build tmux session inside the sandbox returned error connecting to /tmp/tmux-1000/default (Operation not permitted). Reissued that authorized operation with escalation.
- No software test failures in P4.

### What I learned

- A full-queue breakpoint is a sampled pre-edge condition; the published state includes all transitions on the stopping edge.
- Cancellation and output polling can produce trace records without advancing enabled computational cycles.

### What was tricky to build

- The link sees the halt latch on the following system clock; core clock-enable gating prevents an extra computational edge during that acknowledgement delay.
- Several stale records may be discarded together. Their count contributes to dropped-event accounting even though one token can be retained.

### What warrants a second pair of eyes

- Review trace priority and the distinction between model capture bandwidth and physical single-record bandwidth.
- Placement and timing remain to be verified on the expanded design.

### What should be done in the future

- P5 adds source compilation/loading, named inputs, graph layout and truthful debug rendering in React.

### Code review instructions

- Run test-debug.sh and test-link.sh under OSS CAD Suite; read p4-debug.log and p4-uart-debug.log.
- go test -race ./pkg/dataflow ./internal/dataflowide

### Technical details

- Trace event kinds: operand 1, issue 2, completion 3, route 4, output 5, cancel 6, stale 7.
- B flags: mask bits 0..2, clear bit 6, resume bit 7; node/context use 255 as wildcard.
- Debug commit 455637e; graph transport commit 9398755.

## Step 5: Go React program workbench and model browser qualification

Built the typed program editor, graph compilation preview, explicit load/reset path, named inputs, independent .df persistence, dynamic descriptor graph, breakpoint controls, and captured-event table. Loaded source maps are associated with detached historical frames and are attached only after graph readback matches.

Browser qualification compiled the default six-node program, loaded it, supplied 3/4/5, stopped at square issue with value 9, resumed to final result 30, saved/reopened source, rejected an invalid draft, and disabled hardware controls in history. Three screenshots preserve the stopped, completed, and historical model views.

### Prompt Context

**User prompt (verbatim):** See Step 1.

**Assistant interpretation:** Build the programmable graph/compiler/debugger workbench with an intern guide, physical evidence, and phased delivery.

**Inferred user intent:** Make the existing FPGA execution engine programmable and inspectable while preserving a reviewable account of its correctness.

**Commit (code):** a88996c — feat(dataflow): add typed program workbench and physical event debugger UI

### What I did

- Added program.go routes and session operations, compiler source-map cloning, and shared confined file primitives with separate .df/.json namespaces.
- Replaced fixed SVG edges/op names with active graph descriptors and generalized operand inspector labels.
- Added ProgramEditor and Debugger components using existing Redux/RTK Query and Bootstrap conventions.
- Ran Go race tests, TypeScript, 15 frontend tests, embedded frontend generation, and the retained Playwright script.

### Why

- Draft, compiled preview and loaded source must remain distinct to prevent a failed compile from changing physical node interpretation.
- A visible trace-loss indicator is required because the hardware recorder can omit simultaneous events.

### What worked

- Model browser run stopped at enabled cycle 5 and recorded issue value 9; resumed program returned 30.
- Invalid named inputs were rejected before mutation, graph/source snapshots were detached, and reset generations invalidated old history.
- Frontend tests cover stale compilation invalidation, historical breakpoint controls and visible trace loss.

### What didn't work

- Initial TypeScript check identified test fixtures missing new graph/debug fields and Frame.program. Updated the fixtures; the next check and all frontend tests passed.
- make dataflow-frontend initially failed because /home/manuel/.cache/go-build was read-only in the sandbox. Re-ran with GOCACHE=/tmp/gatemate008-go-cache; generation succeeded.

### What I learned

- Input names are independent from physical destination IDs; the six source tokens for the sample come from only three named inputs.
- The model can issue another operation within the same stopping edge, consistent with post-edge snapshots.

### What was tricky to build

- Program load resets the session generation and checks the actual descriptor snapshot before attaching source metadata.
- The reset graph has non-topological IDs, so the renderer uses bounded dependency-depth relaxation rather than assuming increasing IDs.

### What warrants a second pair of eyes

- Review partial named-input injection handling and explicit reset-required state after uncertain serial operations.
- Physical browser evidence remains P6 work; these three screenshots are explicitly model observations.

### What should be done in the future

- Complete timing and physical qualification, capture hardware workbench screenshots, document final wire/API contracts.

### Code review instructions

- go test -race ./pkg/dataflow ./internal/dataflowide; pnpm --dir web check; pnpm --dir web test.
- Run scripts/14-browser-model.js against the model server on port 8088.

### Technical details

- Screenshots: 01-model-breakpoint.png, 02-model-result.png, 03-model-history.png.
- Browser result record: reference/validation/p5-browser-model.json.
- Program source files use .df under the existing confined project root; scenario files remain .json.

## Step 6: Physical timing evidence and first targeted repair

The first expanded design completed placement and routing but failed the required 10 MHz constraint. The timing report identifies a path from staged descriptor bits through duplicate-writer validation to the UART acknowledgement register. This is a load-time control path, not the arithmetic datapath.

Replaced the incremental dynamic writer bitmap with fixed descriptor-pair comparisons. The bitmap caused each later destination check to depend on earlier variable-index updates, producing a serial shift/multiplexer chain. Pairwise comparisons express independent equality checks followed by a reduction. Physical qualification remains open until the rebuilt design meets timing.

### Prompt Context

**User prompt (verbatim):** See Step 1.

**Assistant interpretation:** Build the programmable graph/compiler/debugger workbench with an intern guide, physical evidence, and phased delivery.

**Inferred user intent:** Make the existing FPGA execution engine programmable and inspectable while preserving a reviewable account of its correctness.

### What I did

- Archived the failing timing report as p6-timing-first.log.
- Read the source and destination registers and net definitions of the 107.37 ns critical path.
- Changed duplicate-writer validation to compare fixed descriptor pairs in parallel.
- Started the directed RTL/UART regression before rebuilding.

### Why

- The required clock remains 10 MHz; reducing it would alter the delivered hardware contract.
- The measured path provides a specific implementation cause for a targeted repair rather than changing placement seeds blindly.

### What worked

- The first build fit the device with 18413 CPE_LT cells (44%) and 4146 CPE_FF cells (10%).
- The report clearly attributes the longest path to staged descriptor validation.

### What didn't work

- ERROR: Max frequency for clock 'link.clk': 9.31 MHz (FAIL at 10.00 MHz)
- make: *** [Makefile:9: bit] Error 1
- Critical path: 107.37 ns total, 26.06 ns logic and 81.31 ns routing.

### What I learned

- A compact behavioral bitmap algorithm can serialize variable-index writes after synthesis even for only fourteen possible destination ports.

### What was tricky to build

- Graph validation is combinational and feeds the UART acceptance decision. Its source-level loops must describe bounded parallel circuitry to meet the same clock as the compute engine.

### What warrants a second pair of eyes

- Check duplicate detection for both destinations of one producer and for destinations shared across different producers.

### What should be done in the future

- Complete first repair qualification; if another repair is needed, obey the two-failed-repair stopping rule.

### Code review instructions

- scripts/08-rtl-checks.sh
- Rebuild with scripts/11-build-board.sh and inspect the actual timing report.

### Technical details

- This records the initial timing failure and the first repair, whose physical result is not yet known.

### P6 qualification continuation

- The first timing repair is committed as `0c42939`; its placement report shows a pre-routing estimate of 20.06 MHz, which is not yet the final routed result.
- Full repository checks passed: race, ordinary/embedded build, vet, Glazed lint, vulnerability scan, TypeScript, and fifteen frontend tests.
- The twelve-configuration randomized RTL differential suite passed. Additional graph tests reject a destination shared by separate producers and a write to an unused unary B port.
- Added a dedicated physical test for 96 compiled expressions across three different programs, descriptor readback, issue breakpoints, held halt, resume, overflow and trace clear. It remains unexecuted until the new bitstream is timing-qualified and programmed.
- Code review identified a potential model-only load-window reopening after counter wrap. A dedicated started flag now matches the physical pristine-state intent; a targeted wrap test passes.
- A mobile browser check at 390 pixels passed without horizontal overflow; screenshot `04-model-mobile.png` is retained.

## Step 7: Registered graph validation after the first repair missed timing

The first targeted repair completed routing at 9.88 MHz, improving the original 9.31 MHz result but still failing the 10 MHz requirement. The new critical path continued through the accumulated graph_acceptable decision into staging control. This is the first failed repair attempt.

The second and final permitted repair expresses independent node, edge and duplicate predicates as wires with explicit reductions, then registers the validation result. The UART G command now waits three clocks after selecting the count before consulting it. Directed tests also allow that defined settling boundary. The second physical build is running; if it fails timing, work must stop under the user's debugging rule.

### Prompt Context

**User prompt (verbatim):** See Step 1.

**Assistant interpretation:** Build the programmable graph/compiler/debugger workbench with an intern guide, physical evidence, and phased delivery.

**Inferred user intent:** Make the existing FPGA execution engine programmable and inspectable while preserving a reviewable account of its correctness.

### What I did

- Archived p6-timing-repair1.log with the 9.88 MHz failure and critical path.
- Replaced procedural graph-acceptable priority updates with independent predicate arrays and reductions.
- Registered validation result and count; invalidate cached validity during descriptor writes/commit.
- Added an explicit G validation wait in the UART link.
- Ran all twelve directed configurations, programmable graph, UART and debug simulations successfully before starting the second build.

### Why

- The first parallel equality change left the combined procedural validity decision on a long timing path.
- Loading occurs while computation is pristine and paused, so a few explicit control clocks are available without changing arithmetic semantics.

### What worked

- The second repair passed the complete directed/UART suite, including partial/invalid graph activation and runtime breakpoint behavior.
- The completion-full and simultaneous-event loss test also passed.

### What didn't work

- ERROR: Max frequency for clock 'link.clk': 9.88 MHz (FAIL at 10.00 MHz)
- The first repair critical path remained 101.20 ns from staged state through graph_acceptable to staged_valid control.

### What I learned

- Independent source-level comparisons are insufficient if the final validity accumulator is still represented as a procedural priority chain.

### What was tricky to build

- The cached result must not acknowledge a different staged image or count. Writes/commit invalidate it, count identity is checked, and G waits for the newly selected count to propagate.

### What warrants a second pair of eyes

- Review registered validation ownership and the G wait boundary against direct-core graph activation tests.

### What should be done in the future

- Inspect the second build's final routed timing. Stop if this second repair fails; do not attempt another local timing repair.

### Code review instructions

- p6-repair2-regression.log and p6-repair2-debug.log contain passing simulations.
- p4-board-build.log contains the current build; final timing is not yet known.

### Technical details

- Initial failure 9.31 MHz; first repair 9.88 MHz; second repair pending.
- The 10 MHz constraint is unchanged. No timing exception or reduced clock was introduced.

### Second-repair qualification continuation

- Registered validation and the explicit G wait are committed as `b03729e`; the final routed result remains pending.
- Final model ownership review removed the obsolete exported mutable reset descriptor table and unused fixed-finality completion helper. `ResetGraph()` now constructs a fresh value directly. This cleanup is committed as `2c984d3`, with race-test results in `p6-final-model-review.log`.
- The physical qualification script now refuses programming unless the packed bitstream is newer than the routing report, in addition to rejecting timing errors. This prevents an incomplete build from accidentally selecting the previous laboratory bitstream.

## Step 8: Qualify the routed FPGA and capture the physical workbench

The second timing repair passed at 12.01 MHz against the unchanged 10 MHz constraint. I programmed that newly packed image and exercised both the existing engine contract and the new compiler/debugger path on the board. This completed the physical validation that simulation and model screenshots could not provide.

The browser then compiled and loaded the default program, checked descriptor readback, stopped on its first multiplication, proved that additional ticks leave a halted engine unchanged, and resumed to result 30. Three physical screenshots and the complete observed frames are retained alongside the model evidence.

### Prompt Context

**User prompt (verbatim):** See Step 1.

**Assistant interpretation:** Build the programmable graph/compiler/debugger workbench with an intern guide, physical evidence, and phased delivery.

**Inferred user intent:** Make the existing FPGA execution engine programmable and inspectable while preserving a reviewable account of its correctness.

**Commit (code):** b03729e — registered graph validation; 2c984d3 — independent reset graph ownership; d1be832 — preceding diary checkpoint

### What I did

- Archived the final routing and synthesis logs; checked the packed bitstream was newer than the routing report.
- Stopped the prior physical server with lsof-who -p 8087 -k and ran scripts/20-physical-qualification.sh in tmux.
- Started dataflow008-fpga on port 8087 against /dev/ttyACM0; ran scripts/21-browser-fpga.js and visually inspected the stopped physical frame.
- Added physical timing, resources, exact test counts, execution interpretation, screenshots, and reproduction commands to reference/02.

### Why

- Graph activation and trace RAM are hardware paths; final routed timing and actual UART results are necessary evidence.
- Historical screenshots must contain associated observed state and must not imply FPGA rollback or a complete trace.

### What worked

- Final routing: Info: Max frequency for clock 'link.clk': 12.01 MHz (PASS at 10.00 MHz).
- JTAG programming completed. TestPhysicalQualification passed in 22.09s; TestPhysicalWorkbench passed in 5.55s; package total 27.637s.
- 128 randomized baseline expressions plus 96 compiled expressions passed; held-output protection, epoch rollover, descriptor readback, break/hold/resume, overflow and clear checks passed.
- Physical browser stopped at cycle 6 on issue value 9, remained at cycle 6 on another 100-tick request, resumed and returned 30 at observation cycle 106.
- The final physical frame had 21 trace records and eight dropped events. The UI disclosed that loss. Console: Total messages: 0 (Errors: 0, Warnings: 0).

### What didn't work

- No additional qualification failure. Earlier final timing failures remain documented: 9.31 MHz initially and 9.88 MHz after the first repair. The second repair succeeded; no third repair was attempted.

### What I learned

- The physical issue boundary is cycle 6 for this input order, compared with cycle 5 in the model; the implementation does not claim cycle-equivalent model scheduling.
- One-record-per-clock physical tracing loses simultaneous events even when the 32-entry memory is not full.

### What was tricky to build

- Keeping programming and UART ownership sequential avoided an old server consuming new protocol responses. The source bitstream timestamp check also prevented accidental testing of the previous lab image.
- Registering graph validity required carrying size identity and invalidating on writes so the extra clock boundary did not authorize a different staged image.

### What warrants a second pair of eyes

- Review graph validity cache ownership and UART G wait in the b03729e diff alongside programmable graph tests.
- Review the physical trace priority/loss contract against p6-browser-fpga.json; the recorder intentionally does not capture all events.

### What should be done in the future

- N/A for qualification. Board SRAM contents are volatile and must be programmed again after power loss.

### Code review instructions

- Start with reference/02 and pkg/dataflow/workbench_physical_test.go.
- Replay scripts/20-physical-qualification.sh with exclusive UART ownership, then scripts/21-browser-fpga.js with the physical server running.
- Full software checks and RTL configurations are retained in p6-check-summary.log, p6-differential.log, p6-repair2-regression.log and p6-repair2-debug.log.

### Technical details

- Resources: 18449/40960 CPE_LT, 4153/40960 CPE_FF, 5/64 RAM_HALF.
- Physical example: a=3, b=4, c=5; square=9; offset=20; selected=1; output=30.
- Current physical UI: http://127.0.0.1:8087/. Final guide bundle and phase completion receipt are recorded in the delivery continuation below.

### Delivery continuation

- Physical qualification, screenshots, timing evidence, and this diary step were committed as `5b4117f` — `docs(dataflow): qualify programmable FPGA workbench and capture physical execution`.
- `scripts/23-upload-handoff.sh` completed successfully: `OK: uploaded GATEMATE 008 Implemented Programmable Workbench.pdf -> /ai/2026/09/04/GATEMATE-SYMBOLIC-008`. The initial design remains a separate previously delivered document.
- Printed the P6 DONE slip with the physical qualification commit QR. The ticket now retains thirteen successful print receipts: one overall plan, six phase starts, and six phase completions.
- Checked task `54tm` and closed GATEMATE-SYMBOLIC-008 with all six tasks complete. The final delivery audit checks both upload receipts, thirteen print receipts, seven PNG files, document frontmatter, code fences, and embedded screenshot paths.
- The physical server remains in tmux session `dataflow008-fpga` on port 8087, and the model server remains on 8088. Finished build, qualification and upload sessions have exited. Unrelated tmux sessions were preserved.

## Step 9: Publish the technical workbench report to the Obsidian vault

I wrote a new textbook-style technical article explaining the implemented programmable workbench from typed expressions through physical execution and debugger observations. The 6,439-word report contains six copied screenshots, three Mermaid diagrams, pseudocode, API tables, revision-pinned source references, and a decoded table drawn from archived physical frames.

The report is append-only in Projects/2026/09/05 of the go-go-parc vault. Publishing succeeded, but the vault commit also included four concurrently staged PBUI Handheld articles; I disclosed that mistake rather than rewriting published history.

### Prompt Context

**User prompt (verbatim):**

```text
write a detailed project report for the obsidian vault as a deep dive technical analysis blog post using a textbook writing style (no analogies, see skill).      
 Commit and push the bsidian vault when done (go-go-parc vault).
```

**Assistant interpretation:** Explain the completed programmable workbench in a new illustrated technical article and publish it to the vault repository.

**Inferred user intent:** Preserve a durable explanation of the machine's principles and behavior for future study.

### What I did

- Read obsidian-vault-writing and textbook-authoring skills; inspected the previous fixed-engine article and current implementation files.
- Created various/programmable-dataflow-vault-report.md and scripts/25-publish-vault-report.py.
- Validated source references, Markdown fences, YAML metadata, PNG identities, stop cycle, result, trace count, loss count, and final completion event.
- Copied report and six figures to the vault and pushed commit 79130f230cbf25717e74a4798fb13884b4e02db9.

### Why

- The requested article explains how the completed machine works rather than narrating its implementation history.
- Separate physical and model figures preserve provenance; the article distinguishes cycle-32 completion from cycle-106 output acceptance.

### What worked

- Publisher validation passed after correcting its repository-root calculation.
- The vault note and all figure copies were byte-verified; git push advanced main from 733bb09 to 79130f2.
- The unrelated untracked AgentForum research note remained untracked and unchanged.

### What didn't work

- Initial publisher check failed with AssertionError: pkg/dataflow/compiler.go because ticket.parents[3] selected ttmp. Corrected to ticket.parents[4]; validation passed.
- A concurrent process staged four PBUI Handheld articles after the initial clean-index status inspection. I ran git commit after staging the intended report paths without a fresh reviewed index boundary; commit 79130f2 included those four articles. Push completed before I processed the command output. I disclosed the scope error to the user and did not rewrite pushed history.

### What I learned

- A shared vault index can change between initial inspection and commit. Explicit git add paths do not isolate an ordinary git commit from concurrently staged work.

### What was tricky to build

- The physical trace contains raw tagged Boolean values and intentional losses; the report decodes true and does not synthesize missing transitions.
- Qualification chronology matters: randomized RTL evidence preceded the final registered-validator change, while directed RTL and physical checks followed it.

### What warrants a second pair of eyes

- Review the article against reference/validation/p6-browser-fpga.json and source revision 85969b4.
- The vault commit contains the intended report plus concurrent PBUI articles; no contents were altered to correct commit scope.

### What should be done in the future

- Use a path-limited commit or an isolated index for future commits to a concurrently edited vault.

### Code review instructions

- Run scripts/25-publish-vault-report.py for read-only structural and evidence checks.
- Inspect reference/validation/p7-vault-report-publish.json for destination, word count, source revision, and figure hashes.

### Technical details

- Vault article: Projects/2026/09/05/ARTICLE - GateMate Symbolic - Programmable Dataflow Workbench.md.
- Vault commit: 79130f230cbf25717e74a4798fb13884b4e02db9; remote ssh://github.com/go-go-golems/go-go-parc.

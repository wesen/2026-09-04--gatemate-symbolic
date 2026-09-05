---
Title: Implementation diary
Ticket: GATEMATE-SYMBOLIC-007
Status: active
Topics:
    - fpga
    - gatemate
    - symbolic-computers
    - architecture
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: repo://elastic_dataflow/rtl/dataflow_core.sv
      Note: Operand matching, scheduling, epochs, routing, and paused debug interface
    - Path: repo://elastic_dataflow/rtl/df_unit.sv
      Note: Elastic arithmetic pipeline
    - Path: repo://elastic_dataflow/sim/tb_dataflow.sv
      Note: Directed RTL invariants and parameter matrix
ExternalSources: []
Summary: ""
LastUpdated: 2026-09-04T19:54:24.640534987-04:00
WhatFor: ""
WhenToUse: ""
---


# Implementation diary

## Goal

Explain every design, implementation, validation, and physical delivery step in the dataflow laboratory.

## Step 1: Design contract and work plan

Created the new dataflow ticket and a 3746-word intern guide before implementation. The design covers the book expression, four contexts, typed values, indexed activation reservations, elastic units, cancellation and wrap policy, and a physical tick-controlled UART instrument.

### Prompt Context

**User prompt (verbatim, JSON encoded):**

```json
"Create a new docmgr ticket for the new lab, Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable.\n\nThen, implement it, task by task, commit at appropriate intervals and keep a detailed diary as you work (using the diary format from the skill). https://github.com/wesen/skills/tree/main/brutalist-work-slip"
```

**Assistant interpretation:** Design, publish, implement, and validate the book's elastic dataflow laboratory with task commits, detailed diary, and physical work slips.

**Inferred user intent:** Build a working concurrent symbolic machine and preserve enough explanation and evidence for an intern to understand and maintain it.

### What I did

Read Lab 3 and the existing value, memory, elastic register, and UART implementations. Archived the lab chapter and linked work-slip skill. Printed PLAN and P1 start and created five task phases.

### Why

The new laboratory needs explicit concurrency and epoch ownership contracts before writing RTL; existing sequential search stepping cannot substitute for elastic dataflow.

### What worked

Both print receipts report printed true. Defuddle archived the linked skill. The design names exact interfaces, files, invariants, and tests.

### What didn't work

Web open returned Cache miss for the GitHub tree and raw skill URL; Defuddle successfully fetched the GitHub blob page. Two exploratory reads used nonexistent guessed paths (tagged_pkg.sv and the codex defuddle path); repository discovery and the provided skill catalog resolved the actual files without implementation changes.

### What I learned

An indexed per-node activation reservation can satisfy last-operand capacity closure in the fixed 28-slot machine while avoiding shared ready-FIFO head-of-line dependencies.

### What was tricky to build

Cancellation must invalidate in-flight work without withdrawing a blocked externally offered result. The design therefore backpressures cancellation at that output boundary and enforces global drain on epoch wrap.

### What warrants a second pair of eyes

Review activation storage, output cancellation ownership, epoch wrap, and control-plane clock enable before RTL.

### What should be done in the future

Validate and upload guide, then implement P2 executable models.

### Code review instructions

Run docmgr doctor for GATEMATE-SYMBOLIC-007; inspect the guide and archived lab chapter; confirm upload receipt before P2.

### Technical details

Baseline 7566576. Scope: four contexts, seven fixed nodes including optional COPY fanout, eight-bit epochs, 40-bit typed values in an 80-bit envelope. Eleven slips total for plan plus five start/done pairs.

P1 delivery: docmgr doctor passed; dry-run succeeded; real upload returned `OK: uploaded GATEMATE 007 Elastic Dataflow Intern Guide.pdf -> /ai/2026/09/04/GATEMATE-SYMBOLIC-007`. Implementation starts only after this receipt.

## Step 2: Executable models and IDE scope expansion

Implemented and tested the typed semantic and bounded transaction models. The user added a full React/Go IDE; created a separate 2557-word IDE design and extended the plan with P6 implementation and P7 integration. Debug visibility is now a P3/P4 requirement.

### Prompt Context

**User prompt (verbatim, JSON encoded):**

```json
"Design a full react + golang IDE for the dataflow engine as well. Create a separate design doc for it. I don't know when it makes sense for you to implement it, so you decide."
```

**Assistant interpretation:** Design, publish, implement, and validate the book's elastic dataflow laboratory with task commits, detailed diary, and physical work slips.

**Inferred user intent:** Build a working concurrent symbolic machine and preserve enough explanation and evidence for an intern to understand and maintain it.

### What I did

Added pkg/dataflow types, semantic evaluator, transaction model, and tests. Added IDE design for scenario editing/storage, context controls, real snapshots, pipeline/queue inspection, result assertions, and historical viewing. Extended phase helper and task plan.

### Why

The IDE should be implemented after the engine and physical protocol are qualified, but required debug access must be designed before RTL implementation.

### What worked

All seven model test groups passed on the initial ordinary run. Race tests and vet then passed with normal Go cache access. Coverage includes 240 latency/depth/seed cases, 100 randomized four-context expressions, 1000 arithmetic/envelope cases, COPY, isolated faults, in-flight cancellation, held output, and reduced epoch wrap.

### What didn't work

The first sandboxed race command failed before running tests: open /home/manuel/.cache/go-build/de/dec637d79548ebaffb5b5b86fb57ccea5d945cb2f0961c0927dc1400588a38d6-d: read-only file system. Authorized normal-cache execution passed; no software repair was needed.

### What I learned

The model drains with depth-one queues in the tested cases, while per-node activation reservations retain newly ready work. The IDE requires paused debug pages rather than reconstructed imaginary physical state.

### What was tricky to build

A full scenario IDE is useful with fixed descriptors: editable source schedules and assertions are supported, while topology editing must wait for a genuine graph-loader feature. Cancellation tests need explicit enabled-cycle control on the physical board.

### What warrants a second pair of eyes

Review live-versus-historical state, paged snapshot atomicity, source identity, model fault isolation, and exactly-once activation.

### What should be done in the future

Upload IDE guide and revised plan; commit P2; implement RTL and debug ports in P3.

### Code review instructions

go test -race ./pkg/dataflow -count=1 -v; go vet ./pkg/dataflow; docmgr doctor. Inspect P2-model-tests.log and both guide documents.

### Technical details

The original eleven-slip plan is preserved. The expanded plan has seven phase pairs plus original/revised plan receipts, for sixteen total. P1 remains complete; P2 model tests complete; IDE implementation is deliberately scheduled after P5 engine qualification. The subsequent user prompt was: "or maybe you already are doing that".

IDE delivery receipt: `OK: uploaded GATEMATE 007 Dataflow IDE Design.pdf -> /ai/2026/09/04/GATEMATE-SYMBOLIC-007`. Revised plan receipt reports `printed: true`.

## Step 3: Elastic RTL and paused inspection simulation

Implemented the fixed graph as a bounded RTL engine with operand RAM, activation reservations, separate arithmetic pipelines, completion routing, context cancellation, and debug pages. The first directed simulation matrix passed all twelve queue-depth/multiplier-latency configurations. Synthesis is the next qualification gate; simulation alone does not establish device fit or timing.

### Prompt Context

**User prompt (verbatim, JSON encoded):**

```json
"(see Step 1)"
```

**Assistant interpretation:** Design, publish, implement, and validate the book's elastic dataflow laboratory with task commits, detailed diary, and physical work slips.

**Inferred user intent:** Build a working concurrent symbolic machine and preserve enough explanation and evidence for an intern to understand and maintain it.

### What I did

Added dataflow_pkg.sv, df_fifo.sv, df_unit.sv, and dataflow_core.sv. Added a self-checking RTL testbench and scripts/test.sh. Printed P2 completion and P3 start slips after model commit 77802e8. The IDE guide upload reported success.

### Why

The physical IDE needs access to actual stopped machine state, including synchronous operands and already captured issue operands. Cancellation must preserve an offered output while invalidating other old-epoch work.

### What worked

All initial RTL simulations passed: DEPTH 1/2/8 crossed with LATENCY 1/2/4/8. Checks include outputs 58 and 12, twelve unique activations, 100-cycle output hold, rejection of cancel for that held context, cancellation inside the multiplier, old-input and old-completion suppression, COPY fanout producing 78, isolated duplicate fault, paused RAM inspection, and reduced-width epoch wrap drain guard.

### What didn't work

The first tmux synthesis launch failed before synthesis: error connecting to /tmp/tmux-1000/default (Operation not permitted). Retried using authorized normal tmux access. No RTL repair was required for the simulation matrix.

### What I learned

A paused debug RAM read can change the memory response while issue is waiting. Tracking the previous read address prevents capturing a response belonging to a debug slot after resume; captured operands remain private issue registers.

### What was tricky to build

External input and router deliveries share one commit decision each cycle. Per-slot pending activation storage means the commit cannot deadlock waiting for a shared ready FIFO. The first context fault clears pending work after other same-edge scheduling assignments. Output staleness cleanup remains available while compute is paused.

### What warrants a second pair of eyes

Review cancellation priority, fault versus issue ordering, same-edge queue events, operand read ownership, and the combinational round-robin selection cost in synthesis.

### What should be done in the future

Check synthesis resources; add UART control and Go snapshots; broaden randomized differential tests during engine qualification.

### Code review instructions

Source /home/manuel/fpga/oss-cad-suite/environment, then elastic_dataflow/scripts/test.sh. Inspect reference/validation/P3-rtl-first.log. Synthesis runs through elastic_dataflow/scripts/check-synthesis.sh in tmux.

### Technical details

Debug pages: 0 capabilities; 1 context/queue status; 2..9 counters; 10..11 issue; 12..13 router; 16..23 multiplier stages; 24 validity; 25 ALU; 32..59 operand pairs; 64..91 slot flags; 96..103 input queue; 112..119 completion queue; 128..135 output queue; 144..147 pending errors. Read operand pages only while paused and hold the address for at least two clocks before sampling. Activation traces describe unit admission; model and RTL microcycle counts are intentionally not equated.

Synthesis checkpoint failure (first attempt): `rtl/df_unit.sv:8: ERROR: syntax error, unexpected TOK_IMPORT`. The installed Yosys Verilog frontend rejects the wildcard package import accepted by Icarus. Existing project RTL uses explicit package-qualified references. Applied that convention with the retained one-time script `elastic_dataflow/scripts/qualify-package.py`, then restarted synthesis. Original failure log is retained as `P3-synthesis-import-failure.log`.

## Step 4: Synthesis-oriented scheduler representation

Qualified the wildcard-import repair and reduced the scheduler's synthesis cost. The first synthesizable form expanded a run-time signed integer quotient/remainder into hundreds of thousands of intermediate cells. Rewrote selection using constant slot indices and a pair of bounded priority passes, preserving circular ordering and unit availability filtering.

### Prompt Context

**User prompt (verbatim, JSON encoded):**

```json
"(see Step 1)"
```

**Assistant interpretation:** Design, publish, implement, and validate the book's elastic dataflow laboratory with task commits, detailed diary, and physical work slips.

**Inferred user intent:** Build a working concurrent symbolic machine and preserve enough explanation and evidence for an intern to understand and maintain it.

### What I did

Replaced variable idx division/modulo with elaboration-time scan indices. Added explicit four-range decoding from selected slot to context and subtraction for node. Stopped the oversized synthesis run after its log reported 343860 cells during optimization. Retained that log and the one-time rewrite script. Re-ran all twelve directed simulation configurations successfully.

### Why

The graph has exactly 28 slots. General signed division is unnecessary hardware for this fixed domain. Selecting among constant-index requests expresses the intended bounded logic directly.

### What worked

The import repair reached synthesis optimization; the scheduler rewrite retained all previously tested functional behavior on its first simulation run.

### What didn't work

The initial scheduler representation was impractically large: synthesis reported Computing hashes of 343860 cells of module dataflow_core. This is an implementation cost finding, not a functional simulation failure. The run was stopped before mapping completion and replaced by the bounded implementation.

### What I learned

A software-style variable index followed by division can obscure a small fixed hardware table. Elaboration-time division of loop constants is cheap; division of the rotated run-time index is not.

### What was tricky to build

Two descending priority passes preserve round-robin semantics: first choose the lowest eligible slot, then override it with the lowest eligible slot at or above next_slot. The first pass supplies wraparound when the second has no eligible entry.

### What warrants a second pair of eyes

Inspect generated resource counts and critical paths, and confirm selection order when next_slot points at slot 27 and when one unit is blocked.

### What should be done in the future

Finish synthesis qualification, then implement framed UART control and physical execution.

### Code review instructions

elastic_dataflow/scripts/test.sh after loading the CAD environment; compare P3-rtl-fixed-scheduler.log with the prior matrix. Inspect core-synthesis.log from the fixed run.

### Technical details

Simulation checkpoint commit: 9fffdc3. The new selection uses pending[scan] and closed[scan/7] with compile-time scan values, plus a five-bit next-slot comparison. selected_context decodes ranges 0..6, 7..13, 14..20, 21..27. selected_node subtracts seven times that two-bit context. No public token format changes.

P3 synthesis qualification completed: the fixed scheduler run completed in 8.19 seconds with two `CC_BRAM_20K` operand memories, 3011 `CC_DFF`, 3864 `CC_L2T4`, and 1202 `CC_L2T5`. This is the standalone core with every debug/trace port retained, not a placed board design. The synthesis check reported zero structural problems. Assigned the procedural error-scan loop variable a default to remove irrelevant inferred-loop-variable latch warnings; the final simulation matrix and synthesis were re-run. Full evidence: `P3-rtl-final.log` and `P3-synthesis-final.log`. Physical place/route, UART, and board execution remain P4 work.

P4 first UART simulation failure: `FATAL: sim/tb_dataflow_link.sv:58: input did not remain paused 00000000000600000000`. The status frame correctly placed input count six at bits 39:32; the new test mistakenly checked ready-count bits 31:24. Corrected that assertion (first repair) and reran the UART test. Original evidence is preserved in `P4-uart-first.log`.

P4 initial place/route failure: `ERROR: Max frequency for clock 'link.clk': 7.80 MHz (FAIL at 10.00 MHz)`. The critical path crossed the serial 28-addend ready population count (core line 158) and the debug response checksum, taking 128.24 ns. The first timing repair replaces this count with a balanced pair/quad/octet/half sum tree. The required 10 MHz constraint is unchanged. Original placed timing report retained as `P4-timing-first.log`; both RTL test suites and physical route are rechecked after this change.

P4 Go host and CLI validation: initial domain/protocol tests passed. A later sandbox run failed before compilation with `open /home/manuel/.cache/go-build/90/9010652bc4b87b703a4be1f8e485794547a15e7bb9a0c4ff1d0fb1d7b46b65d0-d: read-only file system`; reran with authorized cache access. The first CLI test then exposed two incorrect assumptions in the test: Glazed 1.4.2 writes structured output to `os.Stdout`, not Cobra's SetOut writer, and explicitly normalizes context.Canceled to successful shutdown (`pkg/cli/cobra.go:167,183`). Reworked the test to capture domain row emission, settings decoding, processor-error propagation, and absence of output under cancellation. The domain already honored context cancellation. Original test errors are retained in `P4-cli-contract-failure.log`.

Large P3 synthesis logs were losslessly compressed to `.log.gz` by retained script `06-compress-evidence.py`; earlier diary references without `.gz` name the same evidence. Use `gzip -cd` to consult them.

## Step 5: UART control and direct Go engine ownership

Added a stop-and-wait UART control plane, a Go serial engine, complete snapshot decoding, and a Glazed example runner. The UART and host tests now pass; the physical bitstream is being routed after a debug population-count timing repair. Added a concrete API/register reference to connect the intern guide to the implemented contract.

### Prompt Context

**User prompt (verbatim, JSON encoded):**

```json
"(see Step 1)"
```

**Assistant interpretation:** Design, publish, implement, and validate the book's elastic dataflow laboratory with task commits, detailed diary, and physical work slips.

**Inferred user intent:** Build a working concurrent symbolic machine and preserve enough explanation and evidence for an intern to understand and maintain it.

### What I did

Implemented reset/inject/tick/cancel/query/poll framing, checksums, timeout handling, bounded ticks, explicit FIFO-full and cancel-blocked responses. Added Engine.Execute, detached model snapshots, serial snapshots, uncertainty handling, and book/copy/fault/cancel examples. Added UART, protocol, snapshot, serial-rejection, CLI decoding, row emission, and processor-error tests.

### Why

A single validated operation type serves CLI execution and the future HTTP/scenario boundary. The serial owner must never infer whether a timed-out mutation ran, nor publish only part of a hardware snapshot.

### What worked

UART tests pass after correcting the test's input-count byte. Go model/protocol/CLI race tests pass after aligning tests with the inspected Glazed 1.4.2 contract. Both RTL suites pass after balancing the ready-count sum tree. Physical route remains in progress.

### What didn't work

Exact UART, cache-permission, Glazed-test, and initial 7.80 MHz timing failures are recorded immediately above with retained logs. No failed bitstream was programmed. Glazed emits to os.Stdout and normalizes context.Canceled; the original tests assumed Cobra SetOut and a propagated cancellation error.

### What I learned

Debug logic participates in timing closure. The ready-count chain feeding UART response checksum dominated a physical path even though computation itself was paused during inspection.

### What was tricky to build

Cancellation selection needs a following control cycle before reading cancel_ready. Synchronous debug operands need a stable address for multiple clocks. Snapshot publication must occur only after every page arrives, while a transport failure requires explicit reset.

### What warrants a second pair of eyes

Review frame lengths/checksum coverage, real versus modeled snapshot fields, request uncertainty semantics, and the enabled-cycle acknowledgement edge.

### What should be done in the future

Finish route and verify all four examples on the physical board; then stress/debug qualification and the React IDE.

### Code review instructions

Load CAD environment; run elastic_dataflow/scripts/test.sh and scripts/test-link.sh. Run go test -race ./pkg/dataflow ./cmd/dataflow-lab. Use go run ./cmd/dataflow-lab --engine model --example book --format json. Consult reference/02-dataflow-api-and-debug-register-reference.md.

### Technical details

P3 closing commit 67f661d is printed on the completion receipt; P4 start receipt is confirmed printed. The API guide records every binary envelope byte and debug page. Physical configuration remains immutable at runtime; no compatibility adapters were added. Large synthesis logs are retained as deterministic gzip files.

## Step 6: Physical board execution and randomized qualification

The board routes at 12.80 MHz against the unchanged 10 MHz constraint and was programmed successfully. All four checked examples pass on the physical engine. Expanded P5 qualification in parallel with the last part of P4 routing: 600 randomized RTL schedules across twelve configurations agree with the Go model for every result and activation value. The user requested UI screenshots; the new dataflow UI is not built yet, so screenshots are scheduled for P6/P7 with source/scenario labels in the diary.

### Prompt Context

**User prompt (verbatim, JSON encoded):**

```json
"take screenshots for the diary and for a report later on, if you have a web UI already, else later."
```

**Assistant interpretation:** Design, publish, implement, and validate the book's elastic dataflow laboratory with task commits, detailed diary, and physical work slips.

**Inferred user intent:** Build a working concurrent symbolic machine and preserve enough explanation and evidence for an intern to understand and maintain it.

### What I did

Stopped the prior graph web server using lsof-who -p 8086 -k, loaded the dataflow bitstream with openFPGALoader, and ran book/copy/fault/cancel via the Go serial host. Saved wire logs, decoded snapshots, synthesis and route logs, and a physical counter summary. Added randomized RTL stress and Go differential tests plus explicit opt-in physical qualification tests.

### Why

Arithmetic outputs alone do not verify exactly-once activation or cancellation provenance. Differential activation comparison, a visible occupied multiplier before cancellation, and real debug snapshots provide stronger evidence. Screenshots should document the actual new UI when it exists.

### What worked

Board timing passed at 12.80 MHz; programming reported Done. Book returned 58/12 with 12 activations; COPY returned 78 with seven; isolated duplicate fault left context one's result 12 intact; canceled context two returned epoch-one result 12 and counted two stale discards. All 2400 randomized RTL expressions and 14400 activation values matched the Go model. Ready/valid stability was asserted at output and both unit exits during stalls and pauses.

### What didn't work

The balanced router temporarily repeated a single remaining routing conflict for many iterations, then converged successfully without changing the seed or weakening constraints. Earlier timing failure and its one successful repair are preserved above. No additional repair was needed.

### What I learned

The compiled board has four contexts, seven fixed descriptors, four multiply stages, and depth-eight queues. The checked host schedules consumed 68/48/48/54 enabled cycles for book/copy/fault/cancel respectively; these include the host's bounded tick batches and are not minimal operation latencies.

### What was tricky to build

P5 simulation qualification overlapped the P4 physical route to use the wait productively. Both phase-start receipts were printed before their work. Captured physical state must remain distinguishable from model state in the forthcoming UI and screenshots.

### What warrants a second pair of eyes

Inspect complete physical frames, stale count two in cancellation, final epochs, and the differential parser's requirement for all 50 cases in each of twelve logs. Treat passing simulation and passing physical timing as separate claims.

### What should be done in the future

Finish the physical randomized/held-output/epoch-wrap test, close P4/P5, then implement the IDE and capture model plus FPGA screenshots for this diary and a later report.

### Code review instructions

scripts/08-physical-examples.sh; scripts/09-physical-stress.sh; elastic_dataflow/scripts/test-stress.sh followed by go test ./pkg/dataflow -run TestRTLDifferential -rtl-log-dir=/absolute/path/to/elastic_dataflow/build/stress. Logs under reference/validation; archived RTL traces are gzip files.

### Technical details

Physical load and run checkpoint builds on commit 2e22871. Phase P5 start slip is confirmed printed. Screenshot request is retained verbatim in this step; no dataflow screenshots exist yet. Board source artifacts remain under ignored elastic_dataflow/build and reproducible Makefile/scripts; generated bitstreams are not committed.

The first opt-in physical stress test stopped before resetting or exercising the board: `physical_test.go:29: open ttmp/2026/09/04/GATEMATE-SYMBOLIC-007--laboratory-3-elastic-dataflow-expression-engine/reference/validation/P5-physical-wire.log: no such file or directory`. Go tests run with the package directory as their working directory. Changed the runner to pass an absolute wire-log path (first repair); retained the original failure and reran the same test.

P5 physical stress completed successfully in 14.95 seconds after the absolute-path repair: 32 randomized four-context batches (128 expressions), a held-output cancellation refusal with unchanged output data across 64 further ticks, and 256 drained cancellations wrapping the physical eight-bit epoch to zero. The final log and complete wire evidence are retained. P4/P5 are now complete; P6 start slip is confirmed printed. No dataflow web server has started yet.

P6 initial browser and frontend checks: the browser ran the book scenario to results 58 and 12 and captured `reference/screenshots/01-model-book-initial.png`. The only browser console error was a missing favicon; the entry HTML now declares an empty data favicon. Screenshot inspection showed overlapping A/B edge labels; they now terminate at separate destination-port heights.

The first React suite passed 13/14 tests. The model-configuration test failed with `Unable to find an element with the text: Model configuration` because the initial loading view prematurely displayed TRANSACTION MODEL before its first API response, so the test's source-label wait was insufficient. Changed the loading badge to CONNECTING and made the configuration assertion await its data. Also restricted example initialization to the first load so clearing the source editor does not repopulate it unexpectedly. Original failing log is retained as `P6-react-first-failure.log`.

## Step 7: Go React scenario IDE and model browser verification

Implemented the separate IDE backend and React workspace, preserving the fixed descriptor graph while making execution schedules editable. The browser now runs the book and cancellation scenarios, saves and reloads projects, validates source, inspects historical frames without enabling mutations, and displays a real occupied model pipeline stage. Captured five screenshots for the diary and later report; physical UI captures follow in P7.

### Prompt Context

**User prompt (verbatim, JSON encoded):**

```json
"(see Step 1)"
```

**Assistant interpretation:** Design, publish, implement, and validate the book's elastic dataflow laboratory with task commits, detailed diary, and physical work slips.

**Inferred user intent:** Build a working concurrent symbolic machine and preserve enough explanation and evidence for an intern to understand and maintain it.

### What I did

Added internal/dataflowide scenario validation, confined atomic project storage, a serialized engine session, immutable bounded history, HTTP routes, and embedded assets. Added cmd/dataflow-ide with Glazed flags, loopback binding, zerolog, and errgroup lifecycle. Added React/Redux/RTK Query source editor, graph, node/operand inspector, pipeline/queues, controls, projects, history, counters, and result console. Added independent Vite entry and build generation.

### Why

The user requested a full Go/React IDE and later asked for screenshots. A scenario IDE can exercise the implemented fixed graph honestly without claiming a dynamic graph compiler. Frame identity and reset generations prevent stale browser controls from mutating newer state.

### What worked

Backend ownership/history/project/HTTP tests and race tests pass. All fourteen frontend tests pass, including the prior graph app tests. Production Vite build succeeds. Browser script 15 verified 58/12 results, project save/load, invalid-source execution blocking, historical control disabling, occupied multiply value 42, epoch-one cancellation result 12, and no horizontal overflow at 390 pixels. Current browser console has zero errors or warnings.

### What didn't work

P6 initial Go command build hit the known read-only cache sandbox restriction; rerun with normal cache access passed. First React test exposed an incorrect initial source badge; fixed CONNECTING state and asynchronous test wait. A missing favicon and overlapping graph destination-port labels were fixed after actual screenshot review. Exact earlier failures/logs are recorded above.

### What I learned

The model cancellation screenshot shows value 42 in multiply stage one after six enabled cycles. The corresponding physical stage position may differ; source labels remain explicit. Scenario input actions may include bounded credit ticks when the input queue is full; their frame labels disclose that interval.

### What was tricky to build

Cancellation or a failed exchange during a mutating request must mark engine state uncertain; the UI then requires reset. A source editor must initialize once rather than restoring an example whenever the user clears it. History owns detached snapshots, and reset clears the previous generation. Browser review also motivated clearing old console events on reset and logging each completed action.

### What warrants a second pair of eyes

Review context/source provenance, disabled historical controls plus server-side expected-frame checks, partial-input action errors, atomic rename storage under os.Root, and pause behavior at action boundaries.

### What should be done in the future

Validate the same UI against the physical serial engine, capture physical in-flight/result screenshots, run final repository checks, finish the intern handoff, upload the final document bundle, and push commits.

### Code review instructions

go test -race ./internal/dataflowide ./pkg/dataflow ./cmd/dataflow-ide; pnpm --dir web test; make dataflow-frontend; go run -tags embed ./cmd/dataflow-ide --engine model. Reproduce browser checks with scripts/15-browser-model-checks.js. Screenshots under reference/screenshots.

### Technical details

P4/P5 completion receipts and P6 start receipt are confirmed printed. Screenshots: 01 initial book UI before port-label polish; 02 corrected model book results; 03 historical view with controls disabled; 04 model multiply in flight; 05 mobile viewport. The source of every current image is the transaction model, not hardware. The new server is on 127.0.0.1:8087. The original graph frontend remains its separate entry and API.

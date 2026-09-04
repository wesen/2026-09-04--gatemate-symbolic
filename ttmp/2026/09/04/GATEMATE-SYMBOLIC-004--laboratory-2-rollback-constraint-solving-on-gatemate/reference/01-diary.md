---
Title: Diary
Ticket: GATEMATE-SYMBOLIC-004
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
    - Path: repo://queens_rollback/Makefile
      Note: Lab 2 implementation or reproducible verification and diary tooling
    - Path: repo://queens_rollback/README.md
      Note: Lab 2 implementation or reproducible verification and diary tooling
    - Path: repo://queens_rollback/rtl/queens_core.sv
      Note: |-
        Snapshot and trail recovery with one mutation owner
        Lab 2 implementation or reproducible verification and diary tooling
    - Path: repo://queens_rollback/rtl/queens_result_printer.sv
      Note: Lab 2 implementation or reproducible verification and diary tooling
    - Path: repo://queens_rollback/rtl/queens_top.sv
      Note: Lab 2 implementation or reproducible verification and diary tooling
    - Path: repo://queens_rollback/rtl/queens_types_pkg.sv
      Note: Lab 2 implementation or reproducible verification and diary tooling
    - Path: repo://queens_rollback/scripts/make_synth.py
      Note: Lab 2 implementation or reproducible verification and diary tooling
    - Path: repo://queens_rollback/sim/conftest.py
      Note: Lab 2 implementation or reproducible verification and diary tooling
    - Path: repo://queens_rollback/sim/tb_queens.sv
      Note: Lab 2 implementation or reproducible verification and diary tooling
    - Path: repo://queens_rollback/sim/tb_queens_top.sv
      Note: Lab 2 implementation or reproducible verification and diary tooling
    - Path: repo://queens_rollback/sim/test_model.py
      Note: Lab 2 implementation or reproducible verification and diary tooling
    - Path: repo://queens_rollback/sim/test_rtl.py
      Note: |-
        Complete live state and output comparisons
        Lab 2 implementation or reproducible verification and diary tooling
    - Path: repo://queens_rollback/sim/test_top.py
      Note: Lab 2 implementation or reproducible verification and diary tooling
    - Path: repo://queens_rollback/tools/oracle.py
      Note: Lab 2 implementation or reproducible verification and diary tooling
    - Path: repo://queens_rollback/tools/queens_model.py
      Note: |-
        Stepwise semantic model and complete event snapshots
        Lab 2 implementation or reproducible verification and diary tooling
    - Path: repo://ttmp/2026/09/04/GATEMATE-SYMBOLIC-004--laboratory-2-rollback-constraint-solving-on-gatemate/scripts/01-initialize-scope.py
      Note: Lab 2 implementation or reproducible verification and diary tooling
    - Path: repo://ttmp/2026/09/04/GATEMATE-SYMBOLIC-004--laboratory-2-rollback-constraint-solving-on-gatemate/scripts/02-print-phase.py
      Note: Lab 2 implementation or reproducible verification and diary tooling
    - Path: repo://ttmp/2026/09/04/GATEMATE-SYMBOLIC-004--laboratory-2-rollback-constraint-solving-on-gatemate/scripts/03-diary.py
      Note: Lab 2 implementation or reproducible verification and diary tooling
    - Path: repo://ttmp/2026/09/04/GATEMATE-SYMBOLIC-004--laboratory-2-rollback-constraint-solving-on-gatemate/scripts/04-profile-backends.py
      Note: Lab 2 implementation or reproducible verification and diary tooling
    - Path: repo://ttmp/2026/09/04/GATEMATE-SYMBOLIC-004--laboratory-2-rollback-constraint-solving-on-gatemate/scripts/05-hardware.py
      Note: Lab 2 implementation or reproducible verification and diary tooling
    - Path: repo://ttmp/2026/09/04/GATEMATE-SYMBOLIC-004--laboratory-2-rollback-constraint-solving-on-gatemate/scripts/06-hardware.sh
      Note: Lab 2 implementation or reproducible verification and diary tooling
    - Path: repo://ttmp/2026/09/04/GATEMATE-SYMBOLIC-004--laboratory-2-rollback-constraint-solving-on-gatemate/scripts/07-finalize-evidence.py
      Note: Lab 2 implementation or reproducible verification and diary tooling
    - Path: repo://ttmp/2026/09/04/GATEMATE-SYMBOLIC-004--laboratory-2-rollback-constraint-solving-on-gatemate/scripts/08-relate-files.py
      Note: Lab 2 implementation or reproducible verification and diary tooling
    - Path: repo://ttmp/2026/09/04/GATEMATE-SYMBOLIC-004--laboratory-2-rollback-constraint-solving-on-gatemate/scripts/09-report-examples.py
      Note: Lab 2 implementation or reproducible verification and diary tooling
    - Path: repo://ttmp/2026/09/04/GATEMATE-SYMBOLIC-004--laboratory-2-rollback-constraint-solving-on-gatemate/scripts/10-backfill-diary.py
      Note: Lab 2 implementation or reproducible verification and diary tooling
    - Path: repo://ttmp/2026/09/04/GATEMATE-SYMBOLIC-004--laboratory-2-rollback-constraint-solving-on-gatemate/scripts/11-export-vault-report.py
      Note: Lab 2 implementation or reproducible verification and diary tooling
    - Path: repo://ttmp/2026/09/04/GATEMATE-SYMBOLIC-004--laboratory-2-rollback-constraint-solving-on-gatemate/scripts/12-record-report-delivery.py
      Note: Lab 2 implementation or reproducible verification and diary tooling
ExternalSources: []
Summary: Chronological investigation and implementation diary for Laboratory 2.
LastUpdated: 2026-09-04T16:57:27.323426618-04:00
WhatFor: ""
WhenToUse: ""
---




# Diary

## Goal

Establish and implement the book's rollback constraint solver while preserving source evidence, decisions, experiments, and validation results.

## Step 1: Reread the lab and establish its scope

I reread the complete Laboratory 2 chapter and its supporting checkpoint/trail chapter before creating this ticket. The project is a fixed eight-queens search engine with explicit reversible domain updates, deterministic alternatives, and accepted result commitment. This step records scope only; no solver RTL or model was implemented.

### Prompt Context

**User prompt (verbatim):**

> Ok, new ticket for the next lab project. Read the lab book again to refresh your memory, and tell me what th eproject is in your own words. (2 paragraphs or so)

**Assistant interpretation:** Identify the next worked lab from the archived book, open its ticket, and briefly explain the project and learning objective.

**Inferred user intent:** Begin the next lab with a refreshed, source-grounded understanding before implementation.

### What I did

- Located the book in GATEMATE-SYMBOLIC-001 sources and inspected the lab sequence.
- Read Laboratory 2, including its experimental sequence and extensions, and the checkpoint/trail substrate.
- Created GATEMATE-SYMBOLIC-004 with a project brief, acceptance criteria, planned tasks, and this diary.

### Why

The next lab introduces semantic rollback across search decisions. Its requirements must come from the book rather than assumptions about extending the previous CPU.

### What worked

The book states exact results: 92 boards, first board `[0,4,7,5,2,6,1,3]`, and FIRST_ONLY cut after accepted output. It also prescribes a full-snapshot baseline before the trail implementation, enabling a measured comparison.

### What didn't work

No failures. The ticket identifier was checked and did not already exist.

### What I learned

The propagated bitmap is restart state, not disposable scratch. The choice stack and mutation trail serve different purposes, and result acceptance is separate from reversible search mutation.

### What was tricky to build

No implementation in this step. The design must preserve log-before-write ordering, publish a complete choice before the branch mutation, restore domains in reverse order, and defer cut until the first solution is accepted.

### What warrants a second pair of eyes

Check the planned model's propagation order and zero-domain handling against the book. Do not turn optional N-queens generalization, heuristics, or multiple contexts into baseline requirements.

### What should be done in the future

Write the design, then implement the independent reference and full-state snapshot baseline before introducing the trail.

### Code review instructions

Read this ticket index alongside book lines 4569–4961 and 3724–3884. Run docmgr doctor for GATEMATE-SYMBOLIC-004. There are no new runtime changes to test.

### Technical details

Domains are eight eight-bit masks; trail and choice records map to separate RAMs. Solutions pack eight three-bit rows into 24 bits. The book asks to simplify around two RAM blocks and 3,000 CPEs. All new scripts are stored in this ticket's scripts folder; the existing source book is related rather than duplicated.

## Step 2: P1 explicit semantic model and oracle

Implemented a production stepwise model with snapshot and trail recovery, plus an independent row-placement oracle. The overall plan and P1 start slips printed successfully before implementation.

### Prompt Context

**User prompt (verbatim):**

> Now implement it, commit at appropriate intervals and keep a detailed diary as you work (using the diary format from the skill)
>
> . Print out a brutalist work slip with the plan / different phases for the ticket. then before stsarting a phase, plrint a split about the phase, and print one when the phase is done.

**Assistant interpretation:** Implement the Lab 2 guide in tested phases, with committed milestones and printed boundaries.

**Inferred user intent:** Deliver a working and measurable rollback solver with reproducible evidence and a detailed implementation record.

### What I did

Added queens_rollback/tools/oracle.py, queens_model.py and model tests. Events include complete domains, propagated bitmap, live choice records, live trail records, accepted output and fault state.

### Why

The RTL needs an executable semantic boundary contract and an independent complete solution stream.

### What worked

16 model tests passed, covering exact enumeration, first-result cut, stable blocked output, helper truth tables, and small logical capacities.

### What didn't work

No model failures. Initial read-only discovery noted the repository has no root .gitignore or AGENTS.md; project-local ignores are used.

### What I learned

The smallest measured sufficient capacities are 32 trail entries and six choices; defaults remain conservatively 64 and eight.

### What was tricky to build

The model separates entering a pending result from accepting it, preserving search state while blocked. Choice entries include complete checkpoint domains for independent restoration assertions even in trail mode.

### What warrants a second pair of eyes

Review event snapshots and first-only cut at acceptance. The oracle uses pairwise geometry rather than domains.

### What should be done in the future

Implement the snapshot RTL and compare full live state at semantic events.

### Code review instructions

Run python3 -m pytest sim/test_model.py -q from queens_rollback.

### Technical details

P1 model tests: 16. Snapshot and trail share deterministic propagation order but have distinct restore events. No Lab 1 runtime files changed.

## Step 3: P2 synchronous snapshot RTL baseline

Implemented the deterministic snapshot controller and synchronous 104-bit checkpoint memory. The model and RTL now agree on every observed domain state and every live checkpoint, with complete ordered outputs.

### Prompt Context

**User prompt (verbatim):**

(see Step 2)

**Assistant interpretation:** Implement the Lab 2 guide in tested phases, with committed milestones and printed boundaries.

**Inferred user intent:** Deliver a working and measurable rollback solver with reproducible evidence and a detailed implementation record.

**Commit (code/work):** cf52672

### What I did

Added queens_types_pkg.sv, queens_core.sv, tb_queens.sv and differential RTL tests. Reused the existing synchronous RAM module directly. Updated choice-record publication to emit its event on the memory-write edge.

### Why

A tested snapshot baseline isolates search semantics before changing restoration storage. Complete checkpoint comparison catches errors hidden by solution counts.

### What worked

23 combined tests pass: 16 model and seven snapshot RTL cases, including stalls, FIRST_ONLY and choices 0/1/5.

### What didn't work

No RTL compile or simulation failure. The P1 commit check initially rejected blank quoted lines with trailing spaces; the diary helper now avoids them.

### What I learned

Checkpoint creation can write an inactive slot before publishing its top. Updating an already live checkpoint must expose its event on the actual write edge.

### What was tricky to build

Synchronous checkpoint reads have explicit wait and capture states. Snapshot restore publishes domains and propagated bitmap together. The shared controller will select a storage backend in P3 rather than duplicate the propagation algorithm.

### What warrants a second pair of eyes

Check full C record comparison and result-ready ownership in the bench. The P1 slip used HEAD as its reference; subsequent slips resolve immutable hashes.

### What should be done in the future

Add trail storage and reverse restoration to the same controller with compile-time backend selection.

### Code review instructions

Source the documented CAD environment and run python3 -m pytest sim -q in queens_rollback.

### Technical details

Snapshot records are 104 bits: 64 domain bits plus the proposed 40-bit choice metadata. No production board integration yet.

## Step 4: P3 trail backend and measured recovery comparison

Implemented trail storage and reverse restoration behind the same deterministic controller. Complete semantic event and live-record comparisons pass for both backends, including near-capacity runs.

### Prompt Context

**User prompt (verbatim):**

(see Step 2)

**Assistant interpretation:** Implement the Lab 2 guide in tested phases, with committed milestones and printed boundaries.

**Inferred user intent:** Deliver a working and measurable rollback solver with reproducible evidence and a detailed implementation record.

**Commit (code/work):** 0f15376

### What I did

Added compile-time USE_TRAIL and TRAIL_CAPACITY, 20-bit synchronous trail memory, 40-bit choice metadata, log-before-apply sequencing, reverse reads, and cut base publication. Added backend profiling.

### Why

Sharing search control keeps the experiment focused on recovery representation while allowing synthesis to remove the unused backend.

### What worked

36 tests pass. Unstalled complete enumeration takes 53951 snapshot cycles and 74523 trail cycles, each producing all 92 ordered boards and 3980 effective domain writes.

### What didn't work

No P3 compile or simulation failures.

### What I learned

The trail is slower and writes more history for this eight-byte domain state. Logical snapshot payload accounting differs from physical full-record requests, especially when updating remaining alternatives.

### What was tricky to build

Trail entries are physically stored in LOG_WRITE and become reachable with the domain update at APPLY. Reverse restoration captures synchronous reads before updating domain and decrementing the top together.

### What warrants a second pair of eyes

Check mark bounds, FIRST_ONLY base changes, and exact full live trail comparison. Read physical request metrics separately from logical payload metrics.

### What should be done in the future

Add per-cycle stability, reset interruption, integrity injection, and stronger result stall/cut tests.

### Code review instructions

Run python3 -m pytest sim -q and ticket scripts/04-profile-backends.py under the documented CAD environment.

### Technical details

Default logical trail64/choice8; observed trail32/choice6. Snapshot and trail output order match the independent oracle. First-only retains the winning trail and discards all choices.

## Step 5: P4 cycle stability reset cut and integrity checks

Strengthened verification from event equality to every-cycle live-state stability. Added reset interruption and corruption tests, and terminal dwell checks that ensure cut and faults cannot resume work.

### Prompt Context

**User prompt (verbatim):**

(see Step 2)

**Assistant interpretation:** Implement the Lab 2 guide in tested phases, with committed milestones and printed boundaries.

**Inferred user intent:** Deliver a working and measurable rollback solver with reproducible evidence and a detailed implementation record.

**Commit (code/work):** 214a39a

### What I did

Added monitors for domains, propagated bitmap, counts and every live choice/trail entry between semantic events and on faults. Added reset at log-write, restore-apply, pending result and checkpoint write; injected corrupt marks, trail flags and source masks.

### Why

Matching event snapshots alone can miss premature publication or state changes that are repaired before the next event.

### What worked

All 34 RTL tests pass in 29.38 seconds. The unchanged model suite previously passed 16 tests. Long result stalls preserve state, and the bench observes 64 additional terminal cycles without another transfer or event.

### What didn't work

No P4 failures or repair retries.

### What I learned

Corruption tests need a one-cycle monitor exemption for the deliberate testbench mutation, followed by normal stability checks through the fault event.

### What was tricky to build

Reset discards all live memory reachability and the comparison restarts from a fresh model after the RESTART marker. The physical RAM arrays remain uncleared.

### What warrants a second pair of eyes

Inspect the integrity guards before restore writes and the distinction between injected test state and subsequent protected behavior. BAD_DOMAIN_INDEX remains reserved because the implemented domain interface is exactly three bits.

### What should be done in the future

Integrate board formatter, synthesize both backends, capture output and compare routed resources.

### Code review instructions

Source the CAD environment and run python3 -m pytest sim/test_rtl.py -q -x in queens_rollback.

### Technical details

Integrity faults cover corrupt trail marks and record flags; BAD_ONEHOT is checked before propagation. Complete model comparisons include all live records after reset.

## Step 6: P5 board integration and UART verification

Integrated the rollback solver with the existing board reset and UART blocks. The formatter holds each accepted board until its complete record is transmitted and emits an explicit completion count or fault record.

### Prompt Context

**User prompt (verbatim):**

(see Step 2)

**Assistant interpretation:** Implement the Lab 2 guide in tested phases, with committed milestones and printed boundaries.

**Inferred user intent:** Deliver a working and measurable rollback solver with reproducible evidence and a detailed implementation record.

### What I did

Added queens_top, queens_result_printer, UART-level simulation, and reproducible synthesis commands.

### Why

Physical verification requires observable terminal status and an output protocol that tolerates the UART being slower than the solver.

### What worked

All eight UART integration tests passed: both storage modes, enumeration, cut, resource faults, and reset during transmission.

### What didn't work

None. Command: source /home/manuel/fpga/oss-cad-suite/environment; python3 -m pytest sim/test_top.py -q -x. Result: 8 passed in 3.83s.

### What I learned

Explicit D and F records make completion and resource exhaustion distinguishable from silence.

### What was tricky to build

The UART samples start/data synchronously. The formatter advances only on UART readiness and retains a final wait state through the last transmitted byte.

### What warrants a second pair of eyes

Review final-byte ownership, button reset synchronization, and terminal-record priority after the last accepted result.

### What should be done in the future

Synthesize and route both modes; load full and cut variants on the board and compare captured bytes with the model.

### Code review instructions

Run make test in queens_rollback after sourcing the OSS CAD Suite environment. Read sim/test_top.py and rtl/queens_result_printer.sv.

### Technical details

Protocol: Q:672BE0 CRLF for a 24-bit board, D:0000005C CRLF for a count, F:01 CRLF for a fault. Board clock 10 MHz; UART 115200 baud.

## Step 7: P5 physical execution and measured comparison

Completed synthesis, place and route, and physical UART verification for snapshot enumeration, trail enumeration, and trail FIRST_ONLY. All three images produced exactly the modeled bytes, including terminal counts; the board now holds the first-only trail image.

### Prompt Context

**User prompt (verbatim):**

(see Step 2)

**Assistant interpretation:** Implement the Lab 2 guide in tested phases, with committed milestones and printed boundaries.

**Inferred user intent:** Deliver a working and measurable rollback solver with reproducible evidence and a detailed implementation record.

**Commit (code/work):** ca666fd

### What I did

Ran the complete 58-test suite. Added scripts/05-hardware.py and 06-hardware.sh for reproducible tmux builds and serial capture, and 07-finalize-evidence.py for measurement extraction. Archived build logs, image hashes, UART binaries, complete first-result simulation traces, and the implementation README.

### Why

The prior laboratory demonstrated that simulation alone does not establish correct synthesized behavior. Comparing every byte after programming validates the deployed result sequence and explicit completion status.

### What worked

58 passed in 51.71s. Snapshot and trail each captured 932 bytes: 92 Q records and D:0000005C. Trail-first captured 22 bytes: Q:672BE0 and D:00000001. Final routed Fmax: 30.19, 30.98, and 33.72 MHz respectively, all passing 10 MHz.

### What didn't work

No software, synthesis, routing, or hardware validation failures. A read-only status command checked P5-hardware.exit before creation and zsh reported no matches found; the build continued normally and eventually wrote exit 0.

### What I learned

Trail reduced mapped RAM_HALF from three to two and CPE_LT from 2221 to 2070, but increased full-search cycles from 53951 to 74523 and complete-record write traffic from 69888 to 106480 bits. Tiny-board snapshots are cheaper in time and traffic.

### What was tricky to build

The mapped resource units differ from the lab shorthand: nextpnr reports RAM_HALF, CPE_LT, and CPE_FF, not whole blocks or whole CPEs. The report preserves those units. UART capture begins before configuration so startup results cannot be lost. Internal state traces are verified in simulation; physical captures establish external behavior only.

### What warrants a second pair of eyes

Review log-before-write state ordering, result handshake versus cut, and synchronous RAM wait/capture boundaries. Compare P5-measurements.json to the final frequency entry in each nextpnr log, rather than the optimistic pre-route entry.

### What should be done in the future

Review the implementation and measured tradeoff before extending the constraint problem. No additional optimization is required for the present fixed-eight-queens scope.

### Code review instructions

Source the OSS CAD Suite environment and run make test in queens_rollback. Read queens_rollback/README.md; inspect backend-profile.json, snapshot-first-solution-trace.log, trail-first-solution-trace.log, and P5-hardware.json. Physical reproduction: run ticket scripts/06-hardware.sh in tmux with exclusive access to the board UART.

### Technical details

Implementation commits: cf52672 model/oracle; 0f15376 snapshots; 214a39a trail; 1a5194f boundary verification; ca666fd board integration. Yosys 0.68+130 dd83bbad2-dirty; router2 seed 2; 10 MHz board clock. Choice high-water 6, trail high-water 32. Two 8-second enumeration captures and one 8-second cut capture matched exactly.

The measured-results package was committed as `6c7b2d3`. The final P5 completion slip printed successfully at 2026-09-04T22:01:44Z and links that commit. All eleven requested receipts report `printed: true`: one overall plan and a start/completion pair for each of five phases. `docmgr doctor --ticket GATEMATE-SYMBOLIC-004 --fail-on error` passed; all seven ticket tasks are checked. The ticket remains available for implementation review.

## Step 8: Retrospective P1: executable state and independent answer stream

This entry backfills the implementation detail behind Step 2 from the committed model, tests, and retained print receipts. It is retrospective documentation, not a newly executed implementation phase. The model made domain writes, checkpoint publication, restoration, and result acceptance independently observable before those operations were assigned clock cycles.

### Prompt Context

**User prompt (verbatim):** "Backfill your diary in detail, btw."

**Assistant interpretation:** Expand the historical record from committed implementation and retained evidence; distinguish retrospective reconstruction from contemporaneous observations.

**Inferred user intent:** Preserve enough detail to understand decisions, reproduce validation, and review the finished solver.

**Commit (code):** cf52672 — feat: add stepwise queens model and independent oracle (2026-09-04T17:39:32-04:00)

### What I did

- Added the independent recursive coordinate solver in tools/oracle.py, then the generator-based Machine in tools/queens_model.py. The oracle checks row and diagonal conflicts without domain propagation.
- Defined Event and Fault encodings and E/C/T frames containing domains, propagation bits, pointers, outputs, and complete live records.
- Represented output preparation as a pending state: step(False) cannot resume the generator past acceptance.
- Retained checkpoint domain tuples in the trail model for exact restoration assertions; these are verification copies, not proposed trail hardware storage.

### Why

The solution set needs an independent oracle, while the microarchitecture needs an exact state-transition specification. Separating those responsibilities reduces the chance that identical rollback mistakes in model and hardware would be accepted merely because their outputs agree.

### What worked

The contemporaneous Step 2 records 16 passing model tests. The committed tests cover both storage modes, full enumeration, first-only, helper truth tables, blocked output, and reduced capacities. The checked report examples later reconfirm the complete ordered output against the independent oracle.

### What didn't work

No model test failure is recorded. Read-only discovery found no root .gitignore or AGENTS.md; project-local ignore rules were used. The earlier diary records a whitespace rejection around quoted blank lines before the commit, but no verbatim diagnostic was retained; this backfill does not fabricate one.

### What I learned

A model event should describe an observable complete transition, not every future RTL state. Entering a pending output is different from accepting it. Full checkpoint comparison can be stronger than comparing only domain masks.

### What was tricky to build

The generator normally advances one event per step. Without a separate waiting flag, a blocked consumer could accidentally advance the generator into output acceptance. The implementation checks waiting and result_ready before resuming. Choice.domain snapshots remain present only as a model assertion aid.

### What warrants a second pair of eyes

Review Machine.step, _write, frame, and the oracle separately. Confirm that the independent oracle does not import the propagation implementation and that first-only cut follows acceptance.

### What should be done in the future

Use this corrected commit attribution when reviewing the implementation. No new implementation work is implied by this retrospective entry.

### Code review instructions

Historical command: python3 -m pytest sim/test_model.py -q, from queens_rollback. Inspect git show cf52672 -- queens_rollback/tools queens_rollback/sim/test_model.py.

### Technical details

The plan receipt is 2026-09-04T21:36:06Z; P1 start is 21:36:09Z and completion 21:39:36Z. The P1 completion YAML contains literal HEAD in its QR URL. At that time the commit was cf52672, but the printed URL is mutable; the original receipt is preserved.

## Step 9: Retrospective P2: snapshot storage and event publication

This entry expands Step 3 using the snapshot implementation commit. The original Step 3 Commit field refers to the preceding model commit cf52672, which was the available milestone while the step was written. The actual snapshot code milestone is 0f15376; this annotation corrects attribution while preserving the original chronological text.

### Prompt Context

**User prompt (verbatim):** (see Step 8)

**Assistant interpretation:** Expand the historical record from committed implementation and retained evidence; distinguish retrospective reconstruction from contemporaneous observations.

**Inferred user intent:** Preserve enough detail to understand decisions, reproduce validation, and review the finished solver.

**Commit (code):** 0f15376 — feat: implement synchronous queens snapshot RTL baseline (2026-09-04T17:43:32-04:00)

### What I did

- Implemented fixed-width mask helpers and a single owner for domain and checkpoint state.
- Stored eight 104-bit checkpoint records in the shared synchronous RAM wrapper. The low 40 bits carry metadata, and the upper 64 bits carry domains.
- Separated checkpoint preparation, inactive-slot write, top publication, registered RAM wait, capture, and restoration.
- Added a testbench that emits complete live checkpoint records and a Python harness comparing those records with the model.

### Why

A full snapshot gives a simple recovery baseline. Using the same synchronous RAM discipline planned for trail mode avoids comparing a combinational-memory prototype against a clocked-memory implementation.

### What worked

Step 3 records 23 passing combined tests: 16 model cases and seven snapshot RTL cases. The four full/first and ready/stalled combinations plus choice capacities 0, 1, and 5 exercise the baseline behavior.

### What didn't work

No RTL compile or simulation failure is recorded. The whitespace issue from P1 was fixed in the diary helper. Subsequent print helpers resolve commit identifiers before creating QR references.

### What I learned

Publishing a new checkpoint top may follow an inactive RAM write. Updating remaining alternatives in an already live record must publish UPDATE on the actual write edge, or a state monitor sees an unannounced mutation.

### What was tricky to build

The shared RAM read output and controller registers both use nonblocking assignments. A controller cannot consume a newly registered read on the same edge without accounting for that scheduling. Explicit wait/capture states establish ownership of the consumed word.

### What warrants a second pair of eyes

Review S_CP_WRITE versus S_CP_PUBLISH and S_UPDATE_WRITE. Check domain packing direction: column zero occupies the least significant domain byte. Verify that snapshot restoration installs domains and propagated bits at the same semantic boundary.

### What should be done in the future

Use this corrected commit attribution when reviewing the implementation. No new implementation work is implied by this retrospective entry.

### Code review instructions

Historical command: source /home/manuel/fpga/oss-cad-suite/environment, then python3 -m pytest sim -q. Inspect git show 0f15376 and sim/test_rtl.py full C-record comparison.

### Technical details

P2 start printed at 21:39:39Z; completion at 21:43:38Z. The completion QR points to 0f1537618c78029d0b90f1990d236488a7284f19. The implementation directly reuses symbolic_eval/rtl/sync_sdp_ram.sv without modifying it.

## Step 10: Retrospective P3: log-before-write and reverse restoration

This entry expands Step 4 and attributes the trail implementation to 214a39a. Step 4 originally listed the preceding snapshot milestone 0f15376. The source change added recovery storage and sequencing while preserving the existing propagation algorithm, so measurements compare storage policies under the same search order.

### Prompt Context

**User prompt (verbatim):** (see Step 8)

**Assistant interpretation:** Expand the historical record from committed implementation and retained evidence; distinguish retrospective reconstruction from contemporaneous observations.

**Inferred user intent:** Preserve enough detail to understand decisions, reproduce validation, and review the finished solver.

**Commit (code):** 214a39a — feat: add synchronous mutation trail and reverse queens rollback (2026-09-04T17:47:03-04:00)

### What I did

- Selected a 40-bit checkpoint RAM in trail mode and introduced 64 physical 20-bit trail slots.
- Added S_LOG_WRITE before S_APPLY, with top publication on the domain-update edge.
- Implemented mark checks, reverse synchronous reads, one-domain restoration, and final propagated-bitmap restoration.
- Added logical capacities, cut-base publication, and checked backend profiling. Full-record request metrics later clarified the distinction between logical history counters and actual record widths.

### Why

An old domain must be retained before a new domain becomes visible. A checkpoint mark identifies which suffix of the trail belongs to the current alternative; reverse application restores repeated writes to the same column correctly.

### What worked

Step 4 records 36 passing tests. Unstalled enumeration measured 53,951 snapshot cycles and 74,523 trail cycles with 3,980 changed domain writes in each. Full outputs and live records matched the model.

### What didn't work

No P3 compile or simulation failures are recorded. The measurements contradicted the informal expectation that a trail would necessarily reduce traffic; that was an experimental finding, not a failing test.

### What I learned

The live trail high-water mark is 32 and the choice high-water mark is six. A smaller checkpoint width can still lead to greater aggregate traffic because many mutations occur after a single choice. No-change writes must bypass logging, while zero-domain writes must remain reversible.

### What was tricky to build

A physically written trail slot is not live until the top advances. During reverse recovery, intermediate domains do not yet constitute a complete checkpoint; the saved propagated bitmap is installed only at RESTORED. The model asserts the complete checkpoint at that boundary.

### What warrants a second pair of eyes

Check trail[19:17] column, [16:9] old domain, [8:4] choice level, and [3:0] zero flags. Review base <= mark <= top and capacity checks before consuming a retry alternative.

### What should be done in the future

Use this corrected commit attribution when reviewing the implementation. No new implementation work is implied by this retrospective entry.

### Code review instructions

Historical command: python3 -m pytest sim -q under the CAD environment. Reproduction: ticket scripts/04-profile-backends.py; inspect backend-profile.json and source commit 214a39a.

### Technical details

P3 start printed 21:43:41Z and completion 21:47:07Z. The completion QR points to 214a39a47d5207654330b70edc2d4ddfdc51ac34. Complete-record traffic is 672*104 bits for snapshots versus 672*40+3980*20 bits for the trail.

## Step 11: Retrospective P4: boundaries between events, reset, and corruption

This entry expands Step 5 and identifies 1a5194f as the boundary-verification milestone. Its original Commit field named the preceding trail commit. The additional checks examine what can happen between correct-looking event snapshots, which is where premature publication and reset ownership mistakes can otherwise remain hidden.

### Prompt Context

**User prompt (verbatim):** (see Step 8)

**Assistant interpretation:** Expand the historical record from committed implementation and retained evidence; distinguish retrospective reconstruction from contemporaneous observations.

**Inferred user intent:** Preserve enough detail to understand decisions, reproduce validation, and review the finished solver.

**Commit (code):** 1a5194f — test: verify queens rollback stability reset and cut boundaries (2026-09-04T17:50:19-04:00)

### What I did

- Added every-cycle monitors for domains, propagated bitmap, live pointers, output count, and all live RAM records when no semantic event is emitted or a fault occurs.
- Added reset interruption during log writes, reverse application, pending output, and checkpoint writes; comparisons restart after a RESTART marker against a fresh model.
- Injected invalid checkpoint marks, non-singleton propagation sources, and nonzero trail flags.
- Added seeded readiness patterns, a 200-cycle first-result stall, and 64-cycle terminal dwell checks.

### Why

An event-only checker could miss a mutation that appears early and is repaired before the next event. Reset tests need to verify reachability rather than physical RAM clearing. Terminal dwell checks establish that faults and cut cannot resume execution.

### What worked

The contemporaneous record reports 34 RTL tests passing in 29.38 seconds. The final complete suite later includes these cases. Corrupt sources report BAD_ONEHOT; malformed marks and trail flags report TRAIL_INTEGRITY.

### What didn't work

No P4 failures or unsuccessful repair attempts are recorded. Deliberate corruption is test stimulus, not evidence of an implementation failure.

### What I learned

The testbench itself changes state during corruption injection, so its stability monitor needs a narrowly bounded exemption on that injection cycle. Normal protection checks resume immediately afterward.

### What was tricky to build

Reset leaves stale RAM data physically present. Zeroed tops make it unreachable, and subsequent writes establish new ownership. Tests must compare only live entries and a fresh post-reset run, or they would incorrectly require RAM erasure.

### What warrants a second pair of eyes

Inspect injection exemptions and terminal monitors in tb_queens.sv. Confirm that counters allowed to run during stalls are not confused with protected search state. Integrity checks detect structural inconsistencies but are not a general memory-corruption checksum.

### What should be done in the future

Use this corrected commit attribution when reviewing the implementation. No new implementation work is implied by this retrospective entry.

### Code review instructions

Historical command: python3 -m pytest sim/test_rtl.py -q -x under the CAD environment. Inspect git show 1a5194f and parameterized restart/inject cases.

### Technical details

P4 start printed 21:47:11Z and completion 21:50:24Z. Completion references 1a5194f362458600afdcd82ab16ca01e34caf264. BAD_DOMAIN_INDEX remains a reserved hardware fault because internal column fields cannot encode an index above seven.

## Step 12: Retrospective P5: serial ownership and physical measurement provenance

This entry expands Steps 6 and 7 from the board integration commit, hardware scripts, and archived logs. The implementation commit is ca666fd; the results package is 6c7b2d3 and the final print receipt is 4345ee4. The physical experiment verifies external records, while internal rollback trace claims remain grounded in simulation.

### Prompt Context

**User prompt (verbatim):** (see Step 8)

**Assistant interpretation:** Expand the historical record from committed implementation and retained evidence; distinguish retrospective reconstruction from contemporaneous observations.

**Inferred user intent:** Preserve enough detail to understand decisions, reproduce validation, and review the finished solver.

**Commit (code):** ca666fd — feat: integrate queens solver with board UART and terminal records (2026-09-04T17:56:40-04:00)

### What I did

- Added the one-record formatter, board reset integration, and UART waveform tests. The formatter retains data through the final byte and emits a terminal D or F record exactly once.
- Ran all 58 tests, then built snapshot enumeration, trail enumeration, and trail-first with router2 seed two and the same 10 MHz constraint.
- Armed /dev/ttyACM0 capture before each openFPGALoader invocation, observed eight seconds, and compared the complete bytes with model-generated records.
- Archived image hashes, builds, route reports, loader output, UART binaries, profiles, and simulation first-result traces.

### Why

The previous laboratory exposed a synthesis-only problem on hardware. Physical UART verification is therefore a separate evidence layer. Explicit terminal records distinguish completion from silence, and capture-before-load preserves startup output.

### What worked

All eight board tests passed in 3.83s; the complete suite passed 58 tests in 51.71s. Physical snapshot and trail streams each contained 932 matching bytes, and trail-first contained 22. Routed estimates were 30.19, 30.98, and 33.72 MHz, each passing the actual 10 MHz constraint.

### What didn't work

No synthesis, routing, loading, or physical-byte mismatch occurred. A premature status read saw no P5-hardware.exit file while work was still running; it was subsequently written as 0. P5-tests.log explicitly records the observed tool result rather than pretending to be a raw second test run.

### What I learned

Core acceptance is not UART completion. terminal_sent marks terminal-record selection, not final-bit delivery. nextpnr RAM_HALF and CPE subresource counts must retain their units. Wider snapshots map to more RAM halves even though their logical bit allocation is smaller.

### What was tricky to build

The serial reader must start before configuration and must not compete with another process for the UART. The benchmark excludes output stalls, whereas physical streams are UART-limited; an eight-second observation window cannot be reported as solve time.

### What warrants a second pair of eyes

Review the LAST formatter state, reset during transmission, P5-build-commit.txt, image hashes, final rather than pre-route Fmax, and expected-versus-actual bytes. Fault and internal reset injection were simulated; they were not among the three physical images.

### What should be done in the future

Use this corrected commit attribution when reviewing the implementation. No new implementation work is implied by this retrospective entry.

### Code review instructions

Inspect scripts/05-hardware.py and 06-hardware.sh. Reproduce in tmux with the CAD environment sourced and exclusive physical UART access. Read P5-hardware.json and P5-measurements.json before interpreting performance.

### Technical details

P5 start printed 21:50:27Z and completion 22:01:44Z. The final completion QR references 6c7b2d37099830ff979416674bad0d9ffeed25d2. All eleven receipts report printed:true. The final loaded image was trail-first; repository main was pushed through 4345ee4.

## Step 13: Textbook report, detailed backfill, and vault delivery

Wrote a 5,295-word technical article about the implemented rollback solver and delivered it as a new dated note in go-go-parc. The article teaches domain propagation, complete checkpoint state, synchronous memory publication, reverse restoration, fault precision, output acceptance, and the measured resource/cycle tradeoff. It emphasizes system behavior rather than implementation chronology.

The user also requested detailed diary backfill. Steps 8–12 preserve the original entries while adding retrospective implementation explanations, exact milestone attribution, print chronology, test provenance, and review instructions. They explicitly distinguish reconstructed detail from contemporaneous observations; they do not invent missing diagnostics or new test runs.

### Prompt Context

**User prompt (verbatim, JSON-encoded to preserve whitespace):**

```json
"write a detailed project report for the obsidian vault as a deep dive technical analysis blog post using a textbook writing style (no analogies, see skill).      \n Commit and push the bsidian vault when done (go-go-parc vault).   \n\n^-- for the go-go-parc vault."
```

**Additional user prompt (verbatim):** (see Step 8)

**Assistant interpretation:** Write and publish the system-focused textbook report in the explicitly named vault, and make the implementation diary detailed enough to review and reproduce the completed work.

**Inferred user intent:** Preserve both a durable technical explanation and a trustworthy engineering record.

**Commit (analyzed implementation):** 4345ee4a9eb3460d6e94646d77e08dffa91489a0.

**Commit (vault report):** 8a777e77a7f486e12d3fbf394e7e4d210dae01a9 — pushed to go-go-parc origin/main.

### What I did

- Read the vault-writing, textbook-authoring, and diary skills and inspected the existing tagged-stack CPU article to match vault metadata and technical depth.
- Read queens_core, queens_model, the mask helpers, formatter, board top, shared RAM/UART, test harness, profile metrics, first-result traces, and hardware evidence.
- Created the report in reference/02-inside-the-eight-queens-rollback-engine-technical-project-report.md and added scripts/09-report-examples.py to validate every selected contradiction/retry frame against the archived RTL trace and the model.
- Used scripts/10-backfill-diary.py for the retrospective entries and scripts/11-export-vault-report.py to validate and create a new vault note without replacing any historical note.
- Committed only the new vault article and pushed it. The pre-existing untracked AgentForum research note was left untouched.

### Why

The report needs to explain why the machine works, including the exact state that must be restored and the difference between a core handshake and a completed UART record. The diary needs to retain provenance so a later reader can identify the implementing commit rather than mistaking a previously available milestone for the change itself.

### What worked

The report example script passed and reproduced first board 672BE0, the first contradiction, restoration to mark 15, retry from F0 to 20, and the cut with top/base 29. Export checks verified more than 5,000 body words, balanced code fences, no template placeholders, and required example/metric anchors. The delivered vault body matches the ticket body. Vault HEAD and origin/main both identify the report commit. Docmgr doctor passed for the ticket.

### What didn't work

No model-example, export, Git, or push failure occurred. A read command accidentally included head -60 ' /dev/null'; it reported "head: cannot open ' /dev/null' for reading: No such file or directory" after the intended source reads had succeeded. This was a harmless discovery-command typo, not a software failure. No hardware or RTL changes were needed for the report, so the full 58-test suite was not rerun.

### What I learned

The implementation's terminal_sent name describes selection of the terminal record, not completed physical transmission. The UART divider is 87 at 10 MHz, so 932 bytes require at least 81.084 milliseconds of serial bit time; the eight-second capture window is an observation interval, not solve time. The report states those distinctions explicitly. The trail allocates more logical record bits than snapshots despite using fewer mapped RAM_HALF resources.

### What was tricky to build

The original diary's Steps 3–5 named the preceding milestone because entries were written before their code commits. The backfill preserves those historical entries and records the actual mapping: snapshots 0f15376, trail 214a39a, boundary checks 1a5194f. Raw trace words pack column zero into the least significant byte, so example decoding must follow that order. Physical evidence was restricted to what the UART can observe; internal trace and fault/reset claims are labeled as simulation evidence.

### What warrants a second pair of eyes

Review the record-bit calculations, the distinction between RAM_HALF and whole physical blocks, the serial-time lower bound, and the boundary where propagated bits are restored. Check that the article's fixed source revision and source line references remain appropriate if later implementation work is added.

### What should be done in the future

Use a new dated follow-up note for future solver extensions or new measurements. Do not silently rewrite this report's measured snapshot. No additional implementation or publication work is required for this request.

### Code review instructions

Run ticket scripts/09-report-examples.py and inspect reference/validation/report-examples.json. Compare the report body with the vault note and the stored report-vault-delivery.json hash. Inspect the vault commit to confirm it contains only the intended article. Read retrospective Steps 8–12 alongside the implementing commits and preserved print receipts.

### Technical details

Vault path: Projects/2026/09/04/ARTICLE - GateMate Symbolic - Inside an FPGA Rollback Solver.md. The note uses article frontmatter, native Mermaid diagrams, mathematical definitions, pseudocode, trace excerpts, resource/API tables, immutable source links, and two existing vault wikilinks. No external sources were downloaded, and all scripts written during this work are in the ticket scripts directory.

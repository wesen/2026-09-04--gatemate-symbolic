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

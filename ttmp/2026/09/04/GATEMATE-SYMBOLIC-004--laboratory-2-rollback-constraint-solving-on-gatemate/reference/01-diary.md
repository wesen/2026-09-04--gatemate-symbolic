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
RelatedFiles: []
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

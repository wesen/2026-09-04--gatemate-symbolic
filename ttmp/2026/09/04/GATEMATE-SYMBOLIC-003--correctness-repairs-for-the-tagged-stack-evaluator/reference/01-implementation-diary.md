---
Title: Implementation diary
Ticket: GATEMATE-SYMBOLIC-003
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
Summary: ""
LastUpdated: 2026-09-04T15:55:42.010642239-04:00
WhatFor: ""
WhenToUse: ""
---

# Diary

## Goal

Implement and validate the review findings, preserving commands, decisions, commits, and printed phase receipts.

## Step 1: Repair design and printed plan

Created a new correctness ticket from the implementation review and wrote the repair design before changing production code. The plan has five phases covering bounded bugs, commitment, ROM boundaries, verification, and full build/hardware evidence.

### Prompt Context

**User prompt (verbatim):**

> Create a new ticket, add design doc on how to fix things, then build them. commit at appropriate intervals and keep a detailed diary as you work (using the diary format from the skill)
> 
> Print out a brutalist work slip with the plan / different phases for the ticket. then before stsarting a phase, plrint a split about the phase, and print one when the phase is done.

**Assistant interpretation:** Implement the repair design in phases with reproducible validation, focused commits, detailed diary entries, and printed phase boundaries.

**Inferred user intent:** Turn the documented review findings into a reliable working evaluator with visible progress and reviewable evidence.

### What I did

Read docmgr, diary, and brutalist-work-slip skills; inspected both core interfaces, assembler, and simulation runners. Created ticket GATEMATE-SYMBOLIC-003 and its design, tasks, diary helper, and print helper. Requested the overall plan print using the configured service.

### Why

Keep the reviewed baseline separate from implementation changes and make the semantics of PC escape and fault observation explicit before touching interfaces.

### What worked

Ticket creation and source inspection succeeded. The existing review supplies minimized counterexamples for regression tests.

### What didn't work

The initial discovery command used rg --files -g AGENTS.md with an && chain; no matching file returned exit 1, so its trailing read did not execute. Subsequent direct source reads completed. No software repair failures yet.

### What I learned

The fix must widen every architectural address field while leaving physical ROM addresses narrow. FETCH is an event context, not an ISA opcode.

### What was tricky to build

The printer helper delegates layout generation to the installed work-slip script; no handwritten YAML. Phase receipts and layouts are archived under reference/slips.

### What warrants a second pair of eyes

Review the explicit trace_fetch/fault_fetch interface choice and generalized BRAM context read.

### What should be done in the future

Execute P1 through P5 in order, printing before and after each phase.

### Code review instructions

Read the design and tasks; inspect reference/slips/PLAN-start.log for the real print result.

### Technical details

Baseline production code c7f9dc9; latest review commit 4336966. All ad hoc scripts are kept in this ticket scripts directory.

## Step 2: P1 instruction and assembler guards

Added minimized regression programs and confirmed twelve failures against the original implementation. Implemented the guards and reran the complete suite successfully: 145 tests pass.

### Prompt Context

**User prompt (verbatim):**

(see Step 1)

**Assistant interpretation:** Implement the repair design in phases with reproducible validation, focused commits, detailed diary entries, and printed phase boundaries.

**Inferred user intent:** Turn the documented review findings into a reliable working evaluator with visible progress and reviewable evidence.

**Commit (code/work):** a3b50fb — fix: enforce stack guards and reject invalid assembly

### What I did

Added sim/test_repairs.py with both-core stack cases and parser/capacity boundaries. Initialized register-core retirement depth from current state; rejected empty BRAM DUP; made invalid source an explicit syntax error and checked capacity before every emission path.

### Why

These bugs either expose stale values as live operands or silently alter assembled programs.

### What worked

P1-before.log: 12 failed, 10 passed in 0.47s. P1-after.log: 145 passed in 6.48s. PLAN and P1 start slips returned printed: true.

### What didn't work

Expected pre-fix regression assertions failed as recorded in P1-before.log. Automatic approval initially rejected the done slip, incorrectly citing the pre-fix failures as the post-fix result; a retry supplies the distinct P1-after.log evidence. No failed implementation repair iteration occurred.

### What I learned

A candidate initialized from current depth avoids stale state after retirement outside COMMIT. WORD must pass the same capacity gate as every other emitting instruction.

### What was tricky to build

DUP on empty physical storage still maintains the count invariant after manufacturing a value; the test needs architectural semantics rather than occupancy alone.

### What warrants a second pair of eyes

Check syntax diagnostics and empty-stack faults after DROP and EMIT, not only immediately after reset.

### What should be done in the future

P2 will stage return depth and acquire true BRAM fault context.

### Code review instructions

Run make test in symbolic_eval with the OSS CAD Suite environment; inspect sim/test_repairs.py and P1-before/after logs.

### Technical details

The assembler now supports ROM depth 2..32768 explicitly; all current assembly clients fit this range. No compatibility mode added.

## Step 3: P2 atomic return retirement and deep fault context

Added return-depth timing assertions and five BRAM fault-context regressions. All seven new failing cases now pass, and the complete suite reaches 157 passing tests.

### Prompt Context

**User prompt (verbatim):**

(see Step 1)

**Assistant interpretation:** Implement the repair design in phases with reproducible validation, focused commits, detailed diary entries, and printed phase boundaries.

**Inferred user intent:** Turn the documented review findings into a reliable working evaluator with visible progress and reviewable evidence.

**Commit (code/work):** 3213364 — fix: retire return state atomically and fetch deep fault context

### What I did

Added nrdepth staging in both cores, committed it alongside PC and return writes, and generalized the BRAM tc=1/dc>0 operand read to every instruction. Reused the same response for pop refill and removed the redundant R_POP read purpose.

### Why

Return occupancy must not expose unwritten continuations; fault tags must represent logical operands rather than stale cache registers.

### What worked

P2-before.log: 7 failed, 27 passed in 0.85s. P2-after.log: 157 passed in 6.98s. P1 completion and P2 start slips printed successfully.

### What didn't work

Expected red regressions include ASSERT_FAIL: return depth changed without retirement. Approval review again initially treated the post-fix print result as unverified despite the successful captured log; retried with the completed tool-call evidence. No implementation test failures remained after the first repair.

### What I learned

One context read can serve both fault metadata and pop refill; the extra cycle for otherwise neutral operations makes the fault contract consistent.

### What was tricky to build

nrdepth_d is initialized in EXECUTE but rdepth_d is changed only in COMMIT. A combinational next-state name alone does not imply delayed architectural publication.

### What warrants a second pair of eyes

Inspect CALL/RET candidate ownership and fault tags for BAD_OPCODE, invalid JMP/CALL, empty RET, and type-invalid JZ with a Boolean second value in RAM.

### What should be done in the future

P3 handles one-past-ROM PC and return addresses with explicit fetch flags.

### Code review instructions

Run make test; inspect both bench return-depth monitors and P2-before/after logs.

### Technical details

Actual output record semantics and instruction encodings remain unchanged. RAM read timing grows by one cycle only when context was previously not prefetched.

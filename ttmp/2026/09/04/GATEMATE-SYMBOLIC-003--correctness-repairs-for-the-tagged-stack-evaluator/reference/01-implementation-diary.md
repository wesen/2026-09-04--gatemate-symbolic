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

## Step 4: P3 one-past-ROM and fetch faults

Preserved the model semantics for sequential execution and CALL at the ROM boundary. Both RTL cores now hold one-past-ROM architectural addresses and publish an explicit fetch-fault context rather than wrapping.

### Prompt Context

**User prompt (verbatim):**

(see Step 1)

**Assistant interpretation:** Implement the repair design in phases with reproducible validation, focused commits, detailed diary entries, and printed phase boundaries.

**Inferred user intent:** Turn the documented review findings into a reliable working evaluator with visible progress and reviewable evidence.

### What I did

Widened architectural PC, candidate PC, return addresses, fault/trace/debug PCs; retained narrow guarded ROM addresses. Added trace_fetch/fault_fetch flags, BRAM fetch-context states, integer target comparisons, parameter checks, and ROM-size regressions.

### Why

The model permits the final valid instruction to retire; a later fetch at ROM_DEPTH faults. Return addresses must preserve the same boundary.

### What worked

175 tests pass in 10.44s. Tests cover ROM sizes 2, 7, 16, 1024 and maximum target 32767; last-word HALT, EMIT, CALL/RET, and deep-stack fault tags agree with the model.

### What didn't work

The checked edit script initially raised AssertionError because its ROM-address regex matched a comment as well as code; it wrote no RTL before aborting. Narrowed the regex to the assignment line. The command chain nevertheless ran pre-fix tests: 16 failed, 159 passed in 21.05s, preserved as P3-patch-not-applied.log. The successfully applied repair passed on its first validation run.

### What I learned

ROM_DEPTH[14:0] becomes zero at 32768; integer comparison is required even after PC widths are corrected.

### What was tricky to build

BRAM fetch faults need a context read after a last-word pop leaves tc=1 and dc>0; S_FETCH_CONTEXT captures that value before S_FETCH_FAULT publishes metadata.

### What warrants a second pair of eyes

Verify every architectural address consumer and the meaning of trace_fetch versus fault_fetch. Physical ROM addresses are clamped while the architectural PC is invalid.

### What should be done in the future

P4 adds full-state comparison and improves generated programs.

### Code review instructions

Run make test; inspect test_onepast_push, test_last_word_cases, and test_largest_rom_last_target.

### Technical details

Fetch is not assigned an ISA opcode. Trace format uses the explicit flag to print FETCH, matching the existing Python model. Unsupported core memory dimensions fail elaboration.

## Step 5: P4 complete architectural verification

Both core harnesses now compare every live 40-bit data value, every return address, all trace/fault fields, and accepted output words against the model. Added per-cycle stability assertions, model-validated random fragments, production BRAM capacity tests, constructed states, and restart-under-load board simulation.

### Prompt Context

**User prompt (verbatim):**

(see Step 1)

**Assistant interpretation:** Implement the repair design in phases with reproducible validation, focused commits, detailed diary entries, and printed phase boundaries.

**Inferred user intent:** Turn the documented review findings into a reliable working evaluator with visible progress and reviewable evidence.

**Commit (code/work):** P3 c9805a6; P4 commit recorded in changelog after validation

### What I did

Added state_checks.py, program_generation.py, test_verification.py, and scripts/check_isa.py. Instrumented both benches for complete snapshots and architectural stability. Added simulation-only initial-stack injection, full flag transfer checks, signed overflow/noncanonical cases, and button resets during compute and UART transmission.

### Why

Trace depth alone cannot prove values or return addresses correct. Legal random generation must execute intended instructions rather than terminate accidentally in padding or type faults.

### What worked

P4-state: 175 passed. Generator and capacity: 183 passed. Constructed-state focused suite: 20 passed. Reset suite: 8 passed. Final complete suite: 197 passed in 18.80s. Coverage counters are archived in P4-coverage.json.

### What didn't work

The reset-test edit script initially raised AssertionError because the expected loop used j++ while the source uses j = j + 1. The guard stopped before writing files. Corrected the exact pattern; reset tests passed on the first applied version. No RTL correctness failure was discovered by the strengthened checker.

### What I learned

The original fixes withstand full live-state comparison, including a 514-value stack and nonzero flags that the UART text omits. All 17 opcodes and both JZ outcomes occur in the legal generator coverage test.

### What was tricky to build

Snapshot order must be RAM oldest-first then live top1/top0, not always two cached entries. The UART decoder must cancel a partially sampled byte on reset and resume only after reset releases.

### What warrants a second pair of eyes

Check snapshot extraction independently of model semantics, and verify that no invalid physical stack slots are compared. Initial-stack injection occurs only in simulation after reset and before fetch.

### What should be done in the future

P5 will rebuild the actual board image, update docs, and attempt hardware evidence if the board becomes available.

### Code review instructions

Run make test and inspect build/verification-coverage.json; metadata drift is checked by test_isa_mirrors.

### Technical details

A user-input request asked for the board to be connected because no ttyACM devices were present. Simulation-only instrumentation consumes no hardware BRAM. Test count is 197.

## Step 6: P5 hardware discrepancy before completion

Synthesis, routing, and packing succeeded, and three board programs matched the model. Countdown exposed a new hardware-only discrepancy: the final Boolean true is transmitted with INT tag zero while its payload is still one. P5 is not yet complete.

### Prompt Context

**User prompt (verbatim):**

(see Step 1)

**Assistant interpretation:** Implement the repair design in phases with reproducible validation, focused commits, detailed diary entries, and printed phase boundaries.

**Inferred user intent:** Turn the documented review findings into a reliable working evaluator with visible progress and reviewable evidence.

**Commit (code/work):** 49807a4 — docs: align repaired machine contracts and pin router seed

### What I did

Built with seed 2, captured UART before loading each image, and compared actual bytes with model output. Preserved image hashes, serial bytes, loader logs, and P5-hardware.json.

### Why

Board verification is required to distinguish behavioral simulation from implemented FPGA behavior.

### What worked

Fibonacci emitted T0:00000037; arithmetic emitted T1:00000001; typefault emitted no bytes. Routing seed 2 passed 10 MHz at 15.52 MHz in 72 iterations. All 197 simulations passed.

### What didn't work

Countdown assertion: expected final T1:00000001 but actual final T0:00000001. First five integer lines match. P5-hardware.exit is 1 and the failed image hash is 529d4ea26584df5833f25c60a89a835145a09bd0b326ee62b79de43283f6c253. Initial lsusb inside the sandbox returned unable to initialize libusb: -99; outside-sandbox discovery found the board. The original routing run completed before cancellation, so tmux send-keys returned cannot find pane rather than stopping a running process.

### What I learned

Full behavioral simulation does not by itself validate memory inference and technology mapping. A selective tag discrepancy needs a bounded reproduction before deciding which implementation layer to change.

### What was tricky to build

The first routing seed appeared stuck on one resource conflict but eventually converged at 15.60 MHz. Seed 2 completed much faster. The board was present but hidden by sandbox device restrictions.

### What warrants a second pair of eyes

Do not mark the final phase complete while countdown differs. UART silence for typefault still does not directly expose fault metadata.

### What should be done in the future

Reproduce the countdown capture and run the production-depth configuration; allow at most two repair attempts under the user debugging rule.

### Code review instructions

Inspect P5-hardware.json and P5-countdown-uart.bin; compare the last tag against programs/countdown.asm.

### Technical details

No repair for this new hardware discrepancy has been attempted yet. Final completion slip has not been printed.

## Step 7: P5 isolate the persistent board mismatch

Repeated the same countdown image on hardware and reproduced the final INT tag. The production-depth behavioral simulation emits the correct final BOOL tag, so the discrepancy is not explained by the earlier 30-versus-512 BRAM test configuration.

### Prompt Context

**User prompt (verbatim):**

(see Step 1)

**Assistant interpretation:** Implement the repair design in phases with reproducible validation, focused commits, detailed diary entries, and printed phase boundaries.

**Inferred user intent:** Turn the documented review findings into a reliable working evaluator with visible progress and reviewable evidence.

### What I did

Added scripts/14-repeat-capture.sh and retained the repeated UART bytes. Exported the mapped Yosys netlist with scripts/15-mapped-sim.ys and generated production-depth and mapped simulation benches with scripts/16-board-sim.py.

### Why

A repeat capture and a production-depth simulation distinguish a persistent implementation discrepancy from a one-off serial error or parameter mismatch before any repair attempt.

### What worked

Repeated hardware output again ends T0:00000001. Production-depth simulation ends T1:00000001 with all earlier lines matching.

### What didn't work

Mapped-netlist simulation is still running and has not produced a conclusion. No source fix has been attempted for the board discrepancy.

### What I learned

The reported tag error is reproducible with the same bitstream hash. The behavioral checker passes even at the production stack depth.

### What was tricky to build

The mapped top has no parameters or hierarchical ROM array; its testbench uses the initialized technology-mapped ROM and the existing top-level debug aliases.

### What warrants a second pair of eyes

Keep the behavioral-versus-mapped-versus-board distinction explicit; do not infer a compiler bug until the mapped simulation or signal cone demonstrates it.

### What should be done in the future

Use mapped or intermediate synthesis observation to localize the fault; obey the two-repair-attempt limit.

### Code review instructions

Compare P5-countdown-repeat.bin, P5-production-simulation.log, and the mapped simulation result when available.

### Technical details

The user reiterated: commit at appropriate intervals and keep a detailed diary as you work (using the diary format from the skill). This investigation evidence is committed before attempting a new implementation repair.

## Step 8: P5 repair constant constructors after synthesis reproducer

The first targeted repair succeeded: assigning complete packed values preserves tags and flags through Yosys constant folding. The minimal synthesized regression and all 198 tests pass; board reruns remain pending.

### Prompt Context

**User prompt (verbatim):**

(see Step 1)

**Assistant interpretation:** Implement the repair design in phases with reproducible validation, focused commits, detailed diary entries, and printed phase boundaries.

**Inferred user intent:** Turn the documented review findings into a reliable working evaluator with visible progress and reviewable evidence.

### What I did

Added constructor_probe.sv, tb_constructor.sv and test_synthesis.py. Changed mk_int and mk_bool to packed concatenation assignments. Archived all initial hardware evidence before rebuilding and marked raw serial captures as binary in Git.

### Why

Countdown emitted INT 1 instead of BOOL 1 on the board despite passing RTL simulation. Synthesized netlist simulation reproduced it, isolating the error from serial capture and board transport.

### What worked

Generic lowering showed 40-hxx00000001 for constant true. The isolated regression failed with true=xx00000001 false=xx00000000 zero=xx00000000. Packed assignment passes that regression and the full suite: 198 passed in 18.35s.

### What didn't work

Before repair: FATAL: sim/tb_constructor.sv:11: constructor tags/flags lost: true=xx00000001 false=xx00000000 zero=xx00000000 int=00ffffffff bool=1000000001. No unsuccessful source repair attempts occurred.

### What I learned

Constant and dynamic arguments take distinct frontend paths; only testing dynamic arithmetic results masked constant constructor corruption.

### What was tricky to build

RTL field assignments were semantically valid, but their constant-folded struct temporary lost upper bits in this Yosys frontend. Whole packed assignments remove that lowering ambiguity without changing the value format.

### What warrants a second pair of eyes

Review both signed integer bit preservation and Boolean canonicalization. The regression executes the emitted generic netlist, rather than merely examining RTL text.

### What should be done in the future

Rebuild and repeat all four board captures, then publish final validation and print the P5 completion slip.

### Code review instructions

Source the documented OSS CAD Suite environment; run python3 -m pytest sim/test_synthesis.py -q, then python3 -m pytest -q in symbolic_eval. Read P5-constructor-before.log and P5-final-tests.log.

### Technical details

Initial failing image and UART results remain under reference/validation/before-constructor-fix. User instruction remains: commit at appropriate intervals and keep a detailed diary as you work (using the diary format from the skill).

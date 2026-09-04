---
Title: Correctness repair design and validation plan
Ticket: GATEMATE-SYMBOLIC-003
Status: active
Topics: [fpga, gatemate, symbolic-computers, architecture]
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Implement the findings from GATEMATE-SYMBOLIC-002 with precise retirement and stronger verification.
LastUpdated: 2026-09-04T16:00:00-04:00
WhatFor: Guide implementation and review of the demonstrated correctness defects.
WhenToUse: Implementing, validating, or extending the repaired evaluator.
---

# Correctness repair design

## Objective and scope

Repair review findings F1–F9 while preserving the Python machine's specified instruction semantics. The starting implementation is unchanged from `c7f9dc9`; review-only commits culminate in `4336966`. This ticket changes the assembler, two RTL cores, simulation observation, random generation, and accompanying documentation. No compatibility layer is needed or planned.

The primary acceptance condition is architectural agreement: after every retirement, both logical stacks, the PC, output stream, halted state, and fault state must match the executable model. Internal memory reads may take extra clocks but must not change visible state. A green demonstration alone is insufficient: the review's minimized defects must become regressions.

## P1: instruction and assembler guards

The BRAM PUSH/DUP case must reject DUP when depth is zero before using top0. Register-core EXECUTE must initialize its retirement candidate from current PC/depth, so JMP and SWAP cannot reuse an old depth after EMIT. Preserve instruction-specific overrides.

The assembler must explicitly distinguish blank/comment lines from syntax errors. Invalid nonempty source raises ValueError with line context. Capacity is checked on the common instruction-emission path before either WORD, immediate placeholders, or ordinary instructions append a word. Validate positive ROM size and correct stale tuple annotations/comments.

Regression inputs include empty DUP, DUP after consuming the last value, DUP at capacity, EMIT followed by JMP/SWAP, malformed multi-operand lines, and raw/mixed instructions exceeding capacity. Establish failing tests before fixes, then run the complete suite.

## P2: atomic return state and correct fault context

Introduce `nrdepth_q/d` as staged return depth. EXECUTE initializes it from rdepth and CALL/RET override the candidate. Only COMMIT assigns the architectural rdepth from the candidate, together with the return-address array write and PC update. A per-cycle monitor must detect changes in either live stack outside retirement; the return stack cannot publish an address as live before it has been written.

The BRAM machine may have tc=1 and dc>0. In this state the second logical value is RAM[dc-1], while opnd_q may contain stale data. Generalize decode operand-context acquisition to fetch this value for every instruction when needed, not only arithmetic/SWAP. For DROP/JZ/EMIT the same read also supplies pop refill. This uses one additional read cycle for some instructions but preserves the existing fault tag contract without new public interfaces.

## P3: ROM escape and continuation addresses

Retain the model's existing behavior: an instruction at ROM_DEPTH-1 may retire and advance PC to ROM_DEPTH; the following fetch faults BAD_BRANCH_TARGET with mnemonic FETCH. CALL at the last word saves ROM_DEPTH, and a later RET restores it before the fetch fault. HALT at the last address stays legal.

Separate architectural PC width `clog2(ROM_DEPTH+1)` from physical ROM address width `clog2(ROM_DEPTH)`. Update PC candidates, return addresses, trace PC fields, fault PC fields, debug signals, and benches together. Never fetch an out-of-range memory index; drive a safe physical address while checking the architectural PC in FETCH. For BRAM fetch faults, acquire valid second-operand context first when tc=1/dc>0.

Add explicit `trace_fetch` and `fault_fetch` flags to distinguish a synthetic fetch failure from an actual 5-bit instruction opcode. The flags accompany the existing trace/fault interfaces; no undefined opcode becomes an executable FETCH instruction. The testbench formats FETCH using the flag; the model remains unchanged. Normal instruction faults still retain their actual opcode.

Support ROM sizes 2 through 32768 and stack/return depths at least 2, with parameter checks for invalid configurations. Compare the complete immediate against integer ROM_DEPTH rather than slicing 32768 to a zero 15-bit value. Test small power-of-two and non-power-of-two ROMs and the default 1024 boundary.

## P4: full-state differential verification and useful generation

Add testbench snapshot records containing every live 40-bit data value and every live return address at each retirement/fault. Reconstruct BRAM order as RAM[0:dc], then top1 if tc=2, then top0 if tc>=1. Snapshot only live storage; stale unused words remain permitted.

```text
on a retirement/fault pulse:
    compare pc, complete live data, complete live returns to model
on any other clock after reset:
    require those architectural values to remain unchanged
on FAULT:
    require the pre-instruction architectural values to remain unchanged
```

Existing helpers will compare these snapshots in addition to TRACE/FINAL so directed and random suites share the stronger check. Random generation will use the executable model to select only successful instruction fragments; include bounded CALL/RET and branch fragments with explicit continuations. Deliberate fault injection must actually fault. Legal generation must halt without falling into padding. Capture seed-based opcode/fault/branch and cache-transition coverage rather than equating seed count with coverage.

Add production stack-capacity boundary tests (512 deep + two cached), arithmetic extremes, constructed noncanonical Boolean and flag-bearing values, and reset/output checks where harness support permits. Mechanically verify handwritten ISA/fault/tag/event mirrors against Python metadata instead of introducing generated semantic logic that would reduce model independence.

## P5: documentation, build, and board evidence

Update stale README, Makefile, interface, and playbook statements. Run the full test suite and metadata consistency check; run synthesis, placement/routing, and packing with the current source and selected program recorded. Capture hardware UART for arithmetic, type-fault, Fibonacci, and a burst-output program if the configured board interface is available. Distinguish physical-board evidence from simulation and explicitly record any unavailable interface.

Retain the two-BRAM budget and inspect routed frequency against the 10 MHz clock. Store commands/logs under this ticket and use tmux for long-running build or serial capture sessions. No external source download is needed: the reviewed code and archived specification already supply the contract.

## Commit, diary, and printed phase boundaries

Print one overall plan and a plan/status slip before/after each of P1–P5 using the brutalist work-slip script. Preserve YAML and command receipts in `reference/slips`. Commit the design before implementation, then commit each tested phase and update the diary with commands, observed failures, design choices, and commit IDs. Finish with docmgr doctor and a clean working tree.

## Risks and alternatives

The main interface change is wider architectural PCs plus explicit fetch flags. All in-repository consumers change in the same phase. The extra BRAM operand-context read trades a cycle for reliable fault metadata. Full snapshots are simulation-only and cost no FPGA memory. The model is still not a formal proof, so independent boundary assertions and fixed expected examples remain valuable.

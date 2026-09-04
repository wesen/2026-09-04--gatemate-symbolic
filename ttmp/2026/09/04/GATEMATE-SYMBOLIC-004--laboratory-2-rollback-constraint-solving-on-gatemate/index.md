---
Title: Laboratory 2 rollback constraint solving on GateMate
Ticket: GATEMATE-SYMBOLIC-004
Status: active
Topics:
    - fpga
    - gatemate
    - symbolic-computers
    - architecture
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: repo://ttmp/2026/09/04/GATEMATE-SYMBOLIC-001--symbolic-computer-patterns-and-designs-for-gatemate-analysis-design-and-intern-implementation-guide/sources/Composable_Hardware_Patterns_for_Symbolic_Computers.md
      Note: Laboratory 2 lines 4569–4961 and rollback substrate lines 3724–3884
ExternalSources: []
Summary: Scope and acceptance criteria for the fixed eight-queens rollback solver.
LastUpdated: 2026-09-04T16:57:27.24532388-04:00
WhatFor: ""
WhenToUse: ""
---


# Laboratory 2: rollback constraint solving

## Project

Build a dedicated FPGA constraint-search engine for eight queens. Each column has an eight-bit domain of permitted rows. Assigning a queen narrows the other domains by removing attacked rows and diagonals; newly singleton domains propagate in turn. A contradiction triggers backtracking. A choice stack remembers untried rows and restart marks, while a mutation trail records old domain masks so changes can be undone in reverse order without copying the entire board at each branch.

The architectural subject is reversible semantic state and irreversible result acceptance. Every protected domain write must reserve and publish its undo information before changing the domain. Choice records retain the selected variable, remaining rows, trail mark, saved propagated bitmap, and resume state. Restoring domain masks alone is insufficient: propagation bookkeeping must also resume correctly. While a solved board is blocked at the result interface, the solution and its supporting state stay stable.

This is the next worked lab, a specialized search controller. The book does not require adding a queens instruction to the existing stack CPU. Reuse of memory, channel, reset, and verification infrastructure should be decided during design.

## Required results

- Complete enumeration emits exactly 92 solutions. Select the first unresolved column and its least significant permitted row to preserve deterministic order.
- The first board, listed by columns zero through seven, is `[0, 4, 7, 5, 2, 6, 1, 3]`.
- `FIRST_ONLY` accepts exactly that first board, crosses a cut fence, discards remaining alternatives, and completes. Do not apply the cut before output acceptance.
- `TRAIL_FULL` stops before the protected domain write; `CHOICE_FULL` stops before branch selection. Deliberately small capacities must exercise these paths.
- Unwind to a choice mark restores the saved domain state and propagated bitmap. Only the unwind engine writes domains during restoration.
- Target roughly two physical RAM blocks and no more than about 3,000 CPEs before simplifying. Record actual mapped resources and timing.

## Book-directed experiment sequence

First build and verify a full-snapshot baseline, then replace domain snapshots with choice points and a mutation trail. Compare bytes written, cycles, and maximum storage; do not assume the trail is cheaper for this small problem without measuring. Use an independent recursive solver for the solution set and a domain/trail model for exact transition order. Add result backpressure, shallow-capacity faults, and FIRST_ONLY, then capture a first-solution trace in simulation and on the FPGA.

## Source and current status

Reread the archived book's entire Laboratory 2 chapter (lines 4569–4961), plus the checkpoint/trail/commit substrate chapter (lines 3724–3884). The original book remains in ticket GATEMATE-SYMBOLIC-001; RelatedFiles points to that authoritative local copy. No new resource download was necessary.

Scope established; implementation has not started. See [tasks.md](tasks.md) for the project sequence and [diary](reference/01-diary.md) for this investigation.

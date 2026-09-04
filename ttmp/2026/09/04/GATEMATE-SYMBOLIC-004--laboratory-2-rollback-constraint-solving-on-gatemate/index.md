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
    - Path: repo://queens_rollback/README.md
      Note: Implemented architecture, interfaces, measurements and reproduction
    - Path: repo://queens_rollback/rtl/queens_core.sv
      Note: Single-owner search and recovery controller
    - Path: repo://ttmp/2026/09/04/GATEMATE-SYMBOLIC-001--symbolic-computer-patterns-and-designs-for-gatemate-analysis-design-and-intern-implementation-guide/sources/Composable_Hardware_Patterns_for_Symbolic_Computers.md
      Note: Laboratory 2 lines 4569–4961 and rollback substrate lines 3724–3884
ExternalSources: []
Summary: Implemented eight-queens solver with snapshot/trail comparison and verified physical GateMate execution.
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

Implementation and physical-board validation are complete. All 58 tests pass. Both recovery modes emit all 92 ordered solutions on the FPGA; FIRST_ONLY emits the first board and an explicit count of one. See [implementation guide](../../../../../queens_rollback/README.md), [tasks](tasks.md), [diary](reference/01-diary.md), and [raw hardware evidence](reference/validation/P5-hardware.json). All phase-boundary print receipts are archived under reference/slips/.

## Physical measurements

| Configuration | CPE_LT | CPE_FF | RAM_HALF | Routed Fmax | UART result |
|---|---:|---:|---:|---:|---|
| snapshot | 2221 | 577 | 3 | 30.19 MHz | 92 boards, exact match |
| trail | 2070 | 460 | 2 | 30.98 MHz | 92 boards, exact match |
| trail-first | 2125 | 467 | 2 | 33.72 MHz | 1 boards, exact match |

All variants use the same 10 MHz constraint and routing seed 2. Frequencies are final post-route estimates, not the earlier placement estimates. Resource columns retain nextpnr units: CPE_LT and CPE_FF are subresources, and RAM_HALF counts half blocks; they must not be relabeled as whole CPEs or whole BRAMs. UART captures were armed before programming and continued for eight seconds. The final loaded image is trail-first. Internal semantic traces are simulation evidence; the board captures expose results and explicit terminal status.

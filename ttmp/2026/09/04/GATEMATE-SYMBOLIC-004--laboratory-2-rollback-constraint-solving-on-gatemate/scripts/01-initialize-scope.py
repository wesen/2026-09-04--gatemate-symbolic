#!/usr/bin/env python3
"""Populate the new Lab 2 ticket from the reread book; no implementation."""
from pathlib import Path

ticket = Path(__file__).resolve().parents[1]
def populate(relative, summary, body):
    path = ticket / relative
    front = path.read_text().split('---', 2)[1]
    front = front.replace('Summary: ""', 'Summary: ' + summary)
    path.write_text('---' + front + '---\n\n' + body.strip() + '\n')

populate('index.md', 'Scope and acceptance criteria for the fixed eight-queens rollback solver.', '''
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
''')

(ticket / 'tasks.md').write_text('''# Tasks

- [x] Reread Laboratory 2 and rollback substrate; establish scope and acceptance criteria.
- [ ] Design semantic state, trace format, memory interfaces, and precise fault boundaries.
- [ ] Implement independent recursive reference and deterministic full-snapshot baseline; verify 92 solutions and first board.
- [ ] Implement domain propagation, choice records, log-before-write mutation trail, and reverse unwind.
- [ ] Compare snapshot versus trail traffic, cycles, and storage high-water marks.
- [ ] Verify randomized result backpressure, shallow capacities, exact restoration, and FIRST_ONLY cut.
- [ ] Synthesize, route, and capture first-solution execution on GateMate; document measured results.
''')

populate('reference/01-diary.md', 'Chronological investigation and implementation diary for Laboratory 2.', '''
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
''')

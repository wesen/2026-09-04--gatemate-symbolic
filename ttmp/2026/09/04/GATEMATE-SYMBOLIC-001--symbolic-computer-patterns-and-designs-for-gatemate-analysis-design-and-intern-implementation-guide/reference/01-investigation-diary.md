---
Title: Investigation Diary
Ticket: GATEMATE-SYMBOLIC-001
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
LastUpdated: 2026-09-04T12:50:16.157400401-04:00
WhatFor: ""
WhenToUse: ""
---

# Diary

## Goal


## Goal

<!-- What is the purpose of this reference document? -->

## Context

<!-- Provide background context needed to use this reference -->

## Quick Reference

<!-- Provide copy/paste-ready content, API contracts, or quick-look tables -->

## Usage Examples

<!-- Show how to use this reference in practice -->

## Related

<!-- Link to related documents or resources -->

Capture the full investigation for ticket GATEMATE-SYMBOLIC-001: importing and studying the
"Composable Hardware Patterns for Symbolic Computers" book, surveying the two prior GateMate
projects (MATE-16 and PCA-Z80), and producing an intern-ready analysis/design/implementation
guide for the first project (Laboratory 1: a precise tagged stack evaluator), stored in the
ticket and uploaded to reMarkable.

## Step 1: Ticket bootstrap and source survey

We created the ticket, imported the source book into `sources/`, added the primary design doc
and this diary, defined the six-phase task list, and printed the up-front plan slip. We then
surveyed the book structure (5 parts, 27+ chapters, 5 laboratories) and both prior GateMate
projects to understand the reusable infrastructure the new project can build on.

### Prompt Context

**User prompt (verbatim):** "Create a new docmgr ticket to go through the symbolic patterns and designs for gatemate, importing ~/Downloads/Composable_Hardware_Patterns_for_Symbolic_Computers.md into the sources/ folder of the ticket and studying it, and also looking at previous gatemate projects in ../ 2026-08-25--vm-cpu-gatemate 2026-08-28--pca-gatemate and then Create a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet points and pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and then upload to remarkable. For the first project. commit at appropriate intervals and keep a detailed diary as you work (using the diary format from the skill) Print out a brutalist work slip with the plan / different phases for the ticket. then before starting a phase, print a slip about the phase, and print one when the phase is done. I attached a gatemate as well"

**Assistant interpretation:** Set up a docmgr research ticket, study the symbolic-patterns book and the two sibling GateMate projects, and write an intern-grade design/implementation guide for the book's first project (Laboratory 1, tagged stack evaluator), then upload to reMarkable; commit at intervals, keep a diary, and print brutalist work slips for the plan and each phase boundary.

**Inferred user intent:** A reusable onboarding artifact that lets a new intern implement the first symbolic-computer laboratory on the GateMate board, grounded in the established project patterns from the two previous repos.

### What I did
- Created ticket GATEMATE-SYMBOLIC-001 with design doc + diary.
- Imported the book (7,697 lines) into ticket `sources/`.
- Added 6 phase tasks; printed the plan work slip, then a "Phase 1 start" slip.
- Surveyed book: Parts I (foundations: pattern contract, commitment, composition calculus, GateMate substrate), II (pattern catalog), III (reusable FPGA substrates), IV (Laboratories 1-5), V (advanced compositions).
- Surveyed `../2026-08-25--vm-cpu-gatemate/mate16`: 29-opcode stack CPU, `tools/opcodes.py` single source of truth, executable reference model `model16.py`, assembler `asm16.py`, held-request I/O bus, UART TX, Makefile targets (versions/test/asm/sim/synth/pnr/bit/load), OSS CAD Suite toolchain.
- Surveyed `../2026-08-28--pca-gatemate/pca_z80`: Z80 on PCA mesh, `z80_isa.py`/`z80_model.py`/`zasm.py`/`placer.py`, `CC_USR_RSTN` config reset cell, phase-driven build-up.

### Why
Establish the evidence base before writing any guide; the guide must reference real files and real, previously verified flows.

### What worked
- Book chapter map and Laboratory 1 full spec (semantic machine, opcode set, controller states, register vs BRAM stack, output commitment, faults, verification plan, GateMate exit criteria) extracted in one pass.
- Prior projects map cleanly onto the book's suggested project organization (models/, rtl/, tb/, tools/).

### What didn't work
- Nothing yet.

### What I learned
- Lab 1 targets: `((7+5)*3)==36 -> BOOL(true)`, plus a precise `TYPE_FAULT` on `BOOL(true)+INT(4)`.
- GateMate stop-build budget for Lab 1: ≤ 2 physical BRAM blocks, ≤ ~2,000 CPEs.
- value40 = tag[3:0] flags[3:0] payload[31:0]; 20-bit instruction format `[19:15] opcode, [14:0] imm`.

### What was tricky to build
- N/A yet (survey phase).

### What warrants a second pair of eyes
- N/A yet.

### What should be done in the future
- Phase 2: deep-read pattern catalog chapters needed by Lab 1 (tagged value word, type dispatch, elastic register, split-lifetime frame).

### Code review instructions
- Inspect ticket layout: `docmgr doc list --ticket GATEMATE-SYMBOLIC-001`.

### Technical details
- Book structure: `grep -n "^# " sources/Composable_Hardware_Patterns_for_Symbolic_Computers.md`.

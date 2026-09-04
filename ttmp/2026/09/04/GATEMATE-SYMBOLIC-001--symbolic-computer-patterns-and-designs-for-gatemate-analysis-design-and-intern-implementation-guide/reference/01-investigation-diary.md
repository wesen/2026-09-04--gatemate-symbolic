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

## Step 2: Deep study of the pattern catalog and Laboratory 1

We read the pattern-catalog chapters relevant to Laboratory 1 — semantic/refinement patterns
(Abstract-Machine Contract, Explicit Semantic State Vector, Structure/Representation Firewall,
Common Fast Path / Precise Slow Path), representation patterns (Tagged Value Word,
Immediate-or-Boxed Split), the reusable substrates of Part III (symbolic_types_pkg with
value40_t/event80_t, rv_reg elastic register and ready/valid ownership rules), the GateMate
substrate chapter (stop-build budgets, block-RAM discipline, toolchain loop), and the complete
Laboratory 1 specification. Together with the MATE-16/PCA-Z80 infrastructure survey, this is
the full evidence base for the intern guide.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Read the book deeply enough to write an accurate intern guide, not a summary.

### What I did
- Extracted `symbolic_types_pkg` (value40_t, event80_t, tag table) verbatim from Part III.
- Extracted the ready/valid ownership rule and `rv_reg` reference implementation.
- Mapped the 8 patterns composed by Laboratory 1 to their catalog pattern cards.
- Confirmed the GateMate toolchain loop (Yosys, nextpnr-himbaechel, gmpack, openFPGALoader) matches the Makefile targets of both prior projects.

### Why
Every claim in the guide must be traceable to a book section or a verified file in a sibling project.

### What worked
- Pattern numbering from the catalog (Pattern 1-5 in ch.7, Pattern 6+ in ch.8 etc.) gives clean API references for the guide.

### What didn't work
- N/A.

### What I learned
- Lab 1 composes exactly 8 named patterns; the guide can be organized around them.
- One-mutation-owner and "no architectural mutation before all checks pass" are the two core invariants.

### What was tricky to build
- N/A (reading phase).

### What warrants a second pair of eyes
- N/A.

### What should be done in the future
- Phase 3/4: write the guide.

### Code review instructions
- `sed -n '4211,4569p' sources/Composable_Hardware_Patterns_for_Symbolic_Computers.md` = full Lab 1 spec.

### Technical details
- Stop-build budget Lab 1: 2 BRAM blocks / 2,000 CPEs.

## Step 3: Design analysis and intern guide written

We translated the Laboratory 1 specification plus the prior-project infrastructure into the
primary deliverable: a complete intern onboarding guide covering background (pattern contracts,
commitment levels), the value representation and ISA, the microarchitecture (FSM, register
stack, BRAM top cache, output commitment, trace), the model-first software method, the
verification ladder, the GateMate substrate/toolchain/budgets, repository layout, a phased
implementation plan, decision records, risks, and full book/prior-project references.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Write the detailed analysis/design/implementation guide for the first project, intern-ready.

### What I did
- Wrote `design-doc/01-intern-onboarding-guide-symbolic-computer-patterns-on-gatemate.md` (~31 KB).
- Organized around the 8 composed patterns from Laboratory 1; every claim line-anchored to the book or to a sibling-repo file.

### Why
The guide is the ticket's primary artifact and the reMarkable upload payload.

### What worked
- Reusing MATE-16 file references (opcodes.py/model16.py/io_block.sv lessons) made the guide concrete rather than generic.

### What didn't work
- Nothing significant.

### What I learned
- The 15-opcode ISA and 7 fault cases fit exactly in the 5-bit opcode field with extension room.

### What was tricky to build
- Keeping the guide faithful to the book (commitment levels, one-mutation-owner, stop-build budgets) without inventing new architecture; every design choice the book leaves open is captured as a decision record (DR-1..DR-4) marked "proposed" for the intern to confirm.

### What warrants a second pair of eyes
- DR-2 (MUL overflow policy) and DR-3 (EQ tag policy) are my proposals, not book mandates — review before implementation.
- Line-number references into the book will drift if the source file is regenerated; they match the imported copy in sources/.

### What should be done in the future
- Validate docs (doctor), upload to reMarkable.

### Code review instructions
- Read the design doc top-to-bottom; spot-check book line refs with `sed -n` against `sources/`.

### Technical details
- Guide sections: 1 exec summary ... 15 glossary; phased plan P0-P6; stop-build budget 2 BRAM / 2000 CPE.

## Step 4: Validation and reMarkable delivery

Doctor passed after adding the missing topic vocabulary (fpga, gatemate, symbolic-computers,
architecture). The design doc + diary were dry-run validated and uploaded as one bundle PDF
to the reMarkable at /ai/2026/09/04/GATEMATE-SYMBOLIC-001.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Validate doc quality, then deliver to reMarkable.

### What I did
- `docmgr vocab add` for 4 topics; `docmgr doctor` -> all checks passed.
- `remarquee upload bundle --dry-run` then real upload; verified with `remarquee cloud ls`.

### Why
Guardrails of the ticket-research skill: doctor-clean before upload, dry-run first.

### What worked
- One-shot doctor pass after vocab fix; upload succeeded first try.

### What didn't work
- None.

### What I learned
- N/A.

### What was tricky to build
- None this step.

### What warrants a second pair of eyes
- Verify the bundle PDF renders the SystemVerilog blocks readably on the tablet.

### What should be done in the future
- Handoff: report ticket path, docs, validation, upload path.

### Code review instructions
- `remarquee cloud ls /ai/2026/09/04/GATEMATE-SYMBOLIC-001 --long`.

### Technical details
- Upload: "GATEMATE-SYMBOLIC-001 Intern Guide.pdf", toc-depth 2, 2 docs.

## Step 5: Implementation P0 — repo bootstrap and board bring-up

We started implementing the intern guide's plan as a new `symbolic_eval/` repo inside this
ticket repo. P0 copies the verified board infrastructure from MATE-16 (constraints, reset
synchronizer, UART transmitter, Makefile flow) and PCA-Z80 (sim-only CC_USR_RSTN model),
builds the blink top, and proves the whole flow on the attached GateMateA1-EVB: sim PASS,
synth/PnR/bit clean, bitstream loaded over DirtyJTAG.

### Prompt Context

**User prompt (verbatim):** "Implement, commit at appropriate intervals and keep a detailed diary as you work (using the diary format from the skill). Print out a brutalist work slip with the plan / different phases for the ticket. then before stsarting a phase, plrint a split about the phase, and print one when the phase is done."

**Assistant interpretation:** Execute the intern guide's P0-P6 implementation plan now: real code, real tests, real hardware, commits and work slips per phase.

**Inferred user intent:** Turn the design doc into a working, verified Laboratory 1 machine on the attached board, with the same diary/slip/commit discipline as the research phase.

### What I did
- Created `symbolic_eval/` with the guide §10 layout (constraints, rtl, sim, tools, programs, scripts, build).
- Copied verified infra: `constraints/olimex_gatematea1_evb.ccf`, `constraints/top.sdc`, `rtl/reset_sync.sv`, `rtl/uart_tx.sv`, `rtl/blink_top.sv`, `sim/tb_blink.sv`, `sim/CC_USR_RSTN.sv`.
- Wrote the Makefile (versions/test/sim/synth/pnr/bit/load) adapted from MATE-16.
- Probed iverilog 14 package/struct/function support with a throwaway test — works (`o=0000000007`), so `symbolic_types_pkg.sv` with packed structs is safe for both iverilog and Yosys.
- `make sim` -> PASS: blink self-test; `make bit` -> clean; `make load` -> "Load SRAM via JTAG: 100.00% Done" (LED blinking on board).

### Why
P0 of the guide: prove clock pin, LED pin, toolchain, config reset, and the evidence workflow before any core RTL exists.

### What worked
- The MATE-16 flow copies over with only path changes; the board (ttyACM0/1, DirtyJTAG via RP2040) loaded first try.

### What didn't work
- First `cp` used `../` instead of `../../` for the sibling projects (paths from inside `symbolic_eval/`); fixed immediately, no damage.

### What I learned
- iverilog 14.0 (devel) fully supports SV packages, packed structs, and package functions — no need for the include-file fallback.

### What was tricky to build
- Nothing this phase; this is exactly why the bootstrap phase exists.

### What warrants a second pair of eyes
- Makefile still hardcodes `blink_top` in synth/pnr; P5 must switch to the real `top`.

### What should be done in the future
- P1: `tools/opcodes.py` + `tools/stack_model.py` + model tests.

### Code review instructions
- `cd symbolic_eval && make sim` (PASS), `make versions`, `git log`.

### Technical details
- Yosys 0.68+130, nextpnr-himbaechel, gmpack, openFPGALoader -b olimex_gatemateevb.

## Step 6: Implementation P1 — opcode table and reference model

We implemented the authoritative instruction table (`tools/opcodes.py`: 15 opcodes, 10 tags,
8 fault codes, 20-bit word codec) and the executable reference model (`tools/stack_model.py`:
Machine with precise faults, commit-trace records). The model reproduces the book's Program A
commit trace character-for-character and Program B's precise TYPE_FAULT. 30 pytest model tests
green.

### Prompt Context

**User prompt (verbatim):** (see Step 5)

**Assistant interpretation:** Implement guide phase P1 (types + model + tests) before any RTL.

### What I did
- `tools/opcodes.py`: single source of truth; Instruction dataclass in the MATE-16 `opcodes.py` shape; `encode/decode/sign_extend_s15`.
- `tools/stack_model.py`: `Value40`, `TraceRecord` (canonical TRACE line format shared with future RTL testbenches), `Machine.step()` implementing every opcode + fault with check-before-mutation discipline; hex loader.
- `sim/test_model.py` + `sim/conftest.py`: 30 tests (book directed tests 1,2,4,5,7,8 + all fault cases + constructed-state tests for NONCANONICAL_BOOL and ARITH_OVERFLOW).

### Why
Model-first: the model is the differential oracle for all RTL phases; its trace format is the contract.

### What worked
- Program A trace matched the book's expected trace table exactly on the first full run after fixing one model bug.

### What didn't work
- First version of `_commit` in the model captured `pc_old` AFTER mutating `self.pc` (bogus `pc_prev`/`rec_pc_old` helpers) — caught by review before running tests; fixed by capturing `pc_old` at function entry.
- 9 test failures on first run: my tests ended programs without HALT and assumed running off the end stops the machine. It does not — unwritten ROM words are 0x00000 = PUSH_S15 0 (zero-filled ROM), so the machine keeps pushing until STACK_OVERFLOW. Fixed the tests and pinned the semantics as defined behavior (RTL ROM must explicitly zero-init to match).

### What I learned
- Zero-word-equals-legal-instruction means `test_pc_past_program` becomes an overflow test; the assembler must always terminate programs with HALT.

### What was tricky to build
- Precise-fault check ordering (capacity -> tags -> canonicality -> branch target) had to be fixed once and documented, because the RTL and tb must replay exactly the same order to produce identical FAULT trace lines.

### What warrants a second pair of eyes
- The `FINAL` line format and TRACE line format are now a frozen contract between model, RTL testbenches, and compare scripts.

### What should be done in the future
- P2: `tools/asm20.py` two-pass assembler + programs/ fixtures.

### Code review instructions
- `cd symbolic_eval && python3 -m pytest sim/test_model.py -q` (30 passed).

### Technical details
- pytest 9.1.1; test runtime 0.06 s.

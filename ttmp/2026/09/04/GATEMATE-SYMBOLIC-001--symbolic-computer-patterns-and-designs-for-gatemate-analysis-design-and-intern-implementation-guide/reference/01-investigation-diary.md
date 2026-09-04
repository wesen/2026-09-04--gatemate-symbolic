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

## Step 7: Implementation P2 — assembler and programs

We implemented the two-pass assembler `tools/asm20.py` (labels, s15/u15 operand validation,
raw `WORD` directive for fault-case words, .hex/.lst/.sym.json outputs, zero-padding to ROM
depth) and the 11 test programs in `programs/`, including the two book exit-criteria programs.
42 pytest tests green (30 model + 12 assembler).

### Prompt Context

**User prompt (verbatim):** (see Step 5)

**Assistant interpretation:** Implement guide phase P2 (assembler + programs).

### What I did
- `tools/asm20.py`: two-pass, no eval, MATE-16 asm16.py discipline; hex output is one 20-bit word per line (5 hex digits) padded to ROM depth.
- Programs: arith (exit A), typefault (exit B), smoke, deep (top-cache spill/refill exercise), branch (JZ taken/not-taken + backward loop), muloverflow, underflow, badop (WORD 0xF8000), badbranch (WORD 0x5FFFF = JMP 0x7FFF), jztype, stackoverflow.

### Why
The assembler + programs are the inputs to every RTL directed test in P3/P4.

### What worked
- Model + assembler together caught three real program bugs before any RTL existed — exactly the point of model-first.

### What didn't work
- Listing corruption bug: comment/label-only lines made `listing` index diverge from word index, so pass-2 replacements clobbered neighboring lines (ADD/MUL lines vanished, PUSH placeholders duplicated). Fixed by recording the listing position in each pending entry.
- `encode()` immediate-range assert initially restricted to s15, breaking legal u15 branch targets (encode(0x0B, 0x7FFF)); now allows the full 15-bit representable range, kind-specific validation stays in the assembler.
- My own branch.asm used an INT as the JZ condition (TYPE_FAULT) and inverted the JZ polarity; badop.asm encoded WORD 0x1F (= PUSH_S15 31) instead of opcode 0x1F (= 0xF8000).

### What I learned
- JZ branches on FALSE: the canonical countdown loop is `EQ 0; JZ loop` (continue while nonzero), falling through on zero.
- Encoding rule of thumb: word = (opcode << 15) | (imm & 0x7FFF).

### What was tricky to build
- Keeping .lst truthful across two passes when not every source line emits a word; solved by carrying (word_index, listing_pos) in pending.

### What warrants a second pair of eyes
- `assemble()` returns unpadded words while the .hex writer pads to ROM depth — the model and RTL both consume the padded hex, so they agree; document this.

### What should be done in the future
- P3: register-stack RTL (`symbolic_types_pkg.sv`, `stack_core.sv`, testbench, differential runner).

### Code review instructions
- `cd symbolic_eval && python3 -m pytest sim/ -q` (42 passed); `make asm PROG=arith && cat build/arith.lst`.

### Technical details
- Exit criteria verified at model level: arith -> FINAL 1 NONE 8 0 1, output BOOL(1); typefault -> FINAL 0 TYPE_FAULT 2 2 0.

## Step 8: Implementation P3 — register-stack RTL, differentially verified

We implemented `rtl/symbolic_types_pkg.sv` (tags, value40_t, opcodes, faults, events — mirrors
tools/opcodes.py) and `rtl/stack_core.sv` (register-stack version of the Lab 1 evaluator:
9-state FSM, one mutation owner, precise faults, EMIT commitment via ready/valid, 64-bit
arithmetic with precise signed 32-bit overflow detection), plus `sim/tb_stack_core.sv` (ROM
host, random backpressure, blocked-output-stability and pc-stability assertions, model-format
TRACE/FINAL output) and `sim/test_directed.py` (differential: model vs RTL line-by-line, plus
stall-seed reruns). 54 tests green; the RTL reproduces the book's Program A commit trace
exactly.

### Prompt Context

**User prompt (verbatim):** (see Step 5)

**Assistant interpretation:** Implement guide phase P3: register-stack RTL + directed differential tests.

### What I did
- `rtl/symbolic_types_pkg.sv`, `rtl/stack_core.sv` (EXECUTE stages complete next-state data: npc/ndepth/two write slots/pending/halt; COMMIT is the single mutation owner; faults pulse from EXECUTE with precise pc/depth).
- `sim/tb_stack_core.sv`: +rom/+stall_seed plusargs, registered stall bit, pre-edge-sampled monitors, watchdog.
- `sim/test_directed.py`: 11 programs x (model trace == RTL trace), 5 stall-seed reruns, exit-criteria assertions.

### Why
The register version isolates semantics and interface behavior from RAM latency (book Lab 1 "Register-stack version").

### What worked
- First full run of arith on the RTL printed the book's expected commit trace character-for-character after the trace-register fix.

### What didn't work
- iverilog: `string`-typed ternary in $display ("Data types string and bool") -> replaced with an if/else into a local string.
- Trace snapshot regs were declared and defaulted but never assigned in the always_ff (only trace_valid_q was) -> all trace fields printed X; fixed by registering all twelve snapshot regs.
- tb assertion races: monitors that sample with `#1` after posedge mixed pre/post-NBA values with the randomizer also using `#1` -> spurious "blocked output not stable". Fixed by making out_ready a registered stall bit and sampling monitors with pre-edge blocking reads + NBA capture.
- Model/RTL hex case mismatch (%08X vs %h lowercase) -> standardized on lowercase in the model.
- My exit-criteria test indexed rtl[-2] forgetting the FINAL line is included.

### What I learned
- In iverilog, keep monitors race-free: registered stimulus, blocking reads at posedge, NBA capture of previous-cycle values.
- iverilog emits "sorry: constant selects..." notes for part-selects in always_comb — informational, not errors.

### What was tricky to build
- Making EXECUTE stage *complete* next-state data (two write slots for SWAP, staged halt, staged fault record with trace pulse fired from the EXECUTE comb block via a task) so COMMIT stays the only place registers mutate.

### What warrants a second pair of eyes
- The do_fault task fires the FAULT trace pulse in the EXECUTE cycle (not a dedicated FAULT state cycle); the state machine then sits in S_FAULT. Confirm no double-pulse path exists (state_d=S_FAULT only entered once since EXECUTE is not re-entered).

### What should be done in the future
- P4: BRAM stack + two-entry top cache + spill/refill boundary tests + random legal/illegal instruction tests.

### Code review instructions
- `cd symbolic_eval && python3 -m pytest sim/ -q` (54 passed).

### Technical details
- Check order frozen: capacity/underflow -> tags -> canonicality -> branch target.

## Step 9: Implementation P4 — BRAM stack with two-entry top cache

We implemented the Split-Lifetime Frame refinement: `rtl/sync_sdp_ram.sv` (request/response-
valid simple dual-port RAM) and `rtl/stack_core_bram.sv` (top0/top1 cache registers over the
deep RAM, tc/dc/dt counters, prefetched RAM reads at DECODE, staged representation updates,
single mutation owner). The BRAM core passes the identical differential suite as the register
core plus the refinement invariant `depth == tc + dc` checked every cycle in the testbench,
plus 24 random-program tests (typed-stack generator with occasional injected illegal
instructions) and 12 random-program tests on the register core, all with random output
backpressure. 101 tests green total.

### Prompt Context

**User prompt (verbatim):** (see Step 5)

**Assistant interpretation:** Implement guide phase P4: BRAM top-cache refinement + boundary/random verification.

### What I did
- `rtl/sync_sdp_ram.sv`, `rtl/stack_core_bram.sv` (states: RESET/FETCH/FETCH_WAIT/DECODE/RDWAIT/EXECUTE/COMMIT/OUTPUT_WAIT/FAULT/HALTED; read purposes R_OPND/R_FILL/R_POP/R_FILL2; DECODE prefetches RAM[dc-1] when the representation will need it — reads are side-effect-free so prefetching before checks cannot change behavior).
- `sim/tb_stack_core_bram.sv`: tb_stack_core.sv + invariant depth==tc+dc each cycle.
- `sim/test_bram.py`: BRAM directed (11 programs), random generator (typed stack model, ~6% injected illegal instructions: ADD-on-mixed-tags, DROP-underflow, BAD_OPCODE), model-vs-RTL with stall seeds; register core on the same random programs.

### Why
The book's exercise 2 asks to prove the two-entry top-cache refinement across push, pop, binary operation, and output stall — the differential trace equality plus the per-cycle invariant is that proof at the testing level.

### What worked
- The core passed all 11 directed programs on the first full run after elaboration fixes.

### What didn't work
- iverilog: enum ternary "This assignment requires an explicit cast" in state_d assignment -> if/else.
- 10 random-test failures, all one line: BAD_OPCODE fault records disagreed on operand tags. Root cause: the model hardcoded tag1/tag0=0 for faults without explicit tags, while the RTL (correctly, per the book's "operand tags" field) reports the current top-of-stack tags. Fixed the model to default to real top-two tags (0 when absent) and guarded the BRAM core's tags by depth (stale cache registers must not leak into fault records when depth < 2).
- Two staging bugs caught by my own review before simulation: BIN refill dc deltas (tc==2: dc-1; tc==1: dc-2) and push spill staging; also a botched edit that briefly deleted SWAP's underflow check — restored and verified by brace-count inspection.

### What I learned
- The refinement's subtle case: binary op with tc==1 needs TWO sequential RAM reads (operand at RAM[dc-1], new top1 at RAM[dc-2]) — prefetched operand at DECODE, second read issued from EXECUTE (R_FILL2).
- EMIT's pop-refill can be prefetched at DECODE because no write can occur between DECODE and acceptance.

### What was tricky to build
- Keeping the invariant true at every trace pulse: all count registers (tc/dc/dt/depth/pc) update atomically at the commit point; RAM reads that repair the representation are internal stuttering before the pulse.

### What warrants a second pair of eyes
- The DECODE prefetch logic (R_OPND/R_FILL/R_POP selection) is the densest code in the repo; the random tests cover it, but a table-driven review against the book's push/pop/binary algorithms is worthwhile.

### What should be done in the future
- P5: board top (rv_reg output stage, UART value printer, LED state), synthesis against the stop-build budget, load to the GateMate.

### Code review instructions
- `cd symbolic_eval && python3 -m pytest sim/ -q` (101 passed, ~13 s).

### Technical details
- BRAM configs tested: DEEP_DEPTH=30 (total 32) and DEEP_DEPTH=6 (total 8, overflow program).

## Step 10: Implementation P5 — board integration, budget, hardware evidence

We integrated the board top (`rtl/top.sv`: CC_USR_RSTN + reset_sync + button-abort restart,
stack_core_bram, program_rom, rv_reg elastic output stage, 13-byte-per-value UART printer,
LED status), verified it end-to-end in simulation (tb_top decodes the UART pin at 115200 8N1),
synthesized within the stop-build budget, and captured the book's two exit criteria on real
hardware over the RP2040 USB serial link.

### Prompt Context

**User prompt (verbatim):** (see Step 5)

**Assistant interpretation:** Implement guide phase P5: board top, synthesis budget check, load, hardware evidence.

### What I did
- `rtl/rv_reg.sv` (book ch. 15 elastic register), `rtl/program_rom.sv` (PROG_HEX macro init, sync read), `rtl/top.sv`.
- `sim/tb_top.sv` + `sim/test_top.py`: full-board sim; UART byte stream compared against the model's EMIT values (`T<tag>:<PAYLOAD-hex>\r\n`).
- `scripts/synth.ys` (MATE-16 pattern); Makefile synth/pnr/bit/load now target `top`.
- Loaded both exit-criteria programs on the GateMateA1-EVB and captured UART.

### Why
The laboratory's completion definition is hardware evidence, not simulation.

### What worked
- Yosys stats: 2 x CC_BRAM_20K (exactly the book's 2-block budget: 1Kx20 ROM + 512x40 stack), 1 CC_MULT, 1894 cells -> 392 CPEs after packing (budget ~2000). nextpnr: 17.11 MHz max (PASS at 10 MHz).
- Hardware: Program A sent exactly `T1:00000001\r\n` (BOOL(true)) at 115200 8N1 over /dev/ttyACM0, then halted. Program B sent nothing (precise TYPE_FAULT, no output transfer).

### What didn't work
- A chain of yosys SV limitations, each fixed portably (no `return` in functions -> classic function-name assignment; no file-scope/module-header `import` -> fully-qualified `pkg::name` references everywhere via a scripted pass; no multi-declarator struct typedefs -> one per line; no `parameter string` and macro-string $readmemh via -D on the command line -> .ys script file with yosys's own tokenizer, per the MATE-16 pattern; enum-typed struct fields broke width inference in package functions -> plain logic [3:0] tag field).
- First hardware capture was empty: the program runs at configuration time and its output was gone before the port opened; fixed by reading the serial port concurrently with the bitstream load (and the button-abort restart exists for re-runs).
- tb_top initially waited only 40 bit-times after halt, truncating the printer drain (EMIT commits at rv_reg acceptance, not at wire transmission) -> extended to 60 byte-times.
- test_top expected lowercase payload hex; the printer deliberately emits uppercase -> fixed the test.
- The diary append for this step used `../../ttmp/...` from inside symbolic_eval/ and silently failed (bash: No such file or directory); re-appended from the repo root - always run diary writes from the ticket root.

### What I learned
- "EMIT commits at acceptance" means acceptance by the elastic stage; the E-level (wire) event can come tens of thousands of cycles later - the tb drain window must cover it.
- On this board, /dev/ttyACM0 is the RP2040 CDC UART; ACM1 is the JTAG control channel.

### What was tricky to build
- Keeping one code base synthesizable by yosys AND simulatable by iverilog: every SV feature used had to be in the intersection; the qualification pass (pkg:: prefix) had to avoid comments and port names (word-boundary regex on a fixed name list).

### What warrants a second pair of eyes
- The board evidence for Program B is "no UART bytes + documented LED coding"; an ILA or trace-RAM capture of the fault record would make it undeniable - left as an extension.

### What should be done in the future
- P6: README finalization, ticket wrap-up.

### Code review instructions
- `cd symbolic_eval && make bit PROG=arith && make load` (board prints `T1:00000001`).
- `python3 -m pytest sim/ -q` (105 passed).

### Technical details
- Hardware capture: `stty -F /dev/ttyACM0 115200 raw -echo; cat /dev/ttyACM0` during `openFPGALoader -b olimex_gatemateevb build/top.bit`.

## Step 11: Implementation P6 — wrap-up

Final README (exit criteria, machine summary, method, measured results, test inventory,
portability notes), full test re-run (105 green), tool versions recorded, ticket tasks
closed out.

### Prompt Context

**User prompt (verbatim):** (see Step 5)

**Assistant interpretation:** Close out the implementation: documentation, final validation, handoff.

### What I did
- Rewrote `symbolic_eval/README.md` with hardware evidence, budgets, LED/UART documentation, and the yosys/iverilog portability notes.
- Re-ran the whole suite: 105 passed.

### Why
An intern must be able to reproduce and extend the work from the README alone.

### What worked
- Everything green on the final run; board left running Program A.

### What didn't work
- Nothing.

### What I learned
- The Laboratory 1 exit criteria are fully achievable within the book's budgets with margin (392/2000 CPEs, 2/2 BRAM).

### What was tricky to build
- N/A.

### What warrants a second pair of eyes
- Final repo state: `git log --oneline` (7 implementation commits), `python3 -m pytest sim/ -q`.

### What should be done in the future
- Extensions: CALL/RET, trace RAM / ILA fault capture, Laboratories 2-5 reuse of this substrate.

### Code review instructions
- Start at `symbolic_eval/README.md`, then `rtl/stack_core_bram.sv`, then run the suite.

### Technical details
- Final: 105 tests, 2 BRAM, 392 CPEs, 17.11 MHz, board evidence captured for both exit criteria.

## Step 12: Extension — LED polarity fix and CALL/RET with interesting programs

User feedback: "Is this working according to plan? can you make more interesting programs?
with CALL/RET? how hard is trace-RAM/ILA? can you blink the LED? I don't see the LED being
on." We fixed the LED (the EVB user LED is active-low; our "solid ON" drove the pin high =
dark), implemented the book's CALL/RET extension (separate 16-entry return-address stack,
two new precise faults, fault codes widened to 4 bits), wrote recursive fib(10) plus sq,
countdown, and both rstack fault programs, and verified everything differentially (123
tests) and on hardware (fib emits `T0:00000037` = 55).

### Prompt Context

**User prompt (verbatim):** "Is this working according to plan? can you make more interesting programs? with CALL/RET? how hard is trace-RAM/ILA? can you blink the LED? I don't see the LED being on."

**Assistant interpretation:** Confirm overall status; fix the LED visibility bug; add CALL/RET with genuinely interesting programs; estimate trace-RAM/ILA effort.

**Inferred user intent:** The machine should be visibly, demonstrably alive on the board and do something more impressive than one arithmetic EMIT.

### What I did
- E1 LED: `assign user_led = ~led_logic` (LiteX names the pin `user_led_n`; halted now lights it solid).
- E2 CALL/RET: `opcodes.py` OP_CALL=0x0F (u15), OP_RET=0x10; fault codes widened 3->4 bits (RSTACK_UNDERFLOW=8, RSTACK_OVERFLOW=9); model return stack + semantics (CALL: capacity check then target; RET: underflow check); both RTL cores got a 16x10-bit register return stack with the staged-write/single-mutation-owner discipline; tbs print rdepth in FINAL/TOPDONE; `programs/fib.asm` (recursive fib(10)), `sq.asm`, `countdown.asm`, `retunderflow.asm`, `calloverflow.asm`.
- E3: synth (2 CC_BRAM_20K, 392 CPEs, 16.56 MHz — budget unchanged), loaded fib on the board, captured `T0:00000037\r\n`.

### Why
The book lists CALL/RET as the first Laboratory 1 extension; recursion is the interesting case for a continuation state (exercise 4).

### What worked
- fib(10)=55 first try on hardware after the tb mnemonic fix; all 123 tests green.

### What didn't work
- A series of silently non-matching patch patterns (comments had been rewritten by the earlier qualification pass; whitespace differed): the EXECUTE CALL/RET insert, the BRAM RSTACK_DEPTH parameter, and a port-width fix all no-oped while printing "patched". Symptom: RTL executed PUSH then BAD_OPCODE at the CALL; the model diff caught everything. Lesson recorded: every scripted patch now asserts the pattern count and the result (grep) before moving on.
- My first countdown.asm was wrong twice over (the subroutine kept its argument AND the loop fed JZ an INT): model caught TYPE_FAULT at pc 6; rewrote with show consuming the copy and an EQ-based zero test.
- The tb op_name lacked CALL/RET so correct traces printed "BAD" as the mnemonic.

### What I learned
- The EVB LED polarity is active-low (user_led_n) — the only reason "halted" looked dead while the machine was fine (UART proved it).
- fib's stack discipline on a pure data stack: keep n alive across both recursive calls via DUP-before-CALL and SWAP between them.

### What was tricky to build
- Widening the fault-code field touched every layer (pkg localparams, core ports/regs, do_fault, tb display) — a frozen-format change (FINAL/TOPDONE lines gained an rdepth field) that had to stay identical across model, both tbs, and all hardcoded test expectations.

### What warrants a second pair of eyes
- rdepth in the trace FAULT records is still the DATA depth (rstack depth only in FINAL); confirm that is the intended contract.
- The random-program generator does not emit CALL/RET (directed programs cover it); extending it would exercise rstack boundaries randomly.

### What should be done in the future
- Trace-RAM (estimate: ~1-2 h, +1 CC_BRAM_20K excluded from the stop-build budget as debug instrumentation): record commit packets into a RAM on trace_valid, stream them over the existing UART in the TRACE line format after HALT/FAULT, diff against the model — closes the model<->hardware loop fully. The GateMate ILA itself needs the proprietary Cologne Chip flow (nextpnr cannot insert one), so trace-RAM is the OSS-native answer.

### Code review instructions
- `python3 -m pytest sim/ -q` (123 passed); `make bit PROG=fib && make load` -> board prints `T0:00000037`, LED solid ON after ~2 ms.

### Technical details
- Budget after extension: 2 x CC_BRAM_20K, 392 CPEs, 16.56 MHz, 1 CC_MULT.

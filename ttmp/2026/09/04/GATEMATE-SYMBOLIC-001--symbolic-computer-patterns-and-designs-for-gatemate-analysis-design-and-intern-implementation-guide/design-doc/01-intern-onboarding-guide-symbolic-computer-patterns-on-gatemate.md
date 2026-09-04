---
Title: ""
Ticket: ""
Status: ""
Topics: []
DocType: ""
Intent: ""
Owners: []
RelatedFiles:
    - Path: abs:///home/manuel/code/wesen/2026-08-25--vm-cpu-gatemate/mate16/rtl/mate16_core.sv
      Note: multi-cycle FSM + precise fault precedent
    - Path: abs:///home/manuel/code/wesen/2026-08-25--vm-cpu-gatemate/mate16/tools/model16.py
      Note: reference-model pattern reused in the guide
    - Path: abs:///home/manuel/code/wesen/2026-08-25--vm-cpu-gatemate/mate16/tools/opcodes.py
      Note: authoritative opcode table pattern reused in the guide
    - Path: abs:///home/manuel/code/wesen/2026-08-28--pca-gatemate/pca_z80/rtl/top.sv
      Note: CC_USR_RSTN startup reset precedent
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Intern Onboarding Guide: A Precise Tagged Stack Evaluator on the GateMate FPGA

*Source book: `sources/Composable_Hardware_Patterns_for_Symbolic_Computers.md` (Laboratory 1, lines 4211–4569; substrate chapters 683–808, 3397–3550, 3550–3724).*

---

## 1. Executive summary

You are going to build **a small computer whose purpose is to compute with typed (tagged) values
safely**, on a real FPGA board (the Olimex GateMateA1-EVB with the Cologne Chip CCGM1A1 chip),
using only open-source tools. It is "Laboratory 1" of a five-laboratory sequence in the book
*Composable Hardware Patterns for Symbolic Computers*. Everything later (rollback solving,
dataflow, lazy reduction, a relational engine) reuses what you build here.

The machine is a **stack evaluator**: it executes bytecode like `PUSH 7; PUSH 5; ADD; ...` where
every value on the stack is a **tagged word** — 4 bits of *what it is* (integer, Boolean,
reference, ...) plus 32 bits of *what it holds*. The defining property of the machine is not
speed but **precision**: if anything is wrong — wrong type, stack underflow, overflow, bad
opcode — the machine enters a named fault state and **the program counter and stack look exactly
as they did before the failing instruction**. No half-applied arithmetic, no popped-then-error.

Two programs are your exit criteria:

```text
Program A:  PUSH_S15 7; PUSH_S15 5; ADD; PUSH_S15 3; MUL; PUSH_S15 36; EQ; EMIT; HALT
            -> emits exactly one value, BOOL(true), then halts.

Program B:  PUSH_TRUE; PUSH_S15 4; ADD
            -> enters TYPE_FAULT at the ADD, stack and pc unchanged, no output.
```

Two prior projects in the same family already solved the "how do I build on this board" problem,
and you will copy their infrastructure shamelessly:

- **MATE-16** — `../2026-08-25--vm-cpu-gatemate/mate16/`: a 16-bit stack CPU on the same board.
  Same board constraints, same toolchain, same "opcode table as single source of truth" method,
  same reference-model-first verification. Your project is MATE-16's method applied to *typed*
  values instead of raw 16-bit words.
- **PCA-Z80** — `../2026-08-28--pca-gatemate/pca_z80/`: a Z80 on a message-passing cell mesh.
  Same repo layout, Makefile, and phase discipline. Look at it when you want to see how a
  bigger system was built up phase by phase.

This guide explains every part of the system: the concepts (pattern contracts, commitment
levels, tagged values), the architecture (state machine, stacks, output FIFO, faults), the
method (reference model first, differential testing), the toolchain, and a phased
implementation plan with pseudocode, diagrams, and exact file references.

---

## 2. Problem statement and scope

**In scope** (Laboratory 1):

- One 40-bit tagged-value stack machine with the 15-opcode ISA in §5.
- A 20-bit instruction format, program ROM.
- A register-stack implementation first, then a BRAM stack with a two-entry top cache.
- Precise faults for the seven named fault cases.
- A ready/valid output channel (the `EMIT` commitment boundary).
- A Python reference model and assembler, differential-tested against the RTL.
- GateMate synthesis, board deployment, hardware evidence.

**Out of scope**: calls/returns (Laboratory 1 extension), heap allocation, garbage collection,
rollback/trails (Lab 2), elastic multi-unit systems (Lab 3), graph reduction (Lab 4),
unification/queries (Lab 5). Keep the ISA exactly as specified so later labs compose cleanly.

---

## 3. Background you need before touching RTL

### 3.1 What "symbolic" means here

The book's phrase *symbolic computer* covers machines whose work is dominated by **symbolic
values, dynamic structure, irregular control, and managed memory** — logic programs, functional
languages, constraint solvers, query engines — as opposed to dense numeric loops. The common
denominator is: values carry **metadata** (a tag), control flow is **data-dependent**, and
correctness depends on **precise, ordered, recoverable state updates**. Your tagged evaluator is
the smallest machine that exhibits all three properties.

### 3.2 The hardware pattern contract

The book's central tool (chapters 3–5) is a way of describing hardware blocks as **contracts**
rather than as blocks. Every pattern names nine fields:

```text
P = <S, D, F, O, V, R, B, T, I>
      state, representation, flow, ordering, visibility,
      recovery, bounds, timing, invariant
```

For your project the two most load-bearing consequences are:

- **No named-block thinking.** A "FIFO" or "ALU" is not a design; the question is always
  *who owns the mutation, when is it visible, what happens on fault*. If a proposed mechanism
  has no overflow behavior or no rollback owner, the empty field **is** the design finding.
- **Composition closure.** When you connect the evaluator to the output FIFO, ask: does the
  composition preserve each side's ordering guarantees? capacity bounds? fault handling?
  (Chapter 5's eight composition laws.)

### 3.3 Commitment levels — the key mental model

The book defines four levels of "how real is this state change" (chapter 4):

```text
µ  microarchitecturally provisional  — inside a pipeline stage, invisible
A  architecturally retired           — pc advanced, stack updated, undeniable
S  semantically committed            — observable in the machine's meaning
E  externally visible                — left the chip / changed a pin
```

Your machine maps them concretely:

- The ALU result sitting in a register while `COMMIT` has not fired: level **µ**.
- `pc_q` and `depth_q` updated on a commit pulse: level **A**.
- The value popped from the stack by `EMIT`: **S** only when the output channel accepted it.
- The UART byte on the wire: **E**.

**The one rule that makes the whole laboratory work: no architectural (A-level) mutation may
occur before all checks for the instruction have passed.** Underflow check, tag check, overflow
check — all of them — happen in `EXECUTE`; only then does `COMMIT` write registers. This is why
faults are precise for free.

### 3.4 The eight patterns you are composing

Laboratory 1 composes exactly these catalog patterns (book ch. 7–10; catalog pattern numbers in
parentheses):

| Pattern | Catalog ref | One-line role in your machine |
|---|---|---|
| Abstract-Machine Contract | Pattern 1 (ch. 7) | bytecode transitions defined independent of cycles |
| Explicit Semantic State Vector | Pattern 2 (ch. 7) | `M = <pc, stack, output, fault, halted>` named in the model |
| Structure/Representation Firewall | Pattern 3 (ch. 7) | model doesn't know registers-vs-BRAM; RTL doesn't know semantics |
| Common Fast Path / Precise Slow Path | Pattern 4 (ch. 7) | fast typed ADD; precise TYPE_FAULT slow path |
| Tagged Value Word | Pattern 6 (ch. 8) | `value40_t` with tag/flags/payload |
| Immediate-or-Boxed Split | Pattern 7 (ch. 8) | small ints inline in the word; (extension) boxed values |
| Split-Lifetime Frame | ch. 9 | top-of-stack cache over synchronous RAM |
| Delayed Irreversible Store | ch. 10 / 25 | `EMIT` commits only through ready/valid acceptance |

When a design question comes up ("what happens if the FIFO is full when EMIT fires?"),
find the pattern card and read its *failure sign* and *verification target* fields. The book
wrote them so you don't have to rediscover the bugs.

---

## 4. System overview

### 4.1 Block diagram

```text
                       +---------------------------------------------+
                       |                stack_core                   |
                       |                                             |
  program_rom -------->| FETCH_REQ/FETCH_WAIT -> ir_q (20 bits)      |
  (1K x 20, BRAM)      |        |                                   |
                       |     DECODE                                 |
                       |        |                                   |
  tag checks, depth -->|     EXECUTE  (all precondition checks)     |
  checks               |        |  next-state data complete            |
                       |     COMMIT  ---- commit pulse ------------->|---> trace packet
                       |        |                                   |     (event80)
                       |   STACK_READ/WAIT  (BRAM version)          |
                       |        |                                   |
   stack RAM <-------->|   two-entry top cache: top0,top1            |
   (BRAM version)      |                                             |
                       |     OUTPUT_WAIT -- out_valid/out_data ----->|---> rv_reg / FIFO
                       |        |        (pop only on out_ready)     |      -> UART TX
                       |   FAULT, HALTED                             |
                       +---------------------------------------------+
```

### 4.2 The value representation (Pattern 6)

Every value everywhere in the machine is a `value40_t`, defined in one shared package (book
ch. 14, verbatim contract):

```systemverilog
package symbolic_types_pkg;
  typedef enum logic [3:0] {
    TAG_INT = 4'h0, TAG_BOOL = 4'h1, TAG_REF = 4'h2, TAG_ATOM = 4'h3,
    TAG_PAIR = 4'h4, TAG_THUNK = 4'h5, TAG_IND = 4'h6,
    TAG_ERROR = 4'hD, TAG_POISON = 4'hE, TAG_EMPTY = 4'hF
  } value_tag_t;

  typedef struct packed {
    value_tag_t  tag;
    logic [3:0]  flags;
    logic [31:0] payload;
  } value40_t;
endpackage
```

For Laboratory 1 you only *produce and consume* `TAG_INT` and `TAG_BOOL`; the other tags exist
so later laboratories extend the design without re-defining the word. `payload` for `BOOL`
must be canonical: `BOOL(1)`/`BOOL(0)` only — a `BOOL` with payload 7 is `NONCANONICAL_BOOL`,
a required fault case.

**Why 40 bits?** It packs nicely into GateMate block RAMs (20/40/80-bit organizations), and it
holds an immediate 32-bit integer inline (Immediate-or-Boxed Split: small values never need a
heap). It is a *physical convention*, not an architectural promise — the model must not care.

### 4.3 The semantic machine (Patterns 1, 2)

The executable specification, independent of any cycle structure:

```text
M = <pc, stack, output_stream, fault, halted>

step(M):
  ir = ROM[pc]
  switch ir.op:
    PUSH_S15 k: require capacity; push INT(signext(k,15)); pc++
    PUSH_TRUE / PUSH_FALSE: push BOOL(1)/BOOL(0); pc++
    ADD:  require depth>=2, both tags INT, no signed overflow
          -> replace top two with INT(sum); depth--; pc++
          else fault(TYPE_FAULT | ARITH_OVERFLOW), state unchanged
    SUB, MUL: analogous (MUL: low 32 bits, or precise overflow — decide, document)
    EQ:  x,y -- BOOL(x == y)      (equality defined by tag policy)
    LT:  INT a, INT b -- BOOL(a < b)
    DUP / DROP / SWAP: capacity/underflow checked first
    JMP t: pc = t
    JZ t: require top tag BOOL; pop; pc = c ? t : pc+1
    EMIT: pop only on output acceptance (Delayed Irreversible Store)
    HALT: halted = true
```

Notice what this model **does not contain**: cycles, RAM latency, register files, FIFOs. That
omission is the Structure/Representation Firewall (Pattern 3): the model and the RTL agree on
`M` and nothing else. Differential testing (§8) compares them exactly there.

### 4.4 The instruction format

```text
20-bit word:   [19:15] opcode (5 bits)   [14:0] signed immediate / branch target
```

Fifteen opcodes fit in 4 bits, but keep 5 — you will add opcodes in the extensions and in
Laboratories 2–5 without re-encoding the ROM.

---

## 5. The ISA (authoritative table)

| Op | Stack effect | Notes / preconditions |
|---|---|---|
| `PUSH_S15 k` | `-- INT(k)` | sign-extend 15-bit immediate; overflow-checked capacity |
| `PUSH_TRUE` | `-- BOOL(1)` | canonical |
| `PUSH_FALSE` | `-- BOOL(0)` | canonical |
| `ADD` | `INT a, INT b -- INT(a+b)` | top is right operand; defined overflow policy |
| `SUB` | `INT a, INT b -- INT(a-b)` | |
| `MUL` | `INT a, INT b -- INT(a*b)` | low 32 bits or precise overflow — pick one, test it |
| `EQ` | `x, y -- BOOL(x==y)` | equality defined by tag policy (same tag + payload) |
| `LT` | `INT a, INT b -- BOOL(a<b)` | typed compare; non-INT -> TYPE_FAULT |
| `DUP` | `x -- x, x` | capacity checked first |
| `DROP` | `x --` | underflow checked first |
| `SWAP` | `x, y -- y, x` | |
| `JMP t` | unchanged | `pc = t` |
| `JZ t` | `BOOL c --` | branch if false; non-BOOL -> TYPE_FAULT |
| `EMIT` | `x --` | pop only on output acceptance |
| `HALT` | unchanged | enter halted state |

Required fault cases (each leaves stack and `pc` at a **precise documented state**):

```text
STACK_UNDERFLOW, STACK_OVERFLOW, TYPE_FAULT, ARITH_OVERFLOW,
BAD_OPCODE, BAD_BRANCH_TARGET, NONCANONICAL_BOOL
```

The fault record (latched; first fault wins; machine stops):

```text
code | faulting pc | opcode | depth | operand tags | optional payload fragment
```

---

## 6. Microarchitecture

### 6.1 The controller FSM

```text
RESET -> FETCH_REQ -> FETCH_WAIT -> DECODE -> EXECUTE -> COMMIT -> FETCH_REQ ...
                                   |             |
                                   |             +-> OUTPUT_WAIT -> (out_ready) -> FETCH_REQ
                                   +-> STACK_READ_REQ -> STACK_READ_WAIT -> EXECUTE
                       any state -> FAULT | HALTED
```

Rules that come from the book, not from convenience:

- **Variable cycles per instruction.** Do not force a uniform pipeline; `ADD`-from-cache and
  `ADD`-with-RAM-refill may take different numbers of cycles. Internal stuttering is legal as
  long as commit semantics are preserved.
- **A commit pulse is generated only** (a) in `COMMIT`, (b) on output acceptance in
  `OUTPUT_WAIT`, or (c) on entry to a precise fault state. The trace packet captures old and
  new `pc`, operation, depth, and visible event.
- **One mutation owner.** Compute the entire next state combinationally, assign registers in
  one `always_ff`. Never write the stack from several unrelated blocks — this is what makes
  precise faults provable instead of aspirational.

The book's canonical state-update style:

```systemverilog
always_comb begin
  pc_d    = pc_q;
  depth_d = depth_q;
  fault_d = fault_q;
  commit  = 1'b0;
  case (state_q)
    EXECUTE: begin
      // validate ALL preconditions here; move to COMMIT
      // only with complete next-state data
    end
    COMMIT: begin
      pc_d   = next_pc_q;
      commit = 1'b1;
    end
  endcase
end
```

### 6.2 Stage 1: the register stack

Start with a 16- or 32-entry register stack, no RAM at all:

```systemverilog
value40_t stack_q [0:STACK_DEPTH-1];
logic [$clog2(STACK_DEPTH+1)-1:0] depth_q;
```

**Purpose**: isolate semantic and interface bugs from RAM latency. Area is *not* the metric at
this stage; correctness of the model match, opcode encoding, fault behavior, result FIFO, and
trace comparison is. Synthesis area becomes the metric only in Stage 2.

### 6.3 Stage 2: BRAM stack with a two-entry top cache (Split-Lifetime Frame)

Once the register version passes, replace the body of the stack with synchronous block RAM and
keep a two-entry cache of the newest values:

```text
top0_q       newest value
top1_q       next value
top_count_q  0, 1, or 2
deep_count_q number of live values in RAM
deep_top_q   first free RAM address
```

The invariant that the entire refinement hinges on:

```text
architectural depth = top_count + deep_count
logical stack (newest -> oldest) = top0, top1, RAM[deep_top-1], RAM[deep_top-2], ...
```

**Push:**

```text
if top_count == 0:   top0 = v;              top_count = 1
elif top_count == 1: top1 = top0; top0 = v; top_count = 2
else:                write old top1 -> RAM[deep_top]; deep_top++; deep_count++;
                     top1 = top0; top0 = v
```

**Binary-result commit (e.g. ADD with both operands cached):**

```text
top0 = result
if deep_count == 0: top_count = 1
else:              request RAM[deep_top-1]
                   on response: top1 = response; deep_top--; deep_count--; top_count = 2
```

The abstract `ADD` commit fires when the new logical stack is *completely represented*. Holding
`result` in a register while waiting for refill is internal stuttering, not a semantic event.

**Block-RAM discipline (book ch. 6, "Block RAM discipline")**:

- Every memory wrapper exposes a **request** and a **response-valid** point. Never let design
  logic depend on combinational read data from a behavioral array — GateMate BRAM is
  synchronous, and a design that "works" in simulation with async reads fails in synthesis.
- Prefer inferred memories; instantiate a GateMate primitive only when inference cannot express
  width/mode/init/packing needs. Keep technology specifics *below* the pattern contract.
- One logical mutation owner per structure, even on dual-port RAM.

### 6.4 Output commitment — `EMIT` (Delayed Irreversible Store)

`EMIT` must **not** pop merely because `out_valid` is asserted. The channel is the
ready/valid protocol with the ownership rule (book ch. 15): *a transfer occurs on an edge where
both valid and ready are true; while blocked, the producer holds the item stable.*

```text
DECODE EMIT:
    validate depth >= 1 (and tag policy)
    pending_value = top0        // include ALL metadata
    state = OUTPUT_WAIT

OUTPUT_WAIT:
    out_valid = 1;  out_data = pending_value
    if out_ready:
        perform stack pop; pc++; emit commit packet; state = FETCH_REQ
```

Backpressure (out_ready low for 0..20+ cycles) must not change the machine state at all — that
is test case 6 in §8. Use the reference one-entry elastic register from the book (ch. 15) as
your first output stage:

```systemverilog
module rv_reg #(parameter int WIDTH = 40) (
  input  logic clk, rst,
  input  logic in_valid, output logic in_ready, input  logic [WIDTH-1:0] in_data,
  output logic out_valid, input  logic out_ready, output logic [WIDTH-1:0] out_data);
  logic full_q;  logic [WIDTH-1:0] data_q;
  assign in_ready  = !full_q || out_ready;
  assign out_valid = full_q;
  assign out_data  = data_q;
  always_ff @(posedge clk)
    if (rst) full_q <= 1'b0;
    else if (in_ready) begin full_q <= in_valid; if (in_valid) data_q <= in_data; end
endmodule
```

Behind it: a FIFO, then a UART transmitter — reuse `rtl/uart_tx.sv` from MATE-16 verbatim
(58 lines, verified on this exact board).

### 6.5 Trace and debug (book ch. 6 "Instrumentation" and ch. 18)

Capture **events, not waveforms**: a compact commit-trace record per retired instruction:

```text
sequence | pc_old | pc_new | op | depth | event (commit/fault/output) | tags
```

Compare this stream against the Python model's trace after every commit — that is the whole
differential-testing loop. On the board, either a trace RAM, the GateMate ILA, or UART streaming
works; leave BRAM headroom if you add the ILA.

---

## 7. Software side: model-first, single source of truth

The most valuable thing to steal from MATE-16 is not RTL — it is the method:

```text
tools/opcodes.py   ONE authoritative table: opcode value, mnemonic, stack effect,
                   preconditions, fault behavior. The assembler, the reference model,
                   and the RTL testbench all import (or are generated from) it.
tools/model16.py   -> your equivalent: stack_model.py, an executable reference model
tools/asm16.py     -> your equivalent: asm20.py, a two-pass assembler (no eval!)
```

Concretely, in MATE-16 (`../2026-08-25--vm-cpu-gatemate/mate16/tools/`):

- `opcodes.py` — `Instruction` dataclass (opcode, mnemonic, length, stack_in/out, effect,
  group) + a closed `Fault` IntEnum. One table prevents the failure mode where four hand-copied
  tables gradually diverge. Copy this shape for your 15 opcodes.
- `model16.py` — `class Machine` with `step()`; each opcode is a small function that checks
  preconditions and faults exactly like §4.3. Your `stack_model.py` is the same idea with
  tagged values.
- `asm16.py` — two-pass assembler producing `.hex/.bin/.lst/.sym.json`.

Your model must additionally emit the **commit trace** (old/new pc, op, depth, event) so the
RTL testbench can diff cycle-by-commit rather than eyeballing final state.

---

## 8. Verification strategy (this is most of the work)

### 8.1 The ladder

1. **Model unit tests** (pure Python, fast): every opcode at minimum stack depth; every fault
   case; both exit-criteria programs produce the exact commit traces from the book.
2. **Directed RTL tests** (iverilog + pytest, MATE-16 pattern `sim/`): golden-trace
   comparison per instruction; the book's Program A trace matched entry-by-entry.
3. **Boundary tests**: all stack boundary transitions around top-cache spill and refill
   (top_count 0/1/2, deep_count 0/1, spill-then-binary-op refills); all arithmetic boundaries;
   output stalls of length 0 through 20 cycles.
4. **Random tests**: maintain a typed software stack, choose only legal instructions, then
   occasionally inject one illegal instruction and assert the precise fault state. Compare
   after every commit.
5. **Board evidence**: the two exit-criteria programs, run on hardware.

### 8.2 Core assertions (adapt reset polarity to your convention)

```text
architectural depth == top_count + deep_count
no stack write while faulted
no pc change without commit or defined reset
blocked output remains stable (out_valid && !out_ready |=> stable out_data)
one output acceptance produces exactly one pop
committed tag is legal and canonical
```

### 8.3 Exit criteria on hardware

```text
Program A: one BOOL(true) output transfer, then HALTED
Program B: TYPE_FAULT at the ADD pc, unchanged depth and operand values, no output transfer
```

---

## 9. GateMate substrate and toolchain

### 9.1 The board and the budget

CCGM1A1: 20,480 CPEs, 40,960 FFs, 32 physical 40-Kbit block RAMs (each splittable into two
20-Kbit halves). Your **stop-build budget** (Laboratory 1):

```text
<= 2 physical RAM blocks   (e.g. 1Kx20 instruction ROM + 40-bit stack organization)
<= ~2,000 CPEs before debug instrumentation
```

A stop-build budget is a *simplify trigger*, not a forecast. If you exceed it, look first for:
register-based deep stack left enabled, wide reset logic, duplicated tag decoders, an
unexpectedly general multiplier. Then simplify the architecture — do not just accept the number.

### 9.2 Toolchain (identical to both prior projects)

OSS CAD Suite (Yosys, nextpnr-himbaechel, iverilog) + gmpack + openFPGALoader:

```bash
source ~/fpga/oss-cad-suite/environment   # every shell
make versions     # record tool versions -> build/tool-versions.txt
make sim          # iverilog testbenches
make test         # pytest suites
make asm          # assemble example programs
make synth        # yosys -> build/*.json + synthesis stats
make pnr          # nextpnr-himbaechel + gmpack -> top.bit
make load         # openFPGALoader to the board
```

Record for every experiment (book ch. 6 "Toolchain loop"): git commit, tool versions, synth
command, PnR command **and seed**, constraint file, resource report, critical path, sim seed,
model digest, hardware trace digest. Without that record, a "faster" result may be tool drift.

### 9.3 Board-level files to copy from MATE-16

- `constraints/olimex_gatematea1_evb.ccf` — verified pin map (clk_10m, fpga_but, user_led,
  UART pins to the RP2040).
- `rtl/reset_sync.sv` — async-assert / sync-release reset (PCA-Z80 reused it unchanged too).
- `rtl/uart_tx.sv` — UART transmitter with start/data handshake.
- `constraints/99-openfpgaloader.rules` — udev rule for DirtyJTAG loading.
- The `CC_USR_RSTN` config-reset cell usage in `pca_z80/rtl/top.sv` line 42 — GateMate-specific
  startup reset you will need in your `top.sv`.

---

## 10. Repository layout

Follow the book's suggested organization (ch. Appendix B), trimmed to one experiment:

```text
symbolic_eval/                 # or: keep sibling-project style, your call (see DR-1)
  Makefile
  constraints/                # CCF pins, SDC, udev rule     (copy from MATE-16)
  rtl/
    symbolic_types_pkg.sv      # value40_t, event80_t, constructors
    rv_reg.sv                 # one-entry elastic register  (book ch. 15)
    sync_sdp_ram.sv           # request/response-valid BRAM wrapper
    fault_latch.sv            # precise first-fault capture
    commit_fifo.sv            # trace/output FIFO
    stack_core.sv             # the FSM + top cache  (you write this)
    program_rom.sv
    top.sv                    # board integration
    uart_tx.sv                # copy from MATE-16
  tools/
    opcodes.py                # authoritative table (copy the MATE-16 shape)
    stack_model.py           # executable reference model + commit trace
    asm20.py                 # two-pass assembler
  programs/
    arith.asm                 # exit-criteria program A
    typefault.asm             # exit-criteria program B
    smoke.asm ...
  sim/
    tb_stack_core.sv
    test_model.py
    test_directed.py
    test_random.py
  scripts/
    synth_sys.ys
  build/                     # gitignored
```

---

## 11. Phased implementation plan

Do not start with RTL. The order below is the cheapest path to a correct machine; each phase
has a stop condition.

**P0 — Bootstrap (day 1).** Repo, Makefile, constraints, `reset_sync.sv`, placeholder `top.sv`
blinking the LED (copy PCA-Z80 Phase 0 exactly). *Stop when: `make bit && make load` blinks.*

**P1 — Types and model.** `symbolic_types_pkg.sv`; `tools/opcodes.py`; `tools/stack_model.py`
with commit trace; pytest unit tests for every opcode and every fault. *Stop when: model
reproduces the book's Program A commit trace (§ "expected commit trace") character-for-character
and Program B's TYPE_FAULT state.*

**P2 — Assembler.** `tools/asm20.py`, two-pass, no `eval`. *Stop when: `arith.asm` and
`typefault.asm` assemble to the exact hex in the book.*

**P3 — Register-stack RTL.** `stack_core.sv` with register stack, full FSM, `rv_reg` output,
fault latch. Directed tests against the model trace. *Stop when: golden-trace diff is empty for
all directed tests, including 20-cycle output stalls.*

**P4 — BRAM stack + top cache.** `sync_sdp_ram.sv`; two-entry cache; spill/refill boundary
tests; random legal/illegal-instruction tests. *Stop when: all §8 assertions hold and the
invariant `depth == top_count + deep_count` is proven across push/pop/binary-op/stall.*

**P5 — Board.** Integrate `top.sv` (ROM init from hex, output over UART), synthesize, check the
stop-build budget, load, run both programs. *Stop when: hardware evidence matches §8.3.*

**P6 — (Optional extensions).** `CALL`/`RET`, boxed integers, capability-tagged REF, replayable
multi-cycle MUL.

---

## 12. Decision records

**DR-1 — Repo: new tree vs. fork of MATE-16.** *Context*: MATE-16 already has the toolchain,
constraints, UART, and test harness. *Options*: (a) fork and modify; (b) fresh tree, copy files.
*Decision*: fresh tree with copied files, for a new intern. *Rationale*: MATE-16's core is a
16-bit untyped ISA; keeping it clean preserves a verified reference, and the copies are small
(pin files, reset, UART). *Status*: proposed.

**DR-2 — MUL overflow policy.** *Context*: book allows "low 32 bits or precise overflow."
*Decision*: precise `ARITH_OVERFLOW` on signed 32-bit overflow (wrap in the model only if you
also add an `ADD`/`SUB` wrap decision — be consistent across ADD/SUB/MUL). *Rationale*: the
laboratory's theme is precision; wrapping is the one lazy spot the book lets you have, and
taking the precise option makes the fault machinery do double duty. *Status*: proposed — the
intern must write the model test either way.

**DR-3 — EQ tag policy.** *Decision*: values are equal iff tags equal AND payloads equal
(bitwise, after canonicalization). Mixed-tag EQ yields `BOOL(false)`, not a fault (only
operations that *need* a type fault). *Status*: proposed.

**DR-4 — Reset during OUTPUT_WAIT.** *Context*: what happens to a pending EMIT on reset?
*Decision* (book's recommendation): treat reset as a global experiment abort; discard pending
output; restart from the initial image. *Status*: accepted for Lab 1.

---

## 13. Risks, traps, and failure signs

- **Combinational-read assumption.** If your testbench models RAM with async reads, RTL will
  pass sim and fail synthesis. Always request/response-valid.
- **Doubled EMIT** (a real bug class from the MATE-16 io_block: strobes derived from `req`
  instead of an acceptance edge). Pop on `out_ready` acceptance only; one acceptance = one pop.
- **Multiple mutation owners.** Any second `always_ff` writing the stack destroys precise-fault
  provability. Review for this explicitly.
- **Register deep stack left enabled** in the BRAM stage — the #1 cause of blowing the 2,000
  CPE budget.
- **ILA headroom**: the GateMate ILA consumes BRAM; budget for it or use trace streaming.
- **Tool drift**: record seeds and versions or your measurements are not experiments.

---

## 14. References

Book (ticket `sources/Composable_Hardware_Patterns_for_Symbolic_Computers.md`, by line):

- Laboratory 1 full spec: lines 4211–4569 (semantic machine, ISA, FSM, register/BRAM stacks,
  EMIT, faults, verification, GateMate exit criteria).
- Pattern contract + nine fields: ch. 3, lines 364–461.
- Commitment levels: ch. 4, lines 462–552.
- Composition laws: ch. 5, lines 553–682.
- GateMate substrate, budgets, toolchain loop: ch. 6, lines 683–807.
- Pattern catalog: ch. 7+ (Pattern 1 Abstract-Machine Contract at line 820, Pattern 2 at 869,
  Pattern 4 Common Fast Path at 967, Pattern 6 Tagged Value Word at 1084, ...).
- `symbolic_types_pkg` verbatim: lines 3397–3470. `rv_reg` verbatim: lines 3550–3610.
- Suggested project organization: lines 7228–7370. Assertion templates: lines 7473–7586.

Prior projects (file references):

- `../2026-08-25--vm-cpu-gatemate/mate16/README.md` — method overview; 165+ test discipline.
- `../2026-08-25--vm-cpu-gatemate/mate16/tools/opcodes.py` — authoritative-table shape to copy.
- `../2026-08-25--vm-cpu-gatemate/mate16/tools/model16.py` — `Machine`/`step` reference-model shape.
- `../2026-08-25--vm-cpu-gatemate/mate16/tools/asm16.py` — two-pass assembler.
- `../2026-08-25--vm-cpu-gatemate/mate16/rtl/mate16_core.sv` — multi-cycle FSM + precise faults.
- `../2026-08-25--vm-cpu-gatemate/mate16/rtl/io_block.sv` — held-request bus, acceptance-edge
  strobes (the doubled-UART-byte lesson).
- `../2026-08-25--vm-cpu-gatemate/mate16/rtl/{reset_sync,uart_tx,top}.sv`, `constraints/`.
- `../2026-08-28--pca-gatemate/pca_z80/README.md` — phase-driven build-up, toolchain install.
- `../2026-08-28--pca-gatemate/pca_z80/rtl/top.sv` — `CC_USR_RSTN` startup reset usage.

---

## 15. Glossary (minimum)

- **value40** — 40-bit tagged word: tag[3:0], flags[3:0], payload[31:0].
- **commit pulse** — the single cycle on which architectural state legally changes.
- **precise fault** — fault state in which pc and stack equal their pre-instruction values.
- **top cache** — top0/top1 registers over the BRAM stack body (Split-Lifetime Frame).
- **stop-build budget** — resource threshold that triggers simplification, not a prediction.
- **stuttering** — internal extra cycles that do not correspond to any semantic event.
- **differential testing** — comparing RTL commit traces against the reference model.

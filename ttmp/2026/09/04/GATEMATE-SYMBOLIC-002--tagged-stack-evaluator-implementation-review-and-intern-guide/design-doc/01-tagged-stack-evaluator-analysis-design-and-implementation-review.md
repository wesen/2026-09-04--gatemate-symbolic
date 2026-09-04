---
Title: Tagged stack evaluator analysis design and implementation review
Ticket: GATEMATE-SYMBOLIC-002
Status: review
Topics:
    - fpga
    - gatemate
    - symbolic-computers
    - architecture
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: repo://symbolic_eval/rtl/stack_core.sv
      Note: Register-stack staging and post-EMIT depth findings
    - Path: repo://symbolic_eval/rtl/stack_core_bram.sv
      Note: Cache representation, DUP and fault-context findings
    - Path: repo://symbolic_eval/rtl/top.sv
      Note: Board integration and output ownership
    - Path: repo://symbolic_eval/sim/test_bram.py
      Note: Verification coverage and random generation review
    - Path: repo://symbolic_eval/tools/asm20.py
      Note: Two-pass assembler and validation findings
    - Path: repo://symbolic_eval/tools/opcodes.py
      Note: ISA and fault metadata authority
    - Path: repo://symbolic_eval/tools/stack_model.py
      Note: Architectural semantics and differential oracle
ExternalSources: []
Summary: Intern guide to the implemented machine, reproduced correctness findings, and a phased remediation design.
LastUpdated: 2026-09-04T15:45:00-04:00
WhatFor: Understand and review the implemented evaluator before extending its instruction set or hardware.
WhenToUse: Onboarding, architecture review, correctness remediation, and planning the next implementation ticket.
---


# Tagged stack evaluator: analysis, design, and implementation review

## 1. Review conclusion and reading guide

This repository implements a small programmable processor for tagged values on an Olimex GateMateA1-EVB FPGA. Its strongest architectural idea is to separate the meaning of an instruction from the physical work required to execute it. Python describes the meaning; two SystemVerilog cores implement it with different stack storage; a common trace format compares their retired instructions. The board design combines the BRAM core with an instruction ROM, a buffered output channel, a hexadecimal formatter, and a UART transmitter.

The implemented scope is substantially more than a sketch: it includes 17 instructions, explicit faults, subroutine calls, recursive Fibonacci, a two-pass assembler, a 123-test suite, synthesis scripts, and documented earlier board demonstrations. This review reran all 123 tests successfully and synthesized the current Fibonacci design. It also reproduced correctness defects that the existing suite does not expose. The design is a useful working laboratory, but its broad claim that both cores implement the same precise machine needs qualification until those defects are addressed.

The most significant findings are BRAM `DUP` accepting an empty stack; the register core restoring an obsolete stack depth after `EMIT` followed by `JMP` or `SWAP`; and both cores wrapping the program counter at the ROM boundary while the model faults. Additional findings concern wrong BRAM fault metadata, return depth becoming visible before retirement, assembler validation, and gaps in the verification strategy. Section 10 contains inputs, evidence, severity, and concrete remediation guidance.

For a new intern, read sections 2–4 to understand the machine, sections 5–8 to understand how it runs, section 9 to understand what the tests establish, and sections 10–12 before changing code. Section 13 is the practical reproduction and learning sequence. Section 14 is the API and source-reference index. Diagrams use fixed-width text so the same document remains readable in Markdown and on reMarkable.

### Scope and evidence vocabulary

The reviewed implementation is commit `c7f9dc9`. New review scripts and documentation are separate changes; no production RTL, Python tool, or existing test was modified during this review. Paths beginning `rtl/`, `sim/`, `tools/`, `programs/`, or `playbooks/` are relative to `symbolic_eval/`. Line numbers refer to the reviewed implementation.

- **Observed now:** code inspection, the 123-test rerun, targeted simulations, assembler probes, and Yosys synthesis recorded in this ticket.
- **Historical evidence:** prior commits, README measurements, and the original ticket diary's board observations. These are attributed rather than described as fresh measurements.
- **Proposed:** remediation and extension design in this report. It is not implemented by this review.
- **Not established:** full formal equivalence, complete fault coverage in RTL, renewed routed timing, or new physical-board behavior.

## 2. What problem the project solves

A conventional integer processor knows that a register contains bits, but often leaves the interpretation of those bits to software. A tagged processor carries a small type identifier beside each payload. Here, an `INT(4)` and a `BOOL(true)` are distinct 40-bit values. `ADD` accepts two integer operands and faults on a Boolean operand; `JZ` requires a canonical Boolean condition. This makes type rules and fault boundaries explicit in hardware.

The machine is stack based. An instruction such as `ADD` has no register operands encoded in its instruction word: it takes the two newest stack values and replaces them with their sum. This keeps instructions compact and makes an executable reference model straightforward. It also makes stack storage, underflow, overflow, and preservation of values across function calls central design concerns.

This is Laboratory 1 of the locally archived *Composable Hardware Patterns for Symbolic Computers*. It is not yet a heap-managing symbolic language runtime. `REF`, `PAIR`, `THUNK`, and related tags reserve a vocabulary for later work, but there is no heap allocator, graph reducer, garbage collector, reference capability checker, or host command protocol in the current implementation. The implemented value-producing instructions construct integers and Booleans; stack manipulation and output move packed values without implementing those future object semantics.

The project is also an experiment in hardware methodology: first define the architectural state, then define observable commitment, then change the implementation while trying to preserve behavior. The register core supplies a simple representation; the BRAM core supplies a more economical physical representation for a deeper stack.

### Implementation history

| Stage | Commit | Concrete contribution |
|---|---|---|
| Design research | `5d46970`, `95ee8ea`, `3cbb775` | Source import, pattern study, initial intern guide |
| P0 | `e9f9cfc` | Clock/reset/blink board bootstrap |
| P1 | `7c4b6b5` | Instruction metadata and executable Python machine |
| P2 | `dc8a3dc` | Assembler and initial programs |
| P3 | `e89523e` | Register-stack RTL and trace comparison |
| P4 | `3e8dba8` | BRAM core, cache occupancy invariant, random tests |
| P5 | `f677d88` | Board wrapper, synthesis, recorded board demonstrations |
| P6 | `e795826` | README and implementation wrap-up |
| Extensions | `51dfc0a`, `9858b2a` | LED polarity correction, CALL/RET, recursive programs |
| Operational documentation | `c7f9dc9` | Five playbooks for extending and debugging the system |

The earlier documentation ticket was closed before several implementation commits. Therefore ticket status alone is not a reliable chronological record of which hardware features existed. Git history and the original diary's later steps supply that context.

## 3. System map: host tools, processor, and physical output

```text
HOST / DEVELOPMENT MACHINE

 programs/*.asm
       |
       v
 tools/asm20.py <------ tools/opcodes.py
       |                       |
       | words                 v
       |               tools/stack_model.py
       |                       |
       |                 expected TRACE/FINAL
       v                       |
 build/*.hex                   | compare in pytest
       |                       |
       +--> Icarus simulation -+--> observed TRACE/FINAL
       |
       +--> Yosys --> nextpnr --> gmpack --> FPGA bitstream

FPGA / BOARD

  program_rom --20-bit instruction--> stack_core_bram
       ^                                   |
       | PC address                        | value40 ready/valid
       +-----------------------------------+       |
                                                   v
 deep RAM <--> top cache                    rv_reg (one slot)
                  |                                |
             ALU + control                         v
                  |                         value printer
             return stack                          |
                                                   v
                                               uart_tx --> host

 configuration reset + button --> reset_sync --> machine and output
 halted / fault_valid -------------------------> status LED
```

The testbench substitutes a synchronous ROM and an output consumer for the board's peripherals. The core is consequently testable without UART timing. Board tests then exercise the integrated path by decoding the serial pin into bytes. This separation is valuable because an arithmetic mismatch and a byte-framing error occur at different layers and require different evidence.

The FPGA synthesizes only the BRAM implementation through `scripts/synth.ys`. The register implementation is a useful comparison target, not an alternate board selection exposed by the current build. `TOP` and `RTL_SYNTH` in the Makefile are not what selects the synthesis hierarchy: the `.ys` file explicitly invokes `synth_gatemate -top top` and lists its input files.

## 4. Architectural contract

### 4.1 State and representation

For the current CALL/RET machine, use this abstract state:

```text
M = <pc, data_stack, return_stack, output, fault_record, halted>

stack order: [oldest, ..., next-to-top, top]
                                    a     b
binary instruction: [..., a, b] -> [..., result(a,b)]
```

The older module comments omit the return stack from their state tuple, but `Machine.rstack` and both cores implement it. Architectural state is the state a program can depend on. Pipeline registers, pending memory reads, and cached representations are implementation state. A valid implementation can spend several clock cycles on one instruction without performing another abstract instruction; those internal cycles are often called stuttering steps.

The packed value is defined by `Value40.word()` in `tools/stack_model.py:57` and `value40_t` in `rtl/symbolic_types_pkg.sv:26`:

```text
bit       39       36 35       32 31                         0
          +---------+-----------+----------------------------+
value40   | tag (4) | flags (4) | payload (32)               |
          +---------+-----------+----------------------------+

word = (tag << 36) | (flags << 32) | payload
INT(-1):  0x00ffffffff
BOOL(1):  0x1000000001
```

`INT` payloads are interpreted as signed two's-complement numbers by arithmetic instructions. The Python value object stores the payload as an unsigned 32-bit pattern; `_s32()` interprets it as signed when needed. Thus an output payload of `ffffffff` can represent integer -1 even though its Python stored payload is 4294967295.

Flags are part of full-value equality. Existing constructors set them to zero; the arithmetic type checks test tags, not flag canonicality. `JZ` validates the Boolean payload as zero or one but does not reject nonzero flags. An extension must therefore specify flag semantics rather than assuming that every operation already enforces a zero-flags rule.

### 4.2 Instruction word and complete ISA

Each instruction is 20 bits: opcode in bits 19:15 and immediate in bits 14:0. `PUSH_S15` interprets that immediate as signed, with range -16384 through 16383. Branches and calls interpret it as an unsigned absolute word address. The default ROM has only 1024 words, so most encodable branch addresses are outside the implemented memory.

| Hex opcode | Instruction | Data-stack effect | Additional rule |
|---|---|---|---|
| 00 | PUSH_S15 k | `S -> S, INT(k)` | Sign-extend 15-bit immediate |
| 01 | PUSH_TRUE | `S -> S, BOOL(1)` | Canonical true |
| 02 | PUSH_FALSE | `S -> S, BOOL(0)` | Canonical false |
| 03 | ADD | `S,a,b -> S,a+b` | Both INT, signed 32-bit result required |
| 04 | SUB | `S,a,b -> S,a-b` | Top is the right operand |
| 05 | MUL | `S,a,b -> S,a*b` | Both INT, signed overflow faults |
| 06 | EQ | `S,a,b -> S,BOOL(a==b)` | Compare tag, flags, payload; mixed tags yield false |
| 07 | LT | `S,a,b -> S,BOOL(a<b)` | Signed integer comparison |
| 08 | DUP | `S,a -> S,a,a` | Requires one value and spare capacity |
| 09 | DROP | `S,a -> S` | Requires one value |
| 0A | SWAP | `S,a,b -> S,b,a` | Requires two values |
| 0B | JMP t | unchanged | Validate t, then set pc=t |
| 0C | JZ t | `S,BOOL(c) -> S` | Validate t even when c is true; jump when c=0 |
| 0D | EMIT | `S,a -> S` | Transfer a and pop at output acceptance |
| 0E | HALT | unchanged | Halt; PC remains at HALT |
| 0F | CALL t | unchanged | Push pc+1 on return stack; jump to t |
| 10 | RET | unchanged | Pop return stack into PC |
| 11–1F | undefined | unchanged on fault | BAD_OPCODE |

The metadata lives in `tools/opcodes.py:81`. The Python assembler and model import it. RTL constants and testbench name tables are handwritten mirrors; no package generator exists in the inspected repository despite the metadata module's opening comment mentioning one. Treat “single source of truth” as the intended authority, not an enforced generation pipeline.

### 4.3 Fault semantics and check order

A precise fault retains the faulting PC, both logical stacks, and the prior output stream. It adds a fault record and stops execution. It does not mean every bit of machine state stays unchanged: fault metadata, trace sequence, and the stopped-state control obviously change.

| Code | Fault | Trigger |
|---|---|---|
| 1 | STACK_UNDERFLOW | Insufficient data operands |
| 2 | STACK_OVERFLOW | A data-stack push exceeds capacity |
| 3 | TYPE_FAULT | Operand tag violates the instruction's requirements |
| 4 | ARITH_OVERFLOW | Signed result is outside [-2^31, 2^31-1] |
| 5 | BAD_OPCODE | Undefined opcode |
| 6 | BAD_BRANCH_TARGET | Invalid target; model also uses it for escaped fetch |
| 7 | NONCANONICAL_BOOL | JZ sees BOOL with payload other than 0 or 1 |
| 8 | RSTACK_UNDERFLOW | RET on an empty return stack |
| 9 | RSTACK_OVERFLOW | CALL on a full return stack |

Capacity checks precede operand interpretation. `JZ` checks underflow, tag, Boolean payload, and target, in that order. `CALL` checks return capacity before its target. This order is observable if one instruction violates multiple requirements. For example, an invalid-target JZ on an integer faults with TYPE_FAULT, and a full-return-stack CALL with an invalid target faults with RSTACK_OVERFLOW.

The fault record includes opcode, PC, data depth, and top-two operand tags. An absent operand contributes tag zero. This last detail is relevant to finding F4: the BRAM implementation must recover the true logical second tag when that operand is in RAM.

### 4.4 Concrete arithmetic example

`programs/arith.asm` computes `((7+5)*3)==36`. These are abstract instruction transitions, not single FPGA clock cycles:

```text
pc  instruction     stack after retirement
0   PUSH_S15 7      [INT(7)]
1   PUSH_S15 5      [INT(7), INT(5)]
2   ADD             [INT(12)]
3   PUSH_S15 3      [INT(12), INT(3)]
4   MUL             [INT(36)]
5   PUSH_S15 36     [INT(36), INT(36)]
6   EQ              [BOOL(1)]
7   EMIT            []             output accepts BOOL(1)
8   HALT            []             pc remains 8
```

By contrast, `PUSH_TRUE; PUSH_S15 4; ADD` stops at PC 2 with `[BOOL(1), INT(4)]` intact and no emitted output. The model test checks those actual values; the core differential trace checks the corresponding depth, PC, fault code, and operand tags.

## 5. Host-side implementation: assembler and executable model

### 5.1 Assembler API and artifacts

`assemble(source: str, rom_depth: int = 1024)` in `tools/asm20.py:49` returns `(words, symbols, listing)`. It accepts semicolon comments, labels, case-insensitive mnemonics, integer operands, and the raw `WORD` directive. Labels identify instruction-word addresses, not byte offsets. `WORD 0xF8000` encodes an undefined opcode; `WORD 0x1F` does not, because the opcode occupies the high five bits.

Pass one assigns label addresses and appends either encoded words or placeholders. Pass two resolves each immediate from a label or integer and performs immediate-specific range checks. Listing positions are stored separately from word positions because comment and label-only lines have listing entries without generating words. That is a good small implementation detail: using instruction index as listing index would patch the wrong listing entry.

```text
assemble(source, ROM_DEPTH):
    words = []; symbols = {}; pending = []
    for line in source:
        parse optional label, opcode, operand
        if label: bind it to len(words), rejecting duplicates
        if instruction needs an immediate:
            remember (word_index, listing_index, operand)
            append placeholder
        else:
            encode instruction or raw WORD
    for unresolved immediate:
        resolve label or integer
        validate signed/unsigned and ROM limits
        patch words[word_index] and listing[listing_index]
    return words, symbols, listing
```

This sketch describes the intended flow; malformed-line handling and the raw-word capacity path have defects described in F6 and F7. The CLI `main()` at line 137 writes `.hex`, `.lst`, and `.sym.json`. Only the hex writer pads to ROM depth. The library returns just actual words. Despite the Makefile comment mentioning `.bin`, the CLI does not generate a binary file.

An uninitialized program tail is represented as zero words, and zero is the legal `PUSH_S15 0` instruction. Falling beyond assembled source therefore continues executing pushes while inside ROM; it is not automatically a HALT or BAD_OPCODE. This is an intentional semantic choice and must be understood when writing tests.

### 5.2 Model API and ownership

`Machine` in `tools/stack_model.py:138` stores the entire abstract machine. `step()` either returns one `TraceRecord`, or `None` when already halted/faulted. `run()` iterates with a step limit, and `run_program()` constructs and returns a machine. Reaching the step limit does not itself raise a timeout error or set a fault, so callers must inspect whether execution actually terminated.

The ordinary `_commit()` path consumes `stack_in`, appends the new values, updates PC, and emits a trace. `EMIT`, `HALT`, CALL, and RET have some state changes in their individual cases, so “single mutation owner” is a conceptual rule rather than one Python function owning literally every field assignment. Each Python step still completes atomically from its caller's perspective.

```python
# Run from symbolic_eval/ after making tools importable.
from asm20 import assemble
from stack_model import run_program

words, labels, listing = assemble("PUSH_S15 3\nDUP\nMUL\nEMIT\nHALT\n")
machine = run_program(words, total_depth=32)
assert machine.halted and machine.fault is None
assert machine.output[0].payload == 9
```

The model is an oracle, not an independent mathematical proof. If the model and RTL share the same misunderstanding, trace comparison can pass. The fixed expected trace for the book arithmetic program provides one independent example, but broader specification-derived properties remain necessary.

## 6. RTL execution and the register-stack implementation

### 6.1 Signals and clocked state

A core consumes `clk`, active-low `rst_n`, and the ROM's 20-bit data. It drives the ROM address continuously from PC. There is no fetch-ready signal: the core assumes exactly the synchronous read timing implemented by the ROM wrapper. Its output channel is ready/valid, its trace is a pulse plus fields with no backpressure, and its fault record remains latched once faulted.

Names ending `_q` are registers holding current state. Names ending `_d` are combinationally calculated values to load at the next clock edge. Assigning `rdepth_d` in EXECUTE is therefore an architectural change at that edge if `rdepth_q <= rdepth_d` runs unconditionally. It does not become “staged” merely because the name ends in `_d`; F5 demonstrates this distinction.

The usual register-core flow is:

```text
RESET --> FETCH --> FETCH_WAIT --> DECODE --> EXECUTE
             ^                                 |
             |         checks pass             v
             +----------------------------- COMMIT
             |                                 |
             |                            HALTED if HALT
             |
             +----- OUTPUT_WAIT <--------------+ EMIT
                      |       |
               ready=0|       | ready=1: output/pop/trace
                      +-------+

EXECUTE --failed check--> FAULT (terminal until reset)
```

FETCH and FETCH_WAIT accommodate the synchronous instruction memory. EXECUTE computes and validates; COMMIT publishes a normal instruction. EMIT uses OUTPUT_WAIT as its separate retirement point. Therefore the precise statement is “normal COMMIT or EMIT acceptance owns retirement,” not that the COMMIT state is the only place any architectural mutation occurs.

### 6.2 Register stack

`rtl/stack_core.sv:18` has defaults ROM_DEPTH=1024, STACK_DEPTH=32, RSTACK_DEPTH=16. Its stack is an array of packed values; `stack_q[depth-1]` is top0 and `stack_q[depth-2]` is top1. Reads are guarded by depth. A binary operation stages a write to `depth-2` and stages `depth-1` as the new depth. SWAP stages two writes. Actual array writes occur in the clocked block only while in COMMIT.

Reset clears occupancy rather than clearing every stack word. This is valid because values beyond live depth are inaccessible. It also illustrates why stale occupancy or an absent underflow check is a correctness problem: stale physical data becomes visible if the logical validity bookkeeping is wrong.

The ALU computes signed 64-bit ADD, SUB, and MUL intermediates. Overflow is detected when bits 63:32 are not a sign extension of bit 31. An overflowing result is not truncated and committed; it becomes ARITH_OVERFLOW. `EQ` compares the whole 40-bit packed value, while LT compares signed payloads after checking INT tags.

### 6.3 CALL/RET and recursive stack discipline

The return stack holds addresses, not tagged values. CALL saves the next instruction address and redirects control; RET restores an address. The return stack is separate because an ordinary data operation should not accidentally consume a continuation. This separation does not supply local-variable frames or automatic argument management.

`programs/fib.asm` implements `fib(n)` with the data convention `(n -- fib(n))`. The base case retains n and returns. In the recursive case, the caller keeps n underneath the first recursive argument, then swaps the first result with n before computing the second argument. The algorithm's data stack preserves values across calls; the return stack preserves where execution resumes.

```text
fib entry:                 [..., n]
prepare first recursive:   [..., n, n-2]
after first CALL:          [..., n, fib(n-2)]
SWAP and subtract one:     [..., fib(n-2), n-1]
after second CALL:         [..., fib(n-2), fib(n-1)]
ADD and RET:               [..., fib(n)]
```

This program is useful integration coverage because it repeatedly combines arithmetic, comparisons, branches, stack rearrangement, calls, and returns. It is still one family of paths. Passing fib(10)=55 does not test CALL at the last ROM address or return-depth publication between clock edges.

## 7. BRAM storage and refinement

### 7.1 Why the top cache exists

A block RAM is compact storage with synchronous reads. Arithmetic wants the newest two stack values immediately, while most older values are idle. The BRAM design keeps those hot values in registers and puts older ones into `sync_sdp_ram`. This is the split-lifetime idea: values with different access patterns live in different physical storage while retaining one logical ordering.

The board configuration has 512 deep words plus two cache slots, for a maximum logical depth of 514. Comparison tests commonly set DEEP_DEPTH=30 to match a 32-entry register stack. The cores implement the same instruction semantics only when configured with corresponding capacities; their defaults do not have equal overflow thresholds.

```text
logical oldest                                  logical newest
+---------+---------+-----+-------------+--------+--------+
| RAM[0]  | RAM[1]  | ... | RAM[dc-1]   | top1   | top0   |
+---------+---------+-----+-------------+--------+--------+
              dc deep values                 tc=2

A valid alternative after a pop:
+---------+---------+-----+-------------+--------+
| RAM[0]  | RAM[1]  | ... | RAM[dc-1]   | top0   |
+---------+---------+-----+-------------+--------+
                                              tc=1
```

Do not assume `tc=min(depth,2)`. A pop can deliberately leave one cached value and a nonempty deep region. The full abstraction is `RAM[0:dc]` followed by the live cache values in oldest-to-newest order. The count invariant is `depth=dc+tc`; a stronger refinement invariant also requires every reconstructed value to equal the model's value at that position.

### 7.2 Spill, pop, and binary result installation

A push with two cached values spills old top1 to RAM[dc], shifts top0 into top1, installs the new value in top0, and increments dc and depth. The write happens at COMMIT. A push with fewer cached values needs no RAM write.

A pop with two cached values shifts top1 into top0 and reduces tc to one. A pop with one cached value and nonempty RAM fetches RAM[dc-1] into a pending refill register and installs it when the instruction retires. Popping the last value makes tc zero. These updates preserve order, not just item count.

For a binary instruction with two cached operands and a nonempty deep region, the core can read the next deeper value before execution and use it as the new top1. With one cached operand, it must first read the second operand from RAM. If another older value exists, a second read obtains the value that will become top1 after the binary result is installed.

```text
binary operation, initially tc=1 and dc>=2:

DECODE       request RAM[dc-1] as operand a
RDWAIT       capture a in opnd_q
EXECUTE      validate a and top0; compute result
             request RAM[dc-2] for later top1
RDWAIT       capture refill in rf_q
COMMIT       top0=result; top1=refill
             tc=2; dc=old_dc-2; depth=old_depth-1
```

The read-purpose enum in `rtl/stack_core_bram.sv:81` distinguishes operand fetch (`R_OPND`), binary refill (`R_FILL`), pop refill (`R_POP`), and the second refill (`R_FILL2`). Reads are side-effect-free. They can occur before a type check without changing architectural values, provided their results are not installed until commitment.

### 7.3 Memory contract and physical mapping

`sync_sdp_ram` at `rtl/sync_sdp_ram.sv:9` has one write port and one synchronous read port. A rising edge samples `wr_en`, write address/data, and read address. Read data is registered. There is no read enable, transaction identifier, or returned valid signal; controller states encode when the response can be used.

The comment says the design never reads and writes the same address in one cycle. That is too strong: a spill into an empty deep region can use address zero for both ports because the default read address is zero. The relevant implementation property is that a collision's returned read value is not consumed as an operand in that cycle sequence. An intern should reason about consumption timing rather than relying on the comment.

Fresh Yosys synthesis maps the Fibonacci board design to two `CC_BRAM_20K` cells and one `CC_MULT`. It reports 495 `CC_L2T4` cells; these are not interchangeable with the earlier routed/packed CPE count of 392. The GateMate synthesis flow includes technology-specific memory mapping; its official command documentation is archived in `sources/yosys-synth-gatemate.md`. See [Yosys synth_gatemate reference](https://yosyshq.readthedocs.io/projects/yosys/en/v0.50/cmd/synth_gatemate.html).

## 8. Output ownership, reset, and board integration

### 8.1 EMIT acceptance is the program-visible event

The core snapshots the value to emit in `pending_q`, then asserts out_valid until out_ready is true. During a stall, both the pending data and validity must remain stable. On the accepting edge the core pops its data stack, advances PC, and emits an OUTPUT trace.

The accepting consumer on the board is `rv_reg`, not the physical serial wire. Thus a program can retire EMIT and HALT while its bytes are still being transmitted. This is intentional buffering, and `sim/tb_top.sv` explicitly drains after halt/fault to avoid truncating the observed stream.

```text
edge A: core valid & rv ready
        ownership of value moves core -> elastic register
        EMIT retires; core may execute subsequent instructions

edge B: rv valid & printer ready
        ownership moves elastic register -> printer.vval_q
        elastic register may accept another value

later:  thirteen bytes leave uart_tx one at a time
```

`rv_reg` implements `in_ready = !full_q || out_ready`, allowing a consumed slot to be replaced on the same edge. It is not a combinational fall-through buffer: output data comes from its register. The printer has its own value register, so one value can be printing while another waits in the elastic slot.

### 8.2 UART API and throughput

`top.byte_of()` at `rtl/top.sv:161` maps a value to 13 bytes: `T`, one tag hex digit, `:`, eight payload hex digits, CR, LF. Flags are omitted. This is a human-readable observation format rather than a lossless value40 transport for arbitrary future flag-bearing values.

`uart_tx` accepts an 8-bit byte when start is high and ready is high. It emits a start bit, eight least-significant-bit-first data bits, and a stop bit. The divider rounds 10,000,000/115,200 to 87 clocks per bit, giving roughly 114,943 baud. A 13-byte value needs at least 130 serial bit times, about 1.131 ms, before small control gaps. These numbers are derived from `rtl/uart_tx.sv:19`, not newly measured wire timing.

The printer states are V_IDLE, V_SEND, and V_LAST. V_IDLE takes ownership of a value; V_SEND issues each byte when the UART is ready; V_LAST waits for the final byte to finish. The UART RX pin is unused, so there is no serial program loader or interactive monitor in this implementation.

### 8.3 Reset and LED behavior

`CC_USR_RSTN` supplies configuration reset. The external button first passes through two clocked synchronizer stages, then combines with configuration reset before `reset_sync`. The reset synchronizer asserts its output asynchronously when its input reset asserts and releases it through two clock stages. The button-to-core path nevertheless includes the button synchronizer, so describing the physical button path itself as immediately asynchronous is imprecise.

Reset is a global experiment abort. It clears occupancy and control in the core, elastic register, formatter, and transmitter. It can interrupt already-accepted output bytes. No exactly-once persistence across reset is promised. There is no dedicated button debounce logic, and current board tests keep the button unpressed; restart under output load is a missing test, not a demonstrated reset failure.

The LED is active-low. Halt drives the logical LED on and therefore the pin low. Fault and running select different counter bits before inversion. At 10 MHz, bits 21 and 23 yield approximately 2.38 Hz and 0.596 Hz. A halted LED does not prove that the UART has drained. A fault LED does not reveal the exact fault code or unchanged stack contents: the board top leaves detailed trace and fault ports unconnected.

The official board product page is archived as context for the board identity and hardware documentation links. Actual pin assignments used here are in the repository CCF, and prior polarity corrections are recorded in Git. See [Olimex GateMateA1-EVB](https://www.olimex.com/Products/FPGA/GateMate/GateMateA1-EVB/open-source-hardware).

## 9. Verification: what passed and what remains unproven

### 9.1 Existing suite, rerun during this review

`reference/validation/baseline-tests.log` records `123 passed in 6.11s`. Collection in `test-inventory.log` gives the current breakdown:

| Suite | Collected tests | What it observes |
|---|---:|---|
| test_model.py | 36 | Abstract semantics, selected constructed states, value packing, CALL/RET |
| test_assembler.py | 12 | Encoding, labels, selected diagnostics, program assembly |
| test_directed.py | 17 | Register core on 16 programs plus book exit criteria |
| test_bram.py | 52 | 16 BRAM programs, 24 random seeds, 12 register random seeds |
| test_top.py | 6 | Integrated UART byte streams and final debug state |
| Total | 123 | Several tests contain multiple program/stall runs |

The README's headline count is current, but its per-file counts and `make test` comment include older numbers. The authoritative fresh inventory is pytest collection, not the narrative count.

The core harnesses compare every printed TRACE and FINAL line against the model. TRACE contains sequence, old/new PC, mnemonic, data-stack depth, and event. FAULT adds code and two tags. OUTPUT adds tag and payload. FINAL contains halted, fault name, PC, data depth, output count, and return depth.

```text
TRACE 7 7 8 EMIT 0 OUTPUT 1 00000001
      | | | |    | |      | |
      | | | |    | |      | payload (hex)
      | | | |    | |      tag
      | | | |    | event
      | | | |    post-instruction data depth
      | | | opcode name
      | | new PC
      | old PC
      sequence

FINAL 1 NONE 8 0 1 0
      | |    | | | return depth
      | |    | | accepted output count
      | |    | data depth
      | |    PC
      | fault name
      halted
```

The benches also check blocked output stability, PC changes only with retirement, and BRAM `depth=tc+dc`. These are useful local properties. They do not compare all data values after each instruction, do not inspect live return addresses, do not require return depth to change only on retirement, and do not prove fault preservation of the entire stack. The new probes all reach the harness's `TB_PASS`, including the cases where the separately compared model disagrees.

### 9.2 Random generation limitations

`sim/test_bram.py:99` tracks a software stack of tags when selecting instructions. Its generator is not a complete typed-program oracle:

- After LT it tracks INT even though LT produces BOOL; only EQ is marked BOOL in that branch.
- The “nop” path appends a PUSH before checking whether capacity allows it.
- An injected ADD or DROP need not be illegal in the generated state, yet generation terminates immediately afterward.
- A taken JZ can land just beyond the generated words, where zero padding executes PUSH_S15 0 rather than a deliberately generated continuation.
- Random generation does not include CALL/RET.

These issues do not invalidate a successful model-versus-RTL comparison for the actual emitted words. They weaken claims about how much legal execution, branch coverage, and intentional fault coverage the random suite provides. The short programs can terminate in an unintended type fault or padding overflow well before the intended length.

### 9.3 Evidence ladder and build reproducibility

The prior diary reports `T1:00000001` for the arithmetic board program, no UART output plus fault indication for the type-error program, and `T0:00000037` for Fibonacci. It reports 392 packed CPEs and 16.56 MHz routed maximum frequency after CALL/RET. Those are historical observations from the original implementation session. This review did not load the FPGA, rerun place and route, capture UART hardware bytes, or inspect live fault state.

Fresh `make synth PROG=fib` completed with two BRAM cells, one multiplier, and 495 CC_L2T4 cells. The log also reports replacement of the small return-address memory with registers and a ROM port resize warning. Neither warning is treated as proof of a runtime defect; synthesis succeeded and the mapped cell counts are archived. The full Yosys log should accompany the summary because resource categories from synthesis and routing differ.

The tool versions are OSS CAD Suite 20260825, Yosys 0.68+130, Icarus 14.0 development snapshot, nextpnr 0.11.1 development revision, openFPGALoader 1.1.1, and Python 3.11.6. A read-only fonts configuration warning appeared in the nextpnr version wrapper; no new routing result was obtained. The versioned Yosys web reference is background documentation, not a claim that the installed compiler is version 0.50.

## 10. Prioritized implementation findings

Severity here describes correctness and engineering impact, not an internet-facing security rating. F1–F3 and F6 should be addressed before relying on arbitrary assembled programs. Each result below was observed at the reviewed commit unless marked as code-inspection evidence.

### F1 — High: BRAM DUP accepts an empty stack

**Location:** `rtl/stack_core_bram.sv:297`, specifically the shared PUSH/DUP case. The code checks total capacity but never checks whether DUP has an operand. By contrast, `rtl/stack_core.sv` handles DUP separately and rejects depth zero.

**Reproducer:** `DUP; HALT`, starting from reset.

```text
expected: FINAL 0 STACK_UNDERFLOW 0 0 0 0
BRAM:     FINAL 1 NONE 1 1 0 0
```

The BRAM core reads the physically present top0 register despite there being no live value. At reset this register is zero, so the machine manufactures an integer-like zero and successfully halts. After earlier activity, the stale value can differ. The occupancy invariant still passes because both the cached count and logical depth were incremented consistently.

**Proposed fix:** In the common case, check `op==DUP && depth_q==0` before the capacity check; then retain the existing push/spill logic. **Acceptance:** empty DUP faults on both cores; DUP after consuming the last item also faults; ordinary and full-capacity DUP retain their existing behavior. Evidence: `stack_core_bram-empty-dup.diff` and `probe-results.json`.

### F2 — High: register-core JMP and SWAP reuse stale staged depth after EMIT

**Locations:** `rtl/stack_core.sv:299` (SWAP), line 314 (JMP), line 373 (COMMIT), and line 394 (OUTPUT_WAIT).

EMIT directly decrements `depth_q` through `depth_d`, but does not update staged `ndepth_q`. The next JMP or SWAP stages a new PC without assigning `ndepth_d`. COMMIT then reinstalls the obsolete staged depth. Instructions that explicitly set ndepth, such as HALT, mask this problem when they immediately follow EMIT.

```text
PUSH_S15 7
EMIT
JMP end
end: HALT

expected FINAL: 1 NONE 3 0 1 0
register FINAL: 1 NONE 3 1 1 0
```

A second probe uses three pushes, EMIT, SWAP, two further EMITs, and HALT. It also diverges; the corrupt depth can expose old physical data as live values, so this is more than a cosmetic trace error. The BRAM core passes both probes because these cases stage its current counts explicitly.

**Proposed fix:** Every successful instruction must build a complete retirement candidate from current architectural state. At minimum, JMP and SWAP must assign the current depth explicitly. Prefer systematic initialization of all candidate fields at instruction execution rather than relying on the previous instruction's staging registers. **Acceptance:** enumerate neutral-stack instructions following EMIT and assert both depth and full remaining values, including under randomized output stalls. Evidence: `stack_core-emit-jmp.diff` and `stack_core-emit-swap.diff`.

### F3 — High: RTL PC wraps at ROM end instead of matching the model's fetch fault

**Locations:** PC and staged-PC widths in both core declarations; sequential `pc_q + 1'b1` assignments; `Machine.step()` fetch bound check in `tools/stack_model.py` near line 211.

Both default RTL PCs are 10 bits, so incrementing 1023 produces zero. The model uses a Python integer and faults on the subsequent fetch at 1024. The probe fills a 1024-word ROM, places JMP 1023 at address zero and PUSH_S15 7 at address 1023, and then runs both cores.

```text
model:    FINAL 0 BAD_BRANCH_TARGET 1024 1 0 0
both RTL: FINAL 0 STACK_OVERFLOW 1023 32 0 0
```

Hardware repeatedly wraps, jumps, and pushes until capacity is exhausted. The mismatch occurs at the first boundary transition, long before the final overflow. The existing test named `test_pc_escape_past_rom_is_bad_branch_target` checks an explicit invalid JMP, not sequential fallthrough from the last valid word.

CALL return addresses have the same width issue: `pc+1` at the final word cannot be stored faithfully. This report demonstrates sequential escape; the CALL-at-last-word case is a required follow-up regression rather than a separately executed probe.

**Proposed design:** Preserve the model's existing boundary policy by representing one-past-ROM PC/return values and checking fetch range before using a physical ROM address. Widen the relevant trace/fault PC fields consistently. Section 11 records the alternative of faulting before the last instruction commits; that would change the model contract and must be an explicit decision. **Acceptance:** final-word arithmetic/push/EMIT, CALL at the boundary, legal last-word HALT, and explicit invalid branches all have agreed traces.

### F4 — Medium: BRAM fault records can report a stale second-operand tag

**Locations:** `rtl/stack_core_bram.sv:176` operand selection, line 250 decode/read planning, and line 606 `do_fault`.

When tc=1 and dc>0, `a_w` selects `opnd_q`. That register is refreshed only for selected binary/SWAP paths. BAD_OPCODE, CALL errors, and other fault paths can report it without fetching the actual second logical value.

```text
PUSH_TRUE
PUSH_S15 2
PUSH_S15 3
DROP
WORD 0xF8000

expected fault suffix: BAD_OPCODE 1 0
BRAM fault suffix:     BAD_OPCODE 0 0
```

The logical second item is BOOL(true) in deep RAM. The final PC, depth, and fault code match, but fault metadata violates the established trace contract. Guarding the tag by depth avoids absent operands; it does not establish that a present operand has been read.

**Proposed fix:** Ensure the fault snapshot has a valid logical top-two view, either through a generalized operand-context read when needed or a dedicated fault-capture read state. Keep PC and stacks unchanged while collecting that context. Avoid changing fault-tag semantics merely to hide the mismatch. **Acceptance:** bad opcode, bad branch, type fault, and return faults with tc=1/dc>0 and a non-INT deep second value. Evidence: `stack_core_bram-deep-fault-tags.diff`.

### F5 — Medium: CALL/RET publish return depth one clock before commitment

**Locations:** `rtl/stack_core.sv:339`, line 355, and line 491; `rtl/stack_core_bram.sv:452`, line 470, and line 687.

Both cores assign `rdepth_d` during EXECUTE and load it unconditionally each clock. The return-address array write and PC transition wait for COMMIT, but the return-depth register does not. A `CALL sub; HALT; sub: RET` probe prints:

```text
CYCLE 5  PC 0 RDEPTH 1 TRACE 0
TRACE 0 0 2 CALL 0 COMMIT
CYCLE 6  PC 2 RDEPTH 1 TRACE 1
...
CYCLE 10 PC 2 RDEPTH 0 TRACE 0
TRACE 1 2 1 RET 0 COMMIT
```

The ordinary trace still matches the model, and current recursive programs succeed. The observed defect is a violation of the stated architectural mutation boundary: before CALL commits, return depth says an address is live although its array write has not yet occurred. That matters for debug observation, refinement assertions, and future interrupts or recovery logic.

**Proposed fix:** Add a staged return-depth candidate and apply it in COMMIT along with the return-address write and PC update. **Acceptance:** rdepth and all live return addresses change only on retirement/reset; failing CALL and RET preserve both; recursive directed tests continue to pass. Evidence: both `*-call-ret-timing.log` files.

### F6 — High: malformed assembly lines are silently treated as comments

**Location:** `tools/asm20.py:29`, `_parse_line()`.

A regex nonmatch returns `(None,None,None)`, the same result used for blank/comment lines. `PUSH_S15 1 2` is therefore discarded instead of rejected. The probe `PUSH_S15 1 2; HALT` returns only `[458752]`, the HALT word, and puts the malformed line into the listing as a comment.

An assembler must preserve the programmer's intended instruction sequence or clearly reject it. Silent omission changes subsequent label positions and can make a broken program look like a valid successful halt.

**Proposed fix:** Detect blank/comment lines explicitly; reject every other regex nonmatch with source name, line number, and original line. **Acceptance:** extra tokens, malformed labels, punctuation, and missing separators raise diagnostics, while legitimate comments and whitespace remain accepted. Evidence: `probe-results.json`, case `malformed-line`.

### F7 — Medium: raw WORD bypasses program capacity validation

**Location:** `tools/asm20.py:75`, WORD branch, compared with the later length check near line 103.

The raw-word branch appends a word and executes `continue` before the capacity check. `assemble('WORD 0\nWORD 0\nWORD 0\n', rom_depth=2)` accepts three words. The CLI can consequently write an oversized image and defer the mismatch to a simulator or synthesizer.

**Proposed fix:** Route all word insertion through one checked append operation or perform a common validation after every emitting branch. **Acceptance:** exact-capacity images succeed; an additional normal instruction, raw word, or mixed sequence fails consistently. Evidence: `probe-results.json`, case `word-capacity`.

### F8 — Medium: test observability and generation leave significant blind spots

**Locations:** `sim/test_bram.py`, `sim/tb_stack_core*.sv`, `sim/test_model.py`.

The count invariant is weaker than value refinement, and normal commits omit values. The random generator has the weaknesses in section 9.2. Noncanonical Boolean testing constructs only Python state; it does not establish the RTL branch. ADD/SUB boundary tests likewise rely on constructed model states, while the program-level overflow fixture is for MUL. The board's 514-slot capacity is synthesized but not exercised at its limit by the current simulation configurations.

**Proposed fix:** Add a testbench-only architectural snapshot and compare complete values and return addresses after every retirement. Track random-program state with executable semantics, record actual opcode/branch/fault coverage, and add deterministic boundary tests. This is a verification design gap supported by the demonstrated escaped bugs, not a claim that every uncovered path is broken.

### F9 — Low: documentation and metadata drift complicate onboarding

The README retains obsolete per-suite counts; Makefile assembly comments promise `.bin`; metadata describes a package generator that is absent; state tuples omit the return stack; `uart_tx.sv` comments refer to an `io_block`/`io_req` not present here; CCF comments describe button inversion not used by current top. The full review does not need a new abstraction layer to resolve these issues: align comments with the code and generate or mechanically verify the duplicated numeric/name tables.

The five existing playbooks are useful starting points, but their statements about retirement and validation should be read alongside F1–F8. Documentation quality cannot substitute for enforcing the contracts it describes.

## 11. Proposed design decisions

These are recommendations for a subsequent implementation ticket, not approved changes to the machine's public behavior.

### Decision DR-1: preserve precise retirement across all architectural fields

- **Context:** Operand-stack staging mostly follows a commit boundary, but return depth publishes early and neutral operations can reuse stale candidates.
- **Options considered:** Patch isolated fields; introduce a complete staged retirement record; merge both RTL cores behind a common implementation.
- **Decision:** Use a complete per-instruction candidate in each existing core, with explicit ownership of PC, both depths, writes, cache updates, and halt state.
- **Rationale:** This directly addresses F2 and F5 while retaining the value of distinct physical implementations.
- **Consequences:** More explicit staging fields and invariants; no compatibility adapter or architectural merge is required.
- **Status:** proposed.

```text
EXECUTE:
    candidate = preserve(current architectural state)
    failure = validate(opcode, operands, capacities, target)
    if failure:
        collect valid fault context; publish fault
    else if opcode == EMIT:
        pending_output = logical_top
        enter OUTPUT_WAIT
    else:
        fill complete candidate, including both stack depths
        enter COMMIT

COMMIT:
    atomically install candidate and all enabled writes
    publish exactly one COMMIT trace

OUTPUT_WAIT:
    offer stable pending_output
    if ready:
        pop data stack using already captured refill
        advance PC; publish exactly one OUTPUT trace
```

Implementation should use the established Yosys/Icarus-compatible RTL subset. A conceptual retirement record does not require a large packed struct if that makes the existing compiler flow harder to inspect. Explicit fields are acceptable as long as completeness and ownership are enforced.

### Decision DR-2: represent the model's one-past-ROM address

- **Context:** A physical ROM address width cannot encode the model's escaped PC or a last-word return address.
- **Options considered:** Define wraparound; fault before the last instruction retires; widen architectural PC and fault on the following fetch.
- **Decision:** Prefer widening architectural PC to represent ROM_DEPTH and retain the current model's subsequent-fetch fault.
- **Rationale:** Preserves existing abstract behavior and makes the mismatch explicit instead of silently changing the specification.
- **Consequences:** PC-related trace and return-stack widths must change together; the physical ROM interface remains a separate address-width concern. Fetch-fault event naming must be defined in RTL because its current op_name table has no FETCH opcode.
- **Status:** proposed; settle the exact FETCH trace representation before implementation.

A useful width distinction is `ADDR_W=max(1,clog2(ROM_DEPTH))` versus `PC_W=max(1,clog2(ROM_DEPTH+1))`. Validate `pc<ROM_DEPTH` before issuing a meaningful fetch. Parameter assumptions, including positive depths and legal upper limits for 15-bit targets, should be asserted during elaboration or simulation. Changing only the arithmetic expression without changing stored widths does not solve truncation.

### Decision DR-3: strengthen observation in the testbench first

- **Context:** Current traces are readable but do not encode enough state to prove representation equivalence.
- **Options considered:** Expand the production UART protocol; add production debug RAM immediately; expose full snapshots only in simulation.
- **Decision:** Start with simulation-only full-state snapshots and retain the concise human-readable trace for diagnosis.
- **Rationale:** Finds value/occupancy/return-address errors without consuming board BRAM budget or changing the UART contract.
- **Consequences:** Tests depend on representation-aware extraction helpers. Those helpers must reconstruct the logical stack accurately and be reviewed independently.
- **Status:** proposed.

```text
snapshot_register_core:
    data = stack_q[0:depth_q]
    returns = rstack_q[0:rdepth_q]

snapshot_bram_core:
    data = deep.mem[0:dc_q]
    if tc_q == 2: append top1_q
    if tc_q >= 1: append top0_q
    returns = rstack_q[0:rdepth_q]

after retirement:
    assert snapshot.data == model.stack as complete 40-bit words
    assert snapshot.returns == model.rstack
    assert pc, halted, fault metadata, output transfers agree

without retirement or reset:
    assert abstract snapshot has not changed
```

This intentionally compares only live storage. Unused RAM/register contents can remain stale. For fault precision, snapshot before the attempted instruction and compare it after FAULT. For output, compare accepted out_data including flags, not merely the truncated UART representation.

### Decision DR-4: reject invalid source at the assembler boundary

- **Context:** The parser conflates malformed source with comments, and multiple append paths enforce different limits.
- **Options considered:** Warn and continue; reject invalid syntax; introduce a new parser dependency.
- **Decision:** Keep the small parser but make invalid syntax an error and centralize word-capacity validation.
- **Rationale:** The grammar is small enough for the current approach; the defects are boundary handling, not lack of a parser framework.
- **Consequences:** Previously silently ignored bad source fails clearly. No backwards-compatibility mode is proposed.
- **Status:** proposed.

```text
parse_line(raw, source_name, line_number):
    if blank_or_comment(raw): return empty
    match = grammar.fullmatch(raw)
    if match is absent: raise diagnostic(source_name, line_number, raw)
    return parsed_fields(match)

append_word(word):
    if len(words) >= rom_depth: raise capacity_error
    words.append(word)
```

## 12. Phased implementation plan and acceptance criteria

### Phase 1: lock down demonstrated defects

Add normal regression tests corresponding to ticket probe cases in `sim/test_assembler.py`, `sim/test_directed.py`, and `sim/test_bram.py`. Preserve minimized input programs inline or as named program fixtures. The tests should fail for the reviewed implementation for the documented reason. Do not modify the model to match an RTL defect just to turn a diff green.

Fix the missing DUP precondition, complete JMP/SWAP depth staging, and assembler validation. These are bounded changes with clear acceptance criteria. Run the affected tests and then `make test`. Commit these behavior fixes separately from broader trace or PC-interface changes so reviewers can inspect each cause and effect.

### Phase 2: make architectural commitment complete

Add staged return depth in both RTL cores and apply it together with return writes in COMMIT. Introduce the cycle assertion that live return state changes only with retirement/reset. Extend data-state assertions at the same time, especially across EMIT stalls and fault transitions. Keep speculative memory reads permitted while forbidding publication of their results before commitment.

For F4, choose and document a fault-context acquisition path. Test faults with one cached value and a nonempty deep stack, ensuring the second operand's tag differs from stale/default tags. Acceptance includes unchanged stack contents, correct tags, and exactly one FAULT event.

### Phase 3: settle and implement address-boundary semantics

Resolve DR-2's FETCH trace representation first. Update model, both cores, trace ports, testbench variables, and name/format handling coherently. Maintain separate physical memory addresses and architectural PCs. Add last-word cases for HALT, PUSH, EMIT, CALL/RET, and a jump to the final valid address.

Also test a small non-power-of-two ROM if that parameterization is intended to be supported. If only a fixed board configuration is intended, state and validate that constraint explicitly rather than leaving unsupported parameter combinations looking public. No compatibility shims are part of this plan.

### Phase 4: improve verification quality and metadata consistency

Implement full-state snapshots and generator coverage counters. Make LT update the generator's tracked type correctly; check capacity before appending; separate deliberate invalid generation from successful terminating generation; avoid accidental dependence on ROM padding except in tests dedicated to padding. Add CALL/RET generation with bounded call depth and controlled termination.

Choose either generation or validation of the RTL opcode/fault/name mirrors from `opcodes.py`. A small generator/check command under `symbolic_eval/scripts/` can emit constants and fail CI if checked-in output differs. Semantics should remain independently implemented in model and RTL; generating all logic from one implementation would reduce oracle independence.

Acceptance is not merely a larger test count. Record observed opcode counts, fault counts, taken/non-taken branches, tc/dc transition coverage, stalls at meaningful boundaries, and configured capacity limits. Exercise total depth 514 and return depth 16 explicitly.

### Phase 5: document and re-establish hardware evidence

Update README counts and commands, comments, playbooks, and fault/trace documentation around the final implementation. Run synthesis, placement/routing, and bitstream packing with versions and program identity recorded. Associate UART captures and resource/timing logs with the exact commit and image hash.

A subsequent hardware validation should run the arithmetic, type-fault, and Fibonacci programs and one burst-output program. Start serial capture before loading or restart with the button while capture is active, as described in PB-04. Precise physical fault proof still needs richer instrumentation than silence and a blinking LED. A bounded trace RAM is a reasonable later extension, but it consumes memory beyond the existing two-block laboratory budget unless resources are deliberately reassigned.

### Validation matrix

| Contract | Concrete cases | Required observation |
|---|---|---|
| Data capacity | Empty consumers, full DUP/push, post-pop DUP | Correct fault and unchanged complete stack |
| Neutral operations | EMIT then JMP/SWAP/CALL/RET/HALT | Depth and live values match at each retirement |
| Arithmetic | Signed boundaries, negative SUB/LT, overflow | Correct tagged result or precise fault |
| BRAM representation | tc 0/1/2, spill, single/two reads, deep fault | Count and value refinement each cycle/retirement |
| Return stack | Empty RET, full CALL, nested calls, failing target | Addresses/depth preserved or committed together |
| PC policy | Last-word HALT, fallthrough, EMIT, CALL/RET | Model/RTL agreement including fetch faults |
| Output ownership | Long stalls, simultaneous buffer consume/replace | Stable offered data and one transfer per retirement |
| Reset | Button during compute, stall, and UART transmission | Defined abort, no stale replay, clean restart |
| Assembler | Invalid syntax, labels, mixed WORD capacity | Clear diagnostics and no omitted instructions |
| Board implementation | Production depth, actual program bitstream | Synthesis/routing budget and captured output |

## 13. Practical onboarding and reproduction

### Day one: execute the abstract machine

Read `tools/opcodes.py`, assemble the arithmetic program, and follow its stack transitions manually. Then inspect `Machine.step()` and compare it with the expected trace in `sim/test_model.py`. Write down what is preserved by TYPE_FAULT and why EQ between differently tagged values returns false rather than faulting.

Run commands from `symbolic_eval/` because several tests open `programs/...` relative to the working directory:

```bash
source ~/fpga/oss-cad-suite/environment
make test
make asm PROG=arith
python3 -m pytest sim/test_model.py sim/test_assembler.py -q
```

Review `build/arith.hex`, `.lst`, and `.sym.json`. Distinguish an instruction address from a byte address and observe that the hex file has ROM padding while `assemble()` returns only source words.

### Day two: follow one instruction through RTL

Start with `stack_core.sv` before reading the BRAM core. Identify current state, staged state, EXECUTE checks, the COMMIT array-write gates, and the OUTPUT_WAIT acceptance edge. Follow PUSH, ADD, and EMIT through the waveform or trace. Then inspect the probe showing stale ndepth: understanding why it fails is a direct test of understanding the `_q`/`_d` discipline.

### Day three: reconstruct the BRAM stack

For each instruction in `programs/deep.asm`, draw RAM contents, tc, dc, top0, and top1. Track logical values after a pop leaves tc=1. Work through the two-read binary case and explain why decrementing deep occupancy during the first read would break fault precision. Compare the count-only assertion in `tb_stack_core_bram.sv` with the stronger snapshot pseudocode in section 11.

### Reproduce this review's probes

From the repository root, run the ticket script:

```bash
bash ttmp/2026/09/04/GATEMATE-SYMBOLIC-002*/scripts/02-baseline.sh
```

It records versions, baseline tests, collection, and probe results. Individual generated benches and all investigation scripts live under ticket `scripts/`; logs, diffs, and hex inputs live under `reference/validation/`. The probe script exits successfully when it successfully runs its investigation. Inspect `match` in `probe-results.json`: success of the script does not mean the baseline machine is correct. Its generated benches derive from existing tests and add observation only.

The defuddle source-collection script is separate so simulation does not require a web connection. It archives two primary pages in `sources/`. The original laboratory book is already archived in the previous ticket; source provenance is recorded in this ticket's `sources/README.md`.

### Build without loading the board

```bash
source ~/fpga/oss-cad-suite/environment
make synth PROG=fib
make bit PROG=fib
```

`synth` assembles the selected program, copies its image to `build/prog.hex`, and runs the fixed `.ys` script. `pnr` then uses the CCF pin constraints and the 100 ns clock constraint; `bit` packs the result. The build is program specific: rebuilding with another PROG changes the embedded instruction ROM. `make sim` tests the original blink top; it is not the complete evaluator simulation. Use `make test` for that.

`make load PROG=fib` also rebuilds through the Makefile dependencies. Omitting PROG uses the default arithmetic program, even if the previous build was Fibonacci. Read PB-04 before a hardware session, particularly its serial-capture ordering. No board programming was performed during this review.

## 14. API and file reference index

### Python APIs

| File and anchor | Interface | Notes |
|---|---|---|
| tools/opcodes.py:81 | Instruction metadata | Opcode, mnemonic, immediate kind, stack requirements/effects |
| tools/opcodes.py:155 | encode(opcode, imm=0) -> int | Packs fields; caller supplies semantic immediate validation |
| tools/opcodes.py:167 | decode(word) -> (opcode, imm15) | Immediate remains unsigned bits until interpreted |
| tools/opcodes.py:172 | sign_extend_s15(imm15) | Converts signed immediate bits to Python integer |
| tools/asm20.py:49 | assemble(source, rom_depth=1024) | Words, symbols, listing; no library padding |
| tools/asm20.py:137 | main(argv=None) | Writes .hex, .lst, .sym.json |
| tools/stack_model.py:57 | Value40(tag,payload,flags=0) | word(), from_word(), text() |
| tools/stack_model.py:85 | make_int(x), make_bool(b) | Construct values with zero flags |
| tools/stack_model.py:107 | TraceRecord.line() | Canonical human-readable event |
| tools/stack_model.py:138 | Machine | step(), run(), final_line(); directly constructible state |
| tools/stack_model.py:357 | load_hex(path) | Read hex instruction list |
| tools/stack_model.py:362 | run_program(...) -> Machine | ROM/data/return depths and max_steps configuration |

### RTL interfaces

| Module and anchor | Parameters | Contract |
|---|---|---|
| rtl/stack_core.sv:18 | ROM_DEPTH, STACK_DEPTH, RSTACK_DEPTH | Synchronous ROM, ready/valid value output, trace pulse, persistent fault |
| rtl/stack_core_bram.sv:24 | ROM_DEPTH, DEEP_DEPTH, RSTACK_DEPTH | Same semantic role; total data capacity DEEP_DEPTH+2; tc/dc debug ports |
| rtl/sync_sdp_ram.sv:9 | DEPTH, WIDTH | One synchronous read plus optional write each clock |
| rtl/program_rom.sv:10 | DEPTH | Registered 20-bit read; PROG_HEX for synthesis initialization |
| rtl/rv_reg.sv:10 | WIDTH | One-entry elastic register; synchronous active-high reset |
| rtl/uart_tx.sv:7 | CLK_HZ, BAUD | start/data accepted while ready; 8N1 serial output |
| rtl/reset_sync.sv:6 | none | Async reset assertion at arst_n, two-stage release |
| rtl/top.sv:22 | ROM_DEPTH, DEEP_DEPTH | 10 MHz clock, button, active-low LED, UART pins |
| rtl/top.sv:161 | byte_of(v,idx) | Tag/payload formatter; flags omitted |
| rtl/blink_top.sv:9 | LED_BIT | P0 clock/reset/LED bring-up design |

Trace fields are meaningful when trace_valid is asserted; no trace_ready exists. Fault fields are meaningful when fault_valid is asserted. out_data is meaningful when out_valid is asserted and must remain stable under backpressure. Reset and terminal state behavior are as described above; standalone wrapper signal polarity is not uniform across all modules.

### Build, tests, and operational references

- `symbolic_eval/Makefile`: dependency order and program selection; `symbolic_eval/scripts/synth.ys`: actual synthesis input list and fixed top parameters.
- `symbolic_eval/constraints/top.sdc:1`: 100 ns clock requirement; `olimex_gatematea1_evb.ccf`: board pin names.
- `symbolic_eval/sim/test_directed.py:21`: directed program inventory and capacity/stall configurations.
- `symbolic_eval/sim/test_bram.py:99`: random program generator; `tb_stack_core_bram.sv`: count invariant and output/PC monitors.
- `symbolic_eval/sim/test_top.py`: serial-output assertions; `tb_top.sv`: UART sampling and post-halt drain.
- `symbolic_eval/playbooks/PB-01-adding-an-opcode.md`: extension checklist; PB-02: trace mismatch investigation; PB-03: fault codes; PB-04: board session; PB-05: RTL portability.
- Prior ticket `GATEMATE-SYMBOLIC-001`, `reference/01-investigation-diary.md`, Steps 10 and 12: historical synthesis and board observations; Step 14: playbook history.
- Prior ticket `sources/Composable_Hardware_Patterns_for_Symbolic_Computers.md`, Laboratory 1 around lines 4211–4569: original local specification and pattern context.

### Review artifacts

- `scripts/01-review-probes.py`: six core programs across both implementations, two assembler probes, and additional return-depth timing observation.
- `scripts/02-baseline.sh`: environment, baseline, inventory, probe runner.
- `scripts/03-collect-sources.sh`: defuddle primary-source archive.
- `reference/validation/probe-results.json`: structured outcomes; `.diff` files: exact model deviations; `.log` files: raw simulation and build evidence.
- `reference/01-investigation-diary.md`: chronological work, failures, commits, and upload evidence.

## 15. Review boundaries and remaining decisions

The evidence supports a working demonstration with a strong model-first foundation and several specific correctness gaps. Fixing these gaps should precede larger features such as boxed integers, heap references, graph reduction, or replayable arithmetic. The proposed work is bounded: enforce existing instruction preconditions, make staged retirement complete, define address escape consistently, and compare full architectural state.

The main open decision is PC/fetch-fault semantics at the ROM boundary, including its trace representation. Secondary decisions concern whether parameter combinations beyond the demonstrated configurations are supported, whether nonzero flags remain valid for INT/BOOL operations, and how much board fault instrumentation belongs inside the laboratory budget. None requires a compatibility adapter by default.

The original hardware reports remain useful evidence of specific executions. A fresh routed build and board capture should be attached to the later correctness implementation, tied to its commit and image, rather than presenting old successful demonstrations as proof that the reviewed machine handles every program correctly.

# symbolic_eval — Laboratory 1: a precise tagged stack evaluator

A 40-bit tagged-value stack machine for the Olimex GateMateA1-EVB
(Cologne Chip CCGM1A1), built with open-source tools only. This is
Laboratory 1 of *Composable Hardware Patterns for Symbolic Computers*
(the book lives in the ticket `sources/`; the intern guide is the ticket
design doc).

## Exit criteria (book Laboratory 1) — VERIFIED ON HARDWARE

```text
Program A:  ((7+5)*3)==36 -> EMIT BOOL(true), HALT
            board UART: "T1:00000001\r\n", then LED solid (halted)

Program B:  BOOL(true) + INT(4) -> TYPE_FAULT at the ADD
            pc and stack unchanged, no output transfer, LED fast-blink
```

## Machine summary

- **value40**: tag[3:0], flags[3:0], payload[31:0] (INT, BOOL, REF, ...).
- **ISA**: 17 opcodes (PUSH_S15, PUSH_TRUE/FALSE, ADD, SUB, MUL, EQ, LT,
  DUP, DROP, SWAP, JMP, JZ, EMIT, HALT, CALL, RET) in a 20-bit word
  `[19:15] opcode, [14:0] imm`.
- **CALL/RET extension** (book Lab 1 extension): a separate 16-entry
  return-address stack (registers). CALL pushes pc+1 and jumps; RET pops
  and jumps. Empty RET / full-call-stack fault precisely.
- **Precise faults**: STACK_UNDERFLOW, STACK_OVERFLOW, TYPE_FAULT,
  ARITH_OVERFLOW, BAD_OPCODE, BAD_BRANCH_TARGET, NONCANONICAL_BOOL,
  RSTACK_UNDERFLOW, RSTACK_OVERFLOW — each leaves pc and stacks exactly as
  before the failing instruction.
- **Commitment**: no architectural mutation before all checks pass; COMMIT
  is the single mutation owner; EMIT pops only on output-channel acceptance.
- **Two implementations of the same contract** (differentially tested
  against one model):
  - `rtl/stack_core.sv` — register stack (32-entry).
  - `rtl/stack_core_bram.sv` — BRAM deep stack + two-entry top cache
    (Split-Lifetime Frame; invariant `depth == tc + dc` checked every
    cycle in the testbench).

## Method (stolen from MATE-16)

`tools/opcodes.py` is the single source of truth for the ISA.
`tools/stack_model.py` is the executable reference model and differential
oracle. `tools/asm20.py` is the two-pass assembler. The RTL testbenches
print TRACE/FINAL lines in exactly the model's format; the pytest suites
diff them line by line, with random output backpressure and random legal /
illegal program generation.

## Layout

```
constraints/   CCF pins, SDC timing (verified on this board in MATE-16)
rtl/           symbolic_types_pkg, stack_core (reg), stack_core_bram + sync_sdp_ram,
               program_rom, rv_reg, uart_tx, reset_sync, top
tools/         opcodes.py, stack_model.py, asm20.py
programs/      arith, typefault (exit criteria), smoke, deep, branch,
               muloverflow, underflow, badop, badbranch, jztype,
               stackoverflow, fib (recursive CALL/RET), sq, countdown,
               retunderflow, calloverflow
sim/           tb_stack_core, tb_stack_core_bram, tb_top + pytest suites
scripts/       synth.ys
build/         generated (gitignored)
```

## Toolchain

```bash
source ~/fpga/oss-cad-suite/environment
make versions    # record tool versions
make test        # 105 tests: model + assembler + differential RTL + board sim
make asm PROG=arith
make bit PROG=arith   # synth -> PnR -> pack (program image baked into ROM)
make load             # load to the board, then read the UART:
stty -F /dev/ttyACM0 115200 raw -echo && cat /dev/ttyACM0
```

Pressing the FPGA button restarts the program (global experiment abort,
design doc DR-4).

## Output format

Each EMIT value becomes 13 UART bytes at 115200 8N1:

```
'T' <tag hex digit> ':' <8 payload hex digits, MSB first> CR LF
```

e.g. `INT(5)` -> `T0:00000005`, `BOOL(true)` -> `T1:00000001`.

## Measured results (stop-build budget: 2 BRAM blocks / ~2000 CPEs)

| Metric | Value |
|---|---|
| Block RAM | 2 x CC_BRAM_20K (1Kx20 ROM + 512x40 stack) |
| CPEs (packed) | 392 / 20480 |
| Max frequency (routed, with CALL/RET) | 16.56 MHz (PASS at 10 MHz board clock) |
| Multipliers | 1 (CC_MULT, 34x34 MUL) |

## LED codes (the EVB user LED is active-LOW: pin low = LED lit)

- running: slow blink (~0.6 Hz)
- halted (success): solid ON — e.g. fib leaves it lit after ~2 ms
- faulted: fast blink (~2.4 Hz)

## Test inventory (123 green)

- `test_model.py` (30): book Program A trace exact, Program B precise
  fault, every opcode at minimum depth, all fault cases (constructed
  states for NONCANONICAL_BOOL / ARITH_OVERFLOW), value packing.
- `test_assembler.py` (12): round-trip, labels, error cases.
- `test_directed.py` (12): register core vs model on all programs +
  random output stalls.
- `test_bram.py` (47): BRAM core on all programs + refinement invariant +
  24 random-program differential tests + 12 register-core random tests.
- `test_top.py` (6): full-board sim, UART byte stream vs model EMITs
  (incl. fib and countdown).
- `test_model.py` CALL/RET: return addresses, nested calls, precise
  RSTACK_UNDERFLOW / RSTACK_OVERFLOW, recursive fib(10)=55.

## Portability notes (yosys ∩ iverilog SystemVerilog subset)

- No `return` in functions (classic function-name assignment).
- No `import` (file-scope or module-header): all package references are
  fully qualified `symbolic_types_pkg::NAME`.
- No multi-declarator typedef'd struct variables.
- ROM init via a `.ys` script file (yosys tokenizer handles the quoted
  filename; `-D` on the command line does not).
- `value40_t.tag` is a plain vector: enum-typed struct fields break yosys
  width inference inside package functions.

## Status

- P0 bootstrap — done (blink on board)
- P1 model — done (30 tests)
- P2 assembler — done (12 tests)
- P3 register RTL — done (differential, book trace exact)
- P4 BRAM + top cache — done (invariant + random tests)
- P5 board — done (both exit criteria on hardware)

Extensions still open (book): boxed integers, capability REF descriptors,
replayable multi-cycle MUL, trace RAM / ILA capture of fault records.
CALL/RET is DONE (recursive fib verified on hardware: `T0:00000037`).

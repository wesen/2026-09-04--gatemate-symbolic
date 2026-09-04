# symbolic_eval — precise tagged stack evaluator

A 40-bit tagged-value stack machine for the Olimex GateMateA1-EVB, implemented
with Python and SystemVerilog using the open-source CAD flow. It implements
Laboratory 1 of the book archived in ticket GATEMATE-SYMBOLIC-001.

The implementation review is GATEMATE-SYMBOLIC-002. Correctness repairs,
regressions, build records, and printed work slips are in GATEMATE-SYMBOLIC-003.

## Machine contract

- Values are `tag[39:36]`, `flags[35:32]`, `payload[31:0]`. INT arithmetic uses
  signed 32-bit payloads and precise overflow faults; constructors clear flags.
- Seventeen 20-bit instructions: PUSH_S15, PUSH_TRUE, PUSH_FALSE, ADD, SUB, MUL,
  EQ, LT, DUP, DROP, SWAP, JMP, JZ, EMIT, HALT, CALL, RET. Opcode occupies bits
  19:15; the remaining 15 bits encode a signed literal or unsigned target.
- EQ compares all 40 bits. JZ requires BOOL payload zero or one. It validates
  its target even when the branch will not be taken.
- The separate return-address stack defaults to 16 entries. CALL saves pc+1;
  RET restores it. Return occupancy and address writes retire together.
- Normal instructions publish architectural changes in COMMIT; EMIT publishes
  on ready/valid acceptance. Preparation cycles and faults preserve PC and both
  logical stacks. Output holds valid/data stable while blocked.
- Faults: STACK_UNDERFLOW, STACK_OVERFLOW, TYPE_FAULT, ARITH_OVERFLOW,
  BAD_OPCODE, BAD_BRANCH_TARGET, NONCANONICAL_BOOL, RSTACK_UNDERFLOW,
  RSTACK_OVERFLOW. A fault records PC, opcode context, data depth, and the true
  top-two logical operand tags (zero when absent), then stops until reset.

Architectural PC and return addresses can represent ROM_DEPTH. An instruction
at the final valid ROM word may retire to that address; the next fetch faults
BAD_BRANCH_TARGET with the trace mnemonic FETCH. `trace_fetch` and `fault_fetch`
distinguish this context from an instruction fault; opcode fields are not
meaningful when these flags are set. Physical ROM addresses remain narrower and
are guarded. ROM sizes 2..32768 and data/return memory depths >=2 are supported
by the core parameter checks.

The register core defaults to 32 operand slots. The board uses the BRAM core
with 512 deep slots plus two cached values: total capacity **514**. The cache can
hold only one value even with nonempty deep RAM. `depth=tc+dc` and complete value
ordering are checked in simulation. Internal synchronous reads acquire operands,
refills, and accurate fault context before publication.

## Layout

```text
tools/          opcodes.py, stack_model.py, asm20.py
rtl/            two cores, tagged types, RAM/ROM, reset, rv_reg, UART, top
programs/       arithmetic, faults, deep stack, branches, fib, sq, countdown
sim/            model/assembler/core/board tests and full-state monitors
scripts/        synth.ys and check_isa.py
constraints/    board pins and 10 MHz timing constraint
playbooks/      PB-01 opcode, PB-02 debug, PB-03 faults, PB-04 board, PB-05 RTL
build/          ignored generated images, simulation files, logs, coverage
```

`tools/opcodes.py` is the metadata authority. The assembler and model import it;
`scripts/check_isa.py` verifies handwritten RTL constants and bench name tables.
The assembler rejects malformed lines and oversized images, including raw WORD
instructions. It returns actual words; the CLI pads `.hex` with zero words
(PUSH_S15 0), and also writes `.lst` and `.sym.json`. It does not write `.bin`.

## Build and test

```bash
source ~/fpga/oss-cad-suite/environment
cd symbolic_eval
make versions
make test                         # 197 tests
python3 scripts/check_isa.py
make asm PROG=fib
make bit PROG=fib                  # synth -> route -> pack
make load PROG=fib                 # rebuilds: keep PROG explicit
```

Run tests from `symbolic_eval/`; some fixtures open relative program paths.
`make sim` is only the original blink bring-up test. `make test` exercises the
evaluator and board simulation. PNR_SEED selects a reproducible router seed;
seed 2 is the repair build default. The source list and top parameters used by
synthesis are in `scripts/synth.ys`.

## Verification

The complete suite contains 197 collected tests:

| File | Count | Scope |
|---|---:|---|
| test_model.py | 36 | Semantic machine and constructed Python states |
| test_assembler.py | 12 | Encoding, labels, diagnostics |
| test_directed.py | 17 | Register-core directed programs and exit criteria |
| test_bram.py | 52 | BRAM directed and both-core generated execution |
| test_repairs.py | 52 | Review regressions, fault context, ROM boundaries |
| test_verification.py | 20 | Metadata, legal coverage, 514-slot capacity, RTL initial states |
| test_top.py | 8 | UART streams, halted LED, restart during compute/UART |

Both core runners compare TRACE/FINAL plus all live data/return values in STATE
records and complete accepted words in XFER records. Per-cycle assertions require
architectural state to stay unchanged between retirements and on faults. Stale
unused memory is intentionally ignored. Initial-stack injection exists only in
simulation to reach noncanonical Boolean and overflow states.

The generator validates candidate instruction fragments with the model, including
bounded nested calls and explicit branch joins. Legal runs terminate deliberately;
injected fault runs really fault. The coverage test exercises all 17 opcodes and
both JZ outcomes. `build/verification-coverage.json` records executed opcode, fault,
branch, and BRAM cache-transition counts. Passing tests are not a formal proof.

## Output and board behavior

Each accepted EMIT becomes 13 bytes at nominal 115200 8N1:

```text
T<tag hex digit>:<8 payload hex digits> CR LF
INT(55) -> T0:00000037
BOOL(1) -> T1:00000001
```

Flags are omitted from the UART text. EMIT retires when the elastic register
accepts the value, before serial transmission finishes. HALT does not imply the
UART is drained. Reset is a global experiment abort and may discard buffered or
partially transmitted output. The button is synchronized before reset assertion;
release restarts from the initial image. UART RX is unused.

The user LED is active-low: halted is solid on, running is slow blink, faulted is
fast blink. The top does not expose the full fault record on UART. Empty UART output
alone does not prove a precise fault occurred. See PB-04 for arming serial capture
before loading, so short programs do not finish before the host opens the port.

## Evidence and limits

The original implementation demonstrated arithmetic, type-fault behavior, and
fib(10)=55 on hardware. The repair ticket preserves fresh simulation/build and
board-capture results separately from those historical claims. The laboratory
budget remains two physical BRAM blocks and roughly 2000 packed CPEs. Consult
GATEMATE-SYMBOLIC-003 validation logs for final routed timing and image hashes.

Future work includes boxed integers, capability REF descriptors, replayable
multi-cycle MUL, and trace RAM. The declared REF/PAIR/THUNK tags do not yet provide
heap management or a graph-reduction runtime. Extra trace memory must be accounted
for separately from the two-BRAM laboratory budget.

## RTL portability

Keep synthesizable code within the tested Yosys/Icarus subset: fully qualified
package names, classic function-name assignment, one typedef-struct declaration
per line, plain vector struct tags, and ROM macro initialization in the `.ys`
script. Preserve synchronous RAM latency and explicit architectural retirement.

# PB-01: Adding or changing an opcode

The ISA has **one source of truth**: `tools/opcodes.py`. Everything else mirrors it.
Work through this checklist in order; the test suite at the end must be green before
you commit.

## Checklist

1. **`tools/opcodes.py`** — add the `Instruction(...)` entry (opcode, mnemonic,
   `imm_kind`, `stack_in`/`stack_out`, effect, description). The assembler picks it
   up automatically. Keep opcodes contiguous; `OPCODE_ILLEGAL_MIN` moves up.
2. **`tools/stack_model.py`** — add the `step()` branch. All checks **before** any
   mutation (capacity -> tags -> canonicality -> branch target). Add a
   `test_model.py` case: minimum-depth behavior, each fault case, and (if the
   opcode is reachable in odd states) a constructed-state test.
3. **`rtl/symbolic_types_pkg.sv`** — add the `OP_*` localparam (keep the value in
   sync with the table; `scripts/check_isa.py` verifies these handwritten mirrors).
4. **Both cores** — `rtl/stack_core.sv` **and** `rtl/stack_core_bram.sv`:
   - `EXECUTE`: the case branch (use `symbolic_types_pkg::`-qualified labels!),
     stage complete next-state data (`npc_d`, `ndepth_d`, `nrdepth_d`, write descriptors).
   - `COMMIT`: apply the staged representation change (stack write, cache update).
   - Remember the single-mutation-owner rule: architectural PC and stacks only change in `COMMIT`
     (and at `EMIT` acceptance in `S_OUTPUT_WAIT`).
5. **Both testbenches** — `sim/tb_stack_core.sv` **and**
   `sim/tb_stack_core_bram.sv`: extend `op_name()` with the mnemonic. A missed
   entry makes correct execution print `BAD` and fails the trace diff.
6. **Program + directed test** — write `programs/<name>.asm`, add it to
   `PROGRAMS` in `sim/test_directed.py` (pick `stack_depth` and stall seeds).
   `sim/test_bram.py` derives its list from it automatically.
7. **Board (optional)** — add the hex to the `test_top.py` expectations if the
   program should run on hardware.

## Verify

```bash
cd symbolic_eval && source ~/fpga/oss-cad-suite/environment
make asm PROG=<name> && python3 -m pytest sim/ -q     # all green
```

## Known traps

- Unqualified case labels (`OP_FOO:` instead of
  `symbolic_types_pkg::OP_FOO:`) elaborate as implicit wires: the machine faults
  `BAD_OPCODE` at your new instruction. See PB-05.
- `stack_in`/`stack_out` in the table must match the RTL staging exactly — the
  model's capacity checks are derived from them.
- If the opcode changes the trace line shape, that is a **frozen-format change**:
  update the model's `TraceRecord.line()`, both testbench `$display`s, and every
  hardcoded expectation in the tests in the same commit.

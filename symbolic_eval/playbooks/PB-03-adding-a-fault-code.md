# PB-03: Adding a fault code

Fault codes are a closed set shared by the model, the package, both cores, and
both testbenches. The field is currently **4 bits** — adding code 16 or higher requires
widening (see the checklist note at the end).

## Checklist

1. **`tools/opcodes.py`** — extend the `Fault` `IntEnum`. Keep `NONE = 0`.
2. **`rtl/symbolic_types_pkg.sv`** — add the `F_*` localparam (`[3:0]`).
3. **Both cores** — call `do_fault(F_...)` from `EXECUTE` at the right point in
   the **check order** (capacity -> tags -> canonicality -> branch target).
   `do_fault` stages the record and fires the `FAULT` trace pulse in the same
   cycle; the machine stops in `S_FAULT` with pc and stacks untouched.
4. **Both testbenches** — extend `fault_name()` (both
   `sim/tb_stack_core.sv` and `sim/tb_stack_core_bram.sv`); a missing entry
   prints `?` and fails the diff.
5. **`tools/stack_model.py`** — the `_fault(...)` call with the same ordering;
   fault records carry the top-two operand tags (0 when absent).
6. **Tests** — model test for the fault (constructed state is fine when the case
   is unreachable from legal programs); a `programs/*.asm` demo if reachable.

## Semantics to preserve

- **Precise:** the faulting instruction leaves pc, data stack, and return stack
  exactly as they were. If your fault can fire after a prefetch (BRAM core),
  the prefetched value is simply discarded — reads are side-effect-free.
- **First fault wins, machine stops.** No recovery in bytecode; a debug
  `CLEAR_FAULT` would be outside the ISA.
- **Check order is contract.** A program failing two checks must produce the
  identical record in model and RTL.

## Width note

Fault ports, registers, `do_fault`'s argument, and `trace_fault` are `[3:0]`.
Codes 0x0–0x9 are used. If you need more than six new codes, widen to `[4:0]`
everywhere in one commit (grep `\[3:0\].*fault` in `rtl/`, `sim/`), and expect
`FINAL`/`FAULT` line changes to touch every hardcoded test expectation.


Fetch faults are published from FETCH (or the BRAM fetch-context states), with
`trace_fetch`/`fault_fetch` asserted. Keep fault PC wide enough for ROM_DEPTH.
For tc=1 and dc>0 the second operand tag comes from a synchronous deep-RAM read;
never use stale opnd_q without acquiring context. Constructed-state RTL cases in
`test_verification.py` cover noncanonical Booleans and arithmetic extremes.

# PB-05: RTL portability rules (yosys ∩ iverilog)

The code must compile under **iverilog** (simulation) and **yosys**
(synthesis). Write in the intersection only. When a construct below tempts
you, use the listed replacement — every rule here was learned from a real
elaboration failure.

## Rules

| Do not use | Use instead |
|---|---|
| `return` in functions | function-name assignment: `mk_int = v;` |
| `import pkg::*;` (anywhere) | fully-qualified names: `symbolic_types_pkg::TAG_INT` |
| `value40_t a, b;` | one declaration per line |
| `parameter string` / `-DPROG_HEX='"..."'` on the command line | `scripts/synth.ys` reads the ROM file (yosys's own tokenizer) |
| enum-typed struct fields (in package functions) | plain `logic [3:0]` field |
| string/enum ternaries | `if`/`else` into a local variable |
| combinational reads of RAM | request/response-valid sync wrapper (`rtl/sync_sdp_ram.sv`) |

Notes:

- `` `ifdef PROG_HEX `` guards the ROM's `$readmemh` so simulation loads ROM
  images hierarchically from the testbench instead.
- iverilog's `sorry: constant selects in always_* processes...` messages are
  informational, not errors. Do not "fix" working code because of them.
- Keep ports flat (`logic [39:0]`); packed structs live inside module bodies.

## The patch rule (process, not syntax)

Every scripted edit to RTL must **assert its match count and be verified by
grep** before the work counts as applied. This project had three patches that
silently no-oped (rewritten comments, whitespace drift) while printing
success. The symptom is diagnostic: `BAD_OPCODE` at exactly the new
instruction means an unqualified case label or a case insert that never
applied — the differential suite catches it in one run, but the fix is to not
ship unverified edits.

## Symptom → cause quick table

| Symptom | Cause |
|---|---|
| `BAD_OPCODE` at the new opcode | unqualified case label / insert never applied |
| trace fields print `x` | snapshot registers declared but never assigned in `always_ff` |
| `unexpected TOK_IMPORT` | leftover `import` statement |
| `This assignment requires an explicit cast` | enum-typed ternary |
| works in sim, wrong in hardware | combinational array read standing in for sync RAM |
| correct run, wrong mnemonic in traces | `op_name()` not extended (both testbenches) |

## Vault background

The generalizable version of this playbook (with the full failure narratives)
lives in the parc vault: `Research/2026/09/04/ARTICLE - Playbook - The yosys
and iverilog SystemVerilog Subset.md`.

# PB-02: Debugging a trace mismatch

The differential tests print a unified diff (`model` vs `rtl`). Read it as a
story: the first differing line is the divergence point; everything after it is
usually fallout.

## Procedure

1. **Recompile first.** Stale `.vvp` files produce phantom failures:
   ```bash
   source ~/fpga/oss-cad-suite/environment
   iverilog -g2012 -s tb_stack_core -o build/tb_reg_d32.vvp \
     rtl/symbolic_types_pkg.sv rtl/stack_core.sv sim/tb_stack_core.sv \
     -Ptb_stack_core.STACK_DEPTH=32
   ```
   (The pytest suite recompiles per config, but manual `vvp` runs do not.)

2. **Look at the first differing line only.** Useful patterns:
   | Diff shape | Likely cause |
   |---|---|
   | RTL stops after N lines, `FINAL ... BAD_OPCODE` | case label not qualified, or the case insert never applied (PB-05) |
   | `BAD` as the op name, correct pc flow | testbench `op_name()` missing the mnemonic |
   | One `depth` wrong from line k on | staging bug: wrong `ndepth_d`/`sdc_d` delta, or a missing cache shift |
   | FAULT line differs only in tags | fault-record operand tags (guarded by depth?) |
   | RTL trace is a prefix of the model's | watchdog timeout or a hang: dump the FSM state (`--vcd` + `gtkwave build/*.vcd`) |
   | Output values differ, trace identical | `EMIT` pending register or printer path, not the core |

3. **Decide which side is right first.** The model is the specification. If the
   RTL matches the book/design doc and the model does not, fix the model — but
   decide explicitly, then update the other side and the docs in the same commit.

4. **Reproduce minimally.** Shrink the failing program to the shortest sequence
   that still diverges, keep it as a directed test, then fix.

## The invariant checks are part of the diff

`tb_stack_core_bram` asserts `depth_o == tc_o + dc_o` every cycle and
`ASSERT_FAIL: ...` lines fail the test even when traces match. An invariant
failure localizes the first bad edge — look at what retired just before it
(the last `TRACE` line printed before the `ASSERT_FAIL`).

## Useful one-liners

```bash
# run one program on both cores and diff against the model
make asm PROG=fib
vvp build/tb_bram_d30.vvp +rom=build/fib.hex +stall_seed=7
# view waves of a hang
iverilog -g2012 -s tb_stack_core -o build/t.vvp rtl/*.sv sim/tb_stack_core.sv
```

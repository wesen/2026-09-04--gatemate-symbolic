# symbolic_eval — Laboratory 1: a precise tagged stack evaluator

A 40-bit tagged-value stack machine for the Olimex GateMateA1-EVB
(Cologne Chip CCGM1A1), built with open-source tools only. This is
Laboratory 1 of *Composable Hardware Patterns for Symbolic Computers*
(book in the ticket `sources/`); the intern guide is the ticket design doc.

## Exit criteria

```text
Program A:  ((7+5)*3)==36 -> EMIT BOOL(true), HALT
Program B:  BOOL(true) + INT(4) -> TYPE_FAULT, pc/stack unchanged, no output
```

## Layout

```
constraints/   CCF pins, SDC timing (verified on this board in MATE-16)
rtl/           synthesizable SystemVerilog
tools/         opcodes.py, stack_model.py, asm20.py
programs/      example .asm programs
sim/           testbenches + pytest suites
scripts/       yosys scripts
build/         generated (gitignored)
```

## Toolchain

```bash
source ~/fpga/oss-cad-suite/environment
make versions   # record tool versions
make test       # software + differential RTL tests
make sim        # P0 blink simulation
make bit        # synth -> PnR -> pack
make load       # load to the board
```

## Status

- P0: bootstrap (blink top) — done

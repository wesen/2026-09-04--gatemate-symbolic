# Changelog

## 2026-09-04

- Initial workspace created


## 2026-09-04

Step 1: ticket bootstrap + survey of book and prior GateMate projects

### Related Files

- /home/manuel/code/wesen/2026-09-04--gatemate-symbolic/ttmp/2026/09/04/GATEMATE-SYMBOLIC-001--symbolic-computer-patterns-and-designs-for-gatemate-analysis-design-and-intern-implementation-guide/sources/Composable_Hardware_Patterns_for_Symbolic_Computers.md — imported source


## 2026-09-04

Step 3: intern onboarding guide (design doc) written, ~31KB, references book + prior projects

### Related Files

- /home/manuel/code/wesen/2026-09-04--gatemate-symbolic/ttmp/2026/09/04/GATEMATE-SYMBOLIC-001--symbolic-computer-patterns-and-designs-for-gatemate-analysis-design-and-intern-implementation-guide/design-doc/01-intern-onboarding-guide-symbolic-computer-patterns-on-gatemate.md — primary deliverable


## 2026-09-04

Step 4: doctor passed (vocab added), bundle uploaded to reMarkable /ai/2026/09/04/GATEMATE-SYMBOLIC-001


## 2026-09-04

Step 5: ticket complete — all phases done, final handoff reported


## 2026-09-04

Impl P0: symbolic_eval bootstrap, blink verified in sim + on board

### Related Files

- /home/manuel/code/wesen/2026-09-04--gatemate-symbolic/symbolic_eval/Makefile — P0 flow


## 2026-09-04

Impl P1: opcodes.py + stack_model.py + 30 model tests green (book Program A trace exact)

### Related Files

- /home/manuel/code/wesen/2026-09-04--gatemate-symbolic/symbolic_eval/tools/stack_model.py — reference model/oracle


## 2026-09-04

Impl P2: asm20.py two-pass assembler + 11 programs, 42 tests green

### Related Files

- /home/manuel/code/wesen/2026-09-04--gatemate-symbolic/symbolic_eval/tools/asm20.py — assembler


## 2026-09-04

Impl P3: register-stack RTL + differential tests, 54 green, book trace exact on RTL

### Related Files

- /home/manuel/code/wesen/2026-09-04--gatemate-symbolic/symbolic_eval/rtl/stack_core.sv — register-stack core


## 2026-09-04

Impl P4: BRAM top-cache core + random/boundary differential tests, 101 green

### Related Files

- /home/manuel/code/wesen/2026-09-04--gatemate-symbolic/symbolic_eval/rtl/stack_core_bram.sv — BRAM core with top cache


## 2026-09-04

Impl P5: board top, 2 BRAM / 392 CPE / 17 MHz, both exit criteria verified on hardware (T1:00000001; typefault silent)

### Related Files

- /home/manuel/code/wesen/2026-09-04--gatemate-symbolic/symbolic_eval/rtl/top.sv — board integration


## 2026-09-04

Impl P6: README finalized, 105 tests green, wrap-up

### Related Files

- /home/manuel/code/wesen/2026-09-04--gatemate-symbolic/symbolic_eval/README.md — final documentation


## 2026-09-04

Extension: LED polarity fixed (active-low), CALL/RET + return stack + fib/sq/countdown programs, 123 tests, fib(10)=55 on hardware

### Related Files

- /home/manuel/code/wesen/2026-09-04--gatemate-symbolic/symbolic_eval/rtl/stack_core_bram.sv — CALL/RET extension


## 2026-09-04

Playbooks: 5 vault Research articles (linked from report, pushed) + 5 repo-local pragmatic playbooks (PB-01..05) with README section

### Related Files

- /home/manuel/code/wesen/2026-09-04--gatemate-symbolic/symbolic_eval/playbooks/PB-01-adding-an-opcode.md — repo-local playbook set


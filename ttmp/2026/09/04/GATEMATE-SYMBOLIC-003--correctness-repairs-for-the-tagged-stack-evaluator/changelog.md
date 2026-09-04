# Changelog

## 2026-09-04

- Initial workspace created


## 2026-09-04

P1 complete: guards and assembler validation fixed; 145 tests pass (a3b50fb). Pre-fix counterexamples and print evidence preserved.


## 2026-09-04

P2 complete: return state retires atomically and deep fault tags are valid; 157 tests pass (3213364).


## 2026-09-04

P3 complete: ROM escape and CALL return addresses preserved with explicit fetch-fault flags; 175 tests pass.


## 2026-09-04

P3 committed c9805a6; P4 committed 662843a. Full-state and reset verification passes 197 tests. P5 seed 2 routing completes in 72 iterations at 15.52 MHz; hardware captures in progress.


## 2026-09-04

P5 ongoing: first three board streams match, countdown final tag differs persistently. Production-depth RTL simulation is correct; mapped-netlist diagnostic running. No repair yet attempted.


## 2026-09-04

P5 synthesis investigation: board countdown mismatch reproduced in mapped netlist, isolated to constant constructor tag loss, repaired with packed assignments in 605f41d; 198 tests pass. Initial failing captures preserved.

### Related Files

- /home/manuel/code/wesen/2026-09-04--gatemate-symbolic/symbolic_eval/rtl/symbolic_types_pkg.sv — Packed constructors preserve tags during constant folding


## 2026-09-04

Implementation complete and ready for review: 198 tests, all nine faults and both branches exercised, 16.03 MHz Fibonacci route at 10 MHz target, four matching board streams. Final constructor repair is 605f41d; full diary and original failures retained.


## 2026-09-04

Published an 8,073-word CPU architecture article to go-go-parc in pushed vault commit 0b45d4a. Added reproducible model examples and article validation; preserved the historical vault report. Diary Step 10 records the source-based research and delivery.


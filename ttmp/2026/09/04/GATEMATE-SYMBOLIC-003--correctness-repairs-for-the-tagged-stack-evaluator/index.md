---
Title: Correctness repairs for the tagged stack evaluator
Ticket: GATEMATE-SYMBOLIC-003
Status: review
Topics:
    - fpga
    - gatemate
    - symbolic-computers
    - architecture
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: repo://symbolic_eval/rtl/stack_core.sv
      Note: Register implementation repairs
    - Path: repo://symbolic_eval/rtl/stack_core_bram.sv
      Note: BRAM implementation repairs
    - Path: repo://symbolic_eval/rtl/symbolic_types_pkg.sv
      Note: Synthesis-safe constructors
ExternalSources: []
Summary: Repair review findings F1–F9 and a hardware-discovered constructor synthesis defect, with full-state regressions and board evidence.
LastUpdated: 2026-09-04T16:31:31.837411014-04:00
WhatFor: Review the implemented evaluator correctness repairs and their validation evidence.
WhenToUse: Reviewing the repairs or extending the assembler, cores, and verification harness.
---



# Correctness repairs for the tagged stack evaluator

## Overview

This ticket implements the findings from GATEMATE-SYMBOLIC-002. It repairs stack
guards, candidate-state initialization, return-stack retirement, deep operand
fault context, ROM escape semantics, and assembler diagnostics. Verification now
compares complete live architectural state and checks deliberately generated
legal and faulting programs. The final hardware pass also exposed and repaired
constant constructor tag loss during synthesis.

## Key Links

- [Repair design and implementation review](design-doc/01-correctness-repair-design-and-validation-plan.md)
- [Detailed implementation diary](reference/01-implementation-diary.md)
- [Final test run: 198 passing tests](reference/validation/P5-final-tests.log)
- [Board captures and bitstream hashes](reference/validation/P5-hardware.json)
- [Original failing board evidence](reference/validation/before-constructor-fix/P5-hardware.json)
- [Printed plan and phase receipts](reference/slips/)

## Status

Current status: **review**. All five implementation phases are complete.

## Topics

- fpga
- gatemate
- symbolic-computers
- architecture

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design-doc/ - Repair plan, implementation map, and synthesis diagnosis
- reference/ - Chronological diary, validation logs, raw captures, and printed slips
- scripts/ - Reproducible diagnostics, phase printing, and delivery validation
- sources/ - Reserved for downloaded references; this implementation used local sources

## Review order

Read the design first, then the diary's pre-fix failures and post-fix outcomes.
Inspect commits `a3b50fb`, `3213364`, `c9805a6`, `662843a`, and `605f41d` for
behavior and verification changes. Documentation and build-seed changes are in
`49807a4`. Use `scripts/13-validate-delivery.py` to check archived final evidence
without reprinting slips or programming the board.

UART silence for the type-fault program establishes absence of output only;
precise fault state is checked in simulation. The tests and hardware samples do
not constitute formal equivalence verification.

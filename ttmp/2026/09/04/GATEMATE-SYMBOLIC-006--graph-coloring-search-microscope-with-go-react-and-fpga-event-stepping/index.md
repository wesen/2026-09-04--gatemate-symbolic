---
Title: Graph coloring search microscope with Go React and FPGA event stepping
Ticket: GATEMATE-SYMBOLIC-006
Status: active
Topics:
    - fpga
    - gatemate
    - symbolic-computers
    - architecture
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: repo://queens_rollback/rtl/queens_core.sv
      Note: Verified source for mutation and rollback sequencing
    - Path: repo://symbolic_eval/rtl/sync_sdp_ram.sv
      Note: Shared synchronous memory contract
    - Path: repo://symbolic_eval/rtl/uart_tx.sv
      Note: Existing board transmitter
ExternalSources: []
Summary: ""
LastUpdated: 2026-09-04T18:26:11.612790646-04:00
WhatFor: ""
WhenToUse: ""
---


# Graph coloring search microscope with Go React and FPGA event stepping

## Overview

Implemented a runtime graph-coloring search microscope for the physical GateMate FPGA with Go and TypeScript/React. The board accepts graphs up to eight vertices and eight colors and publishes lossless semantic events. The browser edits graphs and inspects domains, checkpoints, mutation trails, accepted colorings, and 256 retained snapshots.

The implementation and final checks are complete. The live embedded service is available at <http://127.0.0.1:8086>, using the physical serial engine in tmux `graph-final`. Logs go to `/tmp/graph-api.log`. The board is paused after event 2 of a three-color triangle, ready for interactive stepping.

## Documents and evidence

- [Intern analysis, design, and implementation guide](design-doc/01-graph-coloring-search-microscope-intern-analysis-design-and-implementation-guide.md): principles, source map, pseudocode, diagrams, wire and HTTP contracts, execution walkthrough, measurements, and validation limits.
- [Detailed implementation diary](reference/01-implementation-diary.md): chronological changes, failures, corrections, review instructions, and delivery records.
- [Tasks](tasks.md) and [changelog](changelog.md).
- [Desktop screenshot](reference/validation/P6-desktop.png) and [mobile historical view](reference/validation/P6-mobile-history.png).
- [Physical API acceptance](reference/validation/api-serial-smoke.json), [Go race checks](reference/validation/P6-go-tests.log), [queens regression](reference/validation/P6-queens-tests.log), and [final vulnerability scan](reference/validation/P6-vulnerabilities.log).
- [Initial design upload](reference/validation/design-upload.log) and [completed guide upload](reference/validation/final-upload.log).
- `reference/slips/`: overall plan plus each phase start/completion receipt.
- `scripts/`: reproducible investigation, validation, documentation, and printing helpers.
- `sources/`: consulted primary documentation extracted using Defuddle.

The original guide was delivered before implementation. The final guide uses the separate reMarkable name **GATEMATE 006 Implemented Microscope Guide** under `/ai/2026/09/04/GATEMATE-SYMBOLIC-006`, preserving the earlier document and any annotations.

## Acceptance results

Physical runs returned six triangle colorings (86 events), zero for the two-color triangle (26), two path colorings (32), and one first-only coloring (12). Go race checks with UART simulations, ten frontend tests, both Go build modes, TypeScript/Vite, vet/Glazed lint, and all 58 queens regressions passed. The Go 1.26.8 scan reported zero reachable vulnerabilities. Routed FPGA timing passed the 10 MHz constraint at 26.76 MHz.

Historical viewing is bounded and does not reverse the FPGA. One local service owns one device. Larger graphs, arbitrary constraint tables, persistent sessions, and remote multi-user control are outside this laboratory's implemented scope.

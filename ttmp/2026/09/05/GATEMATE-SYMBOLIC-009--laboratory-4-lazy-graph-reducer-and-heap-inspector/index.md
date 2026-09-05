---
Title: Laboratory 4 lazy graph reducer and heap inspector
Ticket: GATEMATE-SYMBOLIC-009
Status: complete
Topics:
    - fpga
    - gatemate
    - architecture
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: repo://symbolic_eval/rtl/sync_sdp_ram.sv
      Note: Synchronous heap and stack memory building block
ExternalSources: []
Summary: ""
LastUpdated: 2026-09-05T17:59:58.261342027-04:00
WhatFor: ""
WhenToUse: ""
---



# Laboratory 4 lazy graph reducer and heap inspector

## Overview

Implement Laboratory 4 as an inspectable single-evaluator lazy graph machine: a 1024x40 heap, 512x80 continuation stack, checked signed arithmetic, memoized values/errors, UART control and a Go/React heap inspector.

P1–P5 are complete. Model, RTL, host and browser validation passed. The GateMate was programmed successfully after reconnection; 60 randomized physical graphs, five directed examples, bounds tests and physical browser checks passed. Final routed timing is 24.65 MHz at the required 10 MHz. The physical inspector is available at http://127.0.0.1:18090/ while its tmux server and board remain connected.

## Key Links

- [Intern analysis, design and implementation guide](design-doc/01-lazy-graph-reducer-intern-analysis-design-and-implementation-guide.md)
- [Implemented API and illustrated qualification handoff](reference/02-implemented-reducer-api-and-qualification-handoff.md)
- [Detailed diary](reference/01-implementation-diary.md)
- [Archived lab chapter](sources/laboratory-4.md)
- Model inspector: http://127.0.0.1:18089/
- reMarkable: `/ai/2026/09/05/GATEMATE-SYMBOLIC-009`

## Status

Current status: **active**

## Topics

- fpga
- gatemate
- architecture

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts

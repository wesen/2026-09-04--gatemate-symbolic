---
Title: Laboratory 4 lazy graph reducer and heap inspector
Ticket: GATEMATE-SYMBOLIC-009
Status: active
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
LastUpdated: 2026-09-05T17:05:15.016837927-04:00
WhatFor: ""
WhenToUse: ""
---


# Laboratory 4 lazy graph reducer and heap inspector

## Overview

Implement Laboratory 4 as an inspectable single-evaluator lazy graph machine: a 1024x40 heap, 512x80 continuation stack, checked signed arithmetic, memoized values/errors, UART control and a Go/React heap inspector.

P1–P4 are complete. Go model/reference, RTL/protocol and browser checks passed; final routed timing is 24.65 MHz at the required 10 MHz. P5 physical qualification is waiting for the disconnected GateMate board. The initial JTAG programming attempt could not open the device; no physical pass is claimed.

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

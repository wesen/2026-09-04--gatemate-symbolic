---
Title: Programmable dataflow workbench compiler and physical debugger
Ticket: GATEMATE-SYMBOLIC-008
Status: complete
Topics:
    - fpga
    - gatemate
    - architecture
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2026-09-04T23:34:15.949726282-04:00
WhatFor: ""
WhenToUse: ""
---


# Programmable dataflow workbench compiler and physical debugger

## Overview

Build a programmable typed expression workbench on the existing elastic dataflow engine: validated graph loading, compiler and fanout lowering, hardware breakpoints, bounded trace, and a Go/React interface. All implementation phases and physical qualification passed. Final routed timing is 12.01 MHz against the 10 MHz constraint; 224 randomized physical expression results passed. The illustrated handoff includes seven model and FPGA screenshots.

## Key Links

- [Intern analysis, design, and implementation guide](design-doc/01-programmable-dataflow-workbench-intern-analysis-design-and-implementation-guide.md)
- [Implemented API and qualification reference](reference/02-implemented-programmable-workbench-api-and-physical-qualification-reference.md)
- [Detailed implementation diary](reference/01-implementation-diary.md)
- reMarkable design delivery: `/ai/2026/09/04/GATEMATE-SYMBOLIC-008`
- Physical workbench while the board/server are running: http://127.0.0.1:8087/

## Status

Current status: **complete**

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

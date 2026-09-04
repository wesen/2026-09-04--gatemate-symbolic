---
Title: Lab 2 rollback solver analysis design and intern implementation guide
Ticket: GATEMATE-SYMBOLIC-005
Status: active
Topics:
    - fpga
    - gatemate
    - symbolic-computers
    - architecture
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2026-09-04T17:02:16.920581184-04:00
WhatFor: ""
WhenToUse: ""
---

# Lab 2 rollback solver analysis design and intern implementation guide

## Overview

Design investigation for the Laboratory 2 project in GATEMATE-SYMBOLIC-004. The saved software experiment validates all 92 ordered solutions and agreement between snapshot and trail restoration. Diagram sources and rendered figures are available.

The complete 7,261-word guide has now been recovered and validated after the user resumed work. It explains the existing infrastructure, proposed solver APIs, snapshot-to-trail refactoring, memory schedules, precise faults, result acceptance, and implementation/testing phases. Production Lab 2 implementation remains under GATEMATE-SYMBOLIC-004.

See [the complete design and intern guide](design-doc/01-lab-2-snapshot-to-trail-refactoring-and-intern-implementation-guide.md). Delivery records are kept under reference/validation.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field
- [Experiment results](reference/validation/design-experiment.json)
- [Investigation diary](reference/01-investigation-diary.md)

## Status

Current status: **active**

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

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts

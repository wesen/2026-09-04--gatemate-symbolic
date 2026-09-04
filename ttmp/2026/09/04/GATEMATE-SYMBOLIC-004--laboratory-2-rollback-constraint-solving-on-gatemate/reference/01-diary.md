---
Title: Diary
Ticket: GATEMATE-SYMBOLIC-004
Status: active
Topics:
    - fpga
    - gatemate
    - symbolic-computers
    - architecture
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Chronological investigation and implementation diary for Laboratory 2.
LastUpdated: 2026-09-04T16:57:27.323426618-04:00
WhatFor: ""
WhenToUse: ""
---

# Diary

## Goal

Establish and implement the book's rollback constraint solver while preserving source evidence, decisions, experiments, and validation results.

## Step 1: Reread the lab and establish its scope

I reread the complete Laboratory 2 chapter and its supporting checkpoint/trail chapter before creating this ticket. The project is a fixed eight-queens search engine with explicit reversible domain updates, deterministic alternatives, and accepted result commitment. This step records scope only; no solver RTL or model was implemented.

### Prompt Context

**User prompt (verbatim):**

> Ok, new ticket for the next lab project. Read the lab book again to refresh your memory, and tell me what th eproject is in your own words. (2 paragraphs or so)

**Assistant interpretation:** Identify the next worked lab from the archived book, open its ticket, and briefly explain the project and learning objective.

**Inferred user intent:** Begin the next lab with a refreshed, source-grounded understanding before implementation.

### What I did

- Located the book in GATEMATE-SYMBOLIC-001 sources and inspected the lab sequence.
- Read Laboratory 2, including its experimental sequence and extensions, and the checkpoint/trail substrate.
- Created GATEMATE-SYMBOLIC-004 with a project brief, acceptance criteria, planned tasks, and this diary.

### Why

The next lab introduces semantic rollback across search decisions. Its requirements must come from the book rather than assumptions about extending the previous CPU.

### What worked

The book states exact results: 92 boards, first board `[0,4,7,5,2,6,1,3]`, and FIRST_ONLY cut after accepted output. It also prescribes a full-snapshot baseline before the trail implementation, enabling a measured comparison.

### What didn't work

No failures. The ticket identifier was checked and did not already exist.

### What I learned

The propagated bitmap is restart state, not disposable scratch. The choice stack and mutation trail serve different purposes, and result acceptance is separate from reversible search mutation.

### What was tricky to build

No implementation in this step. The design must preserve log-before-write ordering, publish a complete choice before the branch mutation, restore domains in reverse order, and defer cut until the first solution is accepted.

### What warrants a second pair of eyes

Check the planned model's propagation order and zero-domain handling against the book. Do not turn optional N-queens generalization, heuristics, or multiple contexts into baseline requirements.

### What should be done in the future

Write the design, then implement the independent reference and full-state snapshot baseline before introducing the trail.

### Code review instructions

Read this ticket index alongside book lines 4569–4961 and 3724–3884. Run docmgr doctor for GATEMATE-SYMBOLIC-004. There are no new runtime changes to test.

### Technical details

Domains are eight eight-bit masks; trail and choice records map to separate RAMs. Solutions pack eight three-bit rows into 24 bits. The book asks to simplify around two RAM blocks and 3,000 CPEs. All new scripts are stored in this ticket's scripts folder; the existing source book is related rather than duplicated.
